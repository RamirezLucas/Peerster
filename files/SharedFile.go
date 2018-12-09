package files

import (
	"Peerster/data"
	"Peerster/fail"
	"Peerster/frontend"
	"Peerster/messages"
	"os"
	"sync"
)

// FileStatus represents the status of a file (acts as an enumeration).
type FileStatus int

const (
	// UncompleteMatch is used when some of the file's chunk have not been localized on the network.
	UncompleteMatch FileStatus = 0
	// CompleteMatch is used when some all of the file's chunks have been localized on the network
	// but the user has not requested the file yet.
	CompleteMatch FileStatus = 1
	// NoMetafileMultiSource is used when the metafile of a file that must be retreived from multiple
	// sources is unknown.
	NoMetafileMultiSource FileStatus = 2
	// NoMetafileMonoSource is the initial state of a shared file that is retrieved from a single source.
	NoMetafileMonoSource FileStatus = 3
	// MissingChunks is used when the metafile is present but some chunks are still unknown.
	MissingChunks FileStatus = 4
	// Reconstructed is used when all chunks have been downloaded (i.e. the file is complete).
	Reconstructed FileStatus = 5
)

const (
	// ChunkSizeBytes is the size of a chunk in bytes.
	ChunkSizeBytes = 8192
	// HashSizeBytes is the size of a hash in bytes.
	HashSizeBytes = 32
	// MaxNbChunks is the maximum number of chunks allowed.
	MaxNbChunks = ChunkSizeBytes / HashSizeBytes
	// MaxFileSizeBytes is the maximum filesize allowed in bytes.
	MaxFileSizeBytes = MaxNbChunks * ChunkSizeBytes
)

/*SharedFile represents a file for the gossiper. The object contains information about which
chunks are present locally and which are only present on remote peers, as well as the metafile
and the metahash. The file maitains its current status in the `Status` field. */
type SharedFile struct {
	Filename string              // The filename
	Metahash [HashSizeBytes]byte // 32-bytes SHA-256 hash of metafile
	Metafile []byte              // Metafile in RAM

	ChunkBitmap      *data.Bitmap      // Bitmap indicating which chunks are present/missing
	DownloadedChunks []uint64          // List of downloaded chunks index
	RemoteChunks     map[uint64]string // Map from chunk indices to peers possessing these chunks

	Status            FileStatus // The file status
	ChunkCount        uint64     // Number of chunks for this file
	IsDownloaded      bool       // Indicates whether the file was indexed here first or dowloaded
	IsMonosource      bool       // Indicates whether the file is downloaded from a single source
	MetafileQueryPeer string     // The first peer to reply to a SearchRequest for a multisourced file

	mux sync.Mutex // Mutex to manipulate the structure from different threads
}

/*NewSharedFileLocal creates a new instance of SharedFile for a file located on the local machine.
In particular, the memory is already allocated for all fields since the number of chunks is known at creation.
The metahash should be set by the caller after the object is returned.

`filename` The file to index (must be stored in the PathToSharedFiles folder).

`chunkCount` The number of chunks for this file.

The function returns a pointer to the created `SharedFile`. */
func NewSharedFileLocal(filename string, chunkCount uint64) *SharedFile {
	var shared SharedFile

	shared.Filename = filename
	shared.Metafile = make([]byte, chunkCount*HashSizeBytes)

	shared.ChunkBitmap = data.NewBitmap(chunkCount)
	shared.DownloadedChunks = make([]uint64, chunkCount)
	for i := uint64(0); i < chunkCount; i++ {
		shared.ChunkBitmap.SetBit(i)
		shared.DownloadedChunks[i] = i
	}
	shared.RemoteChunks = nil

	shared.Status = Reconstructed
	shared.ChunkCount = chunkCount
	shared.IsDownloaded = false
	shared.IsMonosource = false
	shared.MetafileQueryPeer = ""

	return &shared
}

/*NewSharedFileMonoSource creates a new instance of SharedFile for a file located on the network and
fetched from a single source. In particular, the filename is already known while the number of chunks is unknown.

`filename` The name of the file to index.

`metahash` The file's metahash.

The function returns a pointer to the created `SharedFile`. */
func NewSharedFileMonoSource(filename string, metahash []byte) *SharedFile {
	var shared SharedFile

	shared.Filename = filename
	copy(shared.Metahash[:], metahash[:])
	shared.RemoteChunks = nil
	shared.DownloadedChunks = nil

	// Set status and number of chunks
	shared.Status = NoMetafileMonoSource
	shared.ChunkCount = 0

	shared.IsDownloaded = true
	shared.IsMonosource = true
	shared.MetafileQueryPeer = ""

	// These fields are allocated by SetMetafile when it's called on the SharedFile
	shared.Metafile = nil
	shared.ChunkBitmap = nil

	return &shared
}

/*NewSharedFileMultiSource creates a new instance of SharedFile for a file located on the network and
fetched from possible multiple sources. In particular, the filename and the number of chunks are known*/
func NewSharedFileMultiSource(filename string, chunkCount uint64, metahash []byte) *SharedFile {
	var shared SharedFile

	shared.Filename = filename
	copy(shared.Metahash[:], metahash[:])
	shared.RemoteChunks = make(map[uint64]string)
	shared.DownloadedChunks = nil

	// Set status and number of chunks
	shared.Status = UncompleteMatch
	shared.ChunkCount = chunkCount

	shared.IsDownloaded = true
	shared.IsMonosource = false
	shared.MetafileQueryPeer = ""

	// These fields are allocated by SetMetafile when it's called on the SharedFile
	shared.Metafile = nil
	shared.ChunkBitmap = nil

	return &shared
}

// SetMetafile sets the metafile of a SharedFile from the data contained in reply.Data.
// The functions returns a boolean indicating whether file reconstruction is complete.
func (shared *SharedFile) SetMetafile(reply *messages.DataReply) bool {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// Check arguments and file status
	if reply == nil {
		fail.CustomPanic("SharedFile.SetMetafile", "Invalid arguments (reply) = (%p).", reply)
	} else if shared.Status != NoMetafileMonoSource && shared.Status != NoMetafileMultiSource {
		fail.CustomPanic("SharedFile.SetMetafile", "Trying to set metafile of file with incorrect status %d.", shared.Status)
	}

	// Get number of chunks
	nbChunks := uint64(len(reply.Data) / HashSizeBytes)

	// If the file is multisource we already know the number of chunks
	// In that case silently change the number of chunks
	if shared.Status == NoMetafileMultiSource && shared.ChunkCount != nbChunks {
		fail.LeveledPrint(1, "SetMetafile", `Received metafile length yields %d chunks which
			doesn't match expected number %d for file with metahash %s`, ToHex32(shared.Metahash))
	}
	shared.ChunkCount = nbChunks

	// Set metafile
	shared.Metafile = make([]byte, len(reply.Data))
	copy(shared.Metafile[:], reply.Data)

	// Allocate bitmap
	shared.ChunkBitmap = data.NewBitmap(nbChunks)
	fail.LeveledPrint(1, "SharedFile.SetMetafile", "Created bitmap with length %d", nbChunks)

	// Mark SharedFile as having received a metafile
	if nbChunks == 0 {
		shared.Status = Reconstructed
		shared.AcknowledgeFileReconstructed()

		// Create an empty file
		f, err := os.OpenFile(PathToDownloadedFiles+shared.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			fail.LeveledPrint(1, "WriteChunk", `Failed to open file %s`, shared.Filename)
		}
		f.Close()

		return true
	}

	shared.Status = MissingChunks
	return false
}

// GetChunk returns one chunk of a shared file, or its metafile. If chunkID is
// 0 then the file's metafile is returned, otherwise a chunk of data is returned.
func (shared *SharedFile) GetChunk(chunkID uint64) []byte {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// Check arguments and file status
	if chunkID > shared.ChunkCount {
		fail.CustomPanic("SharedFile.GetChunk", "Invalid arguments (chunkID) = (%d).", chunkID)
	} else if shared.Status != MissingChunks && shared.Status != Reconstructed {
		fail.CustomPanic("SharedFile.GetChunk", "Trying to get chunk from file with incorrect status %d.", shared.Status)
	}

	// Return the metafile
	if chunkID == 0 {
		return shared.Metafile
	}

	// Return one of the file's chunk

	// Compute the filepath
	path := shared.Filename
	if shared.IsDownloaded {
		path = PathToDownloadedFiles + path
	} else {
		path = PathToSharedFiles + path
	}

	// Open the file
	var f *os.File
	var err error
	if f, err = os.Open(path); err != nil {
		fail.LeveledPrint(1, "GetChunk", `Failed to open file %s`, shared.Filename)
		return nil
	}
	defer f.Close()

	// Create a buffer for the chunk
	chunkBuffer := make([]byte, ChunkSizeBytes)
	nbBytesRead := 0

	// Read in the file
	if _, err = f.Seek(int64((chunkID-1)*ChunkSizeBytes), 0); err != nil {
		return nil
	}
	if nbBytesRead, err = f.Read(chunkBuffer); err != nil {
		return nil
	}

	// Return data
	return chunkBuffer[:nbBytesRead]
}

// WriteChunk writes one chunk of data at the file's end.
// ChunkID's start from 1 and extend up to shared.ChunkCount included.
// The functions returns a boolean indicating whether file reconstruction is complete.
func (shared *SharedFile) WriteChunk(chunkID uint64, data []byte) bool {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// Check arguments and file status
	if chunkID == 0 || chunkID > shared.ChunkCount {
		fail.CustomPanic("SharedFile.WriteChunk", "Invalid arguments (chunkID) = (%d).", chunkID)
	} else if shared.Status != MissingChunks {
		fail.CustomPanic("SharedFile.WriteChunk", "Trying to write chunk to file with incorrect status %d.", shared.Status)
	}

	// Open the file in write mode
	f, err := os.OpenFile(PathToDownloadedFiles+shared.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fail.LeveledPrint(1, "WriteChunk", `Failed to open file %s`, shared.Filename)
	}
	defer f.Close()

	// Write the chunk
	if nbBytesWrote, err := f.Write(data); err != nil || nbBytesWrote != len(data) {
		fail.CustomPanic("SharedFile.WriteChunk", "Failed to write file %s.", shared.Filename)
	}

	// Update the chunk map
	if oldVal := shared.ChunkBitmap.SetBit(chunkID - 1); oldVal {
		fail.CustomPanic("SharedFile.WriteChunk", "Incorrect bitmap at index %d for file %s.", chunkID-1, shared.Filename)
	}
	// Update remote chunk map
	if _, ok := shared.RemoteChunks[chunkID]; ok {
		delete(shared.RemoteChunks, chunkID)
	}
	// Update list of downloaded chunks and possible update status
	shared.DownloadedChunks = append(shared.DownloadedChunks, chunkID)
	if uint64(len(shared.DownloadedChunks)) == shared.ChunkCount { // The file has been completly reconstructed
		shared.Status = Reconstructed
		shared.RemoteChunks = make(map[uint64]string)
		shared.AcknowledgeFileReconstructed()
		return true
	}

	return false
}

// GetFileSearchInfo returns a pointer to a SearchResult containing information
// on the SharedFile (metahash and chunk map). The function returns nil if the
// file is not completly initialized.
func (shared *SharedFile) GetFileSearchInfo() *messages.SearchResult {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// If the file doesn't have a filename or a true chunk count return nil
	if shared.Status == NoMetafileMonoSource {
		return nil
	}

	// Create SearchResult
	result := &messages.SearchResult{
		Filename:     shared.Filename,
		MetafileHash: make([]byte, HashSizeBytes),
		ChunkMap:     make([]uint64, len(shared.DownloadedChunks)),
		ChunkCount:   shared.ChunkCount,
	}
	copy(result.MetafileHash[:], shared.Metahash[:])
	for i := 0; i < len(shared.DownloadedChunks); i++ {
		result.ChunkMap[i] = shared.DownloadedChunks[i] + 1
	}

	return result
}

/*UpdateChunkMappings @TODO*/
func (shared *SharedFile) UpdateChunkMappings(mappings []uint64, origin string) bool {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// Return if mappings is nil or if the file is an incorrect status
	if mappings == nil || shared.Status == Reconstructed {
		return false
	}

	for _, chunkID := range mappings { // For each remote chunk
		if chunkID != 0 && shared.Status == UncompleteMatch || shared.Status == CompleteMatch {
			// Update remote chunk location with most recently received SearchReply
			shared.RemoteChunks[chunkID] = origin
		}
	}

	// Check if we now have a complete match
	if shared.Status == UncompleteMatch && uint64(len(shared.RemoteChunks)) == shared.ChunkCount {
		shared.Status = CompleteMatch
		shared.MetafileQueryPeer = origin

		// Send update to frontend
		shared.AcknowledgeCompleteMatch()

		return true
	}

	return false
}

/*GetMetafileQueryPeer @TODO*/
func (shared *SharedFile) GetMetafileQueryPeer() string {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// The file must be in the CompleteMatch state
	if shared.Status != CompleteMatch {
		return ""
	}

	// The MetafileQueryPeer muist be defined
	if shared.MetafileQueryPeer == "" {
		fail.CustomPanic("SharedFile.GetRandomSharingPeer", "Undefined MetafileQueryPeer.")
	}

	shared.Status = NoMetafileMultiSource
	return shared.MetafileQueryPeer
}

/*GetChunkTarget @TODO*/
func (shared *SharedFile) GetChunkTarget(nextChunk uint64, lastOrigin string) string {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// Same origin in case of monosource file
	if shared.IsMonosource {
		return lastOrigin
	}

	// Look into the remote chunk mappings
	if peer, ok := shared.RemoteChunks[nextChunk]; ok {
		return peer
	}

	// Should never happen
	fail.CustomPanic("SharedFile.GetChunkTarget", "No known target for file %s chunk %d",
		shared.Filename, nextChunk)
	return ""
}

/*ChangeName @TODO*/
func (shared *SharedFile) ChangeName(newName string) {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	// Check file state
	if shared.Status != NoMetafileMultiSource {
		fail.CustomPanic("SharedFile.ChangeName",
			"Trying to change name of file with incorrect status %d.", shared.Status)
	}

	// Change the name
	shared.Filename = newName
}

/*SwitchMultiToMonoSource @TODO*/
func (shared *SharedFile) SwitchMultiToMonoSource(newName string) bool {
	// Grab the mutex
	shared.mux.Lock()
	defer shared.mux.Unlock()

	if shared.Status == CompleteMatch {
		shared.Status = NoMetafileMonoSource
		shared.Filename = newName
		return true
	}
	return false
}

// AcknowledgeFileReconstructed should be called when a file has been completely reconstructed.
func (shared *SharedFile) AcknowledgeFileReconstructed() {
	// Send update to frontend
	frontend.FBuffer.AddFrontendIndexedFile(shared.Filename, ToHex32(shared.Metahash))

	fail.LeveledPrint(1, "SharedFile.AcknowledgeFileReconstructed",
		"Reconstructed %s (%d chunks) with metahash %s", shared.Filename, shared.ChunkCount, ToHex32(shared.Metahash))
}

// AcknowledgeCompleteMatch should be called when a file has been matched completly during a SarchRequest.
func (shared *SharedFile) AcknowledgeCompleteMatch() {
	// Send update to frontend
	frontend.FBuffer.AddFrontendAvailableFile(shared.Filename, ToHex32(shared.Metahash))

	fail.LeveledPrint(1, "SharedFile.AcknowledgeCompleteMatch",
		"Available %s (%d chunks) with metahash %s", shared.Filename, shared.ChunkCount, ToHex32(shared.Metahash))
}
