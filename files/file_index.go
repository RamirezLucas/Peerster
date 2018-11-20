package files

import (
	"Peerster/frontend"
	"Peerster/messages"
	"crypto/sha256"
	"os"
	"strings"
	"sync"
)

const (
	// PathToSharedFiles - Path to folder where shared files are stored
	PathToSharedFiles = "_SharedFiles/"
	// PathToDownloadedFiles - Path to folder where downloaded files are stored
	PathToDownloadedFiles = "_Downloads/"
	// ChunkSizeBytes - Size of a chunk in bytes
	ChunkSizeBytes = 8192
	// HashSizeBytes - Size of a hash in bytes
	HashSizeBytes = 32
	// MaxNbChunks - Maximum number of chunks allowed
	MaxNbChunks = ChunkSizeBytes / HashSizeBytes
	// MaxFileSizeBytes - Maximum filesize allowed in bytes
	MaxFileSizeBytes = MaxNbChunks * ChunkSizeBytes
)

// FileIndex - Represents a file index
type FileIndex struct {
	index       map[string]*SharedFile // A mapping from filename to SharedFile structures
	knownHashes map[string]*KnownHash  // A mapping from a known hash to its corresponding file index
	mux         sync.Mutex             // Mutex to manipulate the structure from different threads
}

// NewFileIndex - Creates a new instance of FileIndex
func NewFileIndex() *FileIndex {
	var fileIndex FileIndex
	fileIndex.index = make(map[string]*SharedFile)
	fileIndex.knownHashes = make(map[string]*KnownHash)
	return &fileIndex
}

// AddNewSharedFile - Ads a new indexed file to the index
func (fileIndex *FileIndex) AddNewSharedFile(filename, origin string, metahash []byte) *SharedFile {
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	if _, ok := fileIndex.index[filename]; ok { // We already have a file indexed with this name
		return nil
	}

	// Send update to frontend
	frontend.FBuffer.AddFrontendConstructingFile(filename, ToHex(metahash[:]), origin)

	// Add the file
	newFile := NewSharedFile(filename, metahash, true)
	fileIndex.index[filename] = newFile
	return newFile
}

// IndexNewFile - Indexes a new file
func (fileIndex *FileIndex) IndexNewFile(filename string) {
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	// Append the newly indexed file to the file index
	if _, ok := fileIndex.index[filename]; ok { // We already have a file with the same name
		return
	}

	// Open the file
	var f *os.File
	var err error
	if f, err = os.Open(PathToSharedFiles + filename); err != nil {
		return
	}
	defer f.Close()

	// Check the filesize (must not be too large)
	fi, err := f.Stat()
	if err != nil || fi.Size() > MaxFileSizeBytes {
		return
	}

	// Create new SharedFile
	shared := NewSharedFile(filename, nil, false)

	// Compute total number of chunks
	nbChunks := GetChunksNumberFromRawFile(int(fi.Size()))

	// Buffer for chunk
	chunkBuffer := make([]byte, ChunkSizeBytes)
	nbBytesRead := 0

	// Create the metafile
	shared.Metafile = make([]byte, nbChunks*HashSizeBytes)
	for metafileIndex := uint64(0); metafileIndex < nbChunks; metafileIndex++ {

		// Read the next chunk
		if nbBytesRead, err = f.Read(chunkBuffer); err != nil {
			panic("IndexNewFile(): Indexed file could not be read correctly.")
		}

		// Create hash and append it to metafile
		chunkHash := sha256.Sum256(chunkBuffer[:nbBytesRead])
		if nbCopy := copy(shared.Metafile[metafileIndex*HashSizeBytes:], chunkHash[:]); nbCopy != HashSizeBytes {
			panic("IndexNewFile(): Metafile could not be generated.")
		}

		// Add the hash to the set of known hashes
		fileIndex.knownHashes[ToHex(chunkHash[:])] = NewKnownHash(shared, false, metafileIndex)
	}

	// Create metahash
	metahash := sha256.Sum256(shared.Metafile[:])
	if nbCopy := copy(shared.Metahash, metahash[:]); nbCopy != HashSizeBytes {
		panic("IndexNewFile(): Metafile could not be generated.")
	}

	// Add the metahash to the set of known hashes
	fileIndex.knownHashes[ToHex(metahash[:])] = NewKnownHash(shared, true, 0)

	// Add the new indexed file to the index
	fileIndex.index[filename] = shared

	// Send update to frontend
	frontend.FBuffer.AddFrontendIndexedFile(filename, ToHex(metahash[:]))
}

// GetMatchingFiles returns the list of SearchResult's corresponding to files whose
// filename contains at least one of the keyword contained in the keywords slice.
func (fileIndex *FileIndex) GetMatchingFiles(keywords []string) []*messages.SearchResult {
	// Grab the mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	results := make([]*messages.SearchResult, 0)

	// Iterate over all known files
	for filename, shared := range fileIndex.index {

		// Search for a keyword in the filename
		for _, k := range keywords {
			if strings.Contains(filename, k) {
				// We have a match
				results = append(results, shared.GetFileSearchInfo())
				break
			}
		}

	}

	return results
}

// GetDataFromHash - Gets data corresponding to a given hash. Returns nil if the hash is unknown
func (fileIndex *FileIndex) GetDataFromHash(hash []byte) []byte {
	// Grab the file index mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	if knownHash, ok := fileIndex.knownHashes[ToHex(hash[:])]; ok { // We know this hash

		sharedFile := knownHash.File

		if knownHash.IsMetahash {
			// Return the metafile
			return sharedFile.Metafile
		}
		// Return one of the file's chunk

		// Compute the filepath
		path := ""
		if sharedFile.IsDownloaded {
			path = PathToDownloadedFiles + sharedFile.Filename
		} else {
			path = PathToSharedFiles + sharedFile.Filename
		}

		// Open the file
		var f *os.File
		var err error
		if f, err = os.Open(path); err != nil {
			return nil
		}
		defer f.Close()

		// Create a buffer for the chunk
		chunkBuffer := make([]byte, ChunkSizeBytes)
		nbBytesRead := 0

		// Read in the file
		if _, err = f.Seek(int64(knownHash.ChunkIndex*ChunkSizeBytes), 0); err != nil {
			return nil
		}
		if nbBytesRead, err = f.Read(chunkBuffer); err != nil {
			return nil
		}

		// Return data
		return chunkBuffer[:nbBytesRead]
	}

	return nil
}

// AcknowledgeFileIndexed - To call when a file has been reconstructed entirely
func (fileIndex *FileIndex) AcknowledgeFileIndexed(filename string, metahash []byte) {
	// Send update to frontend
	frontend.FBuffer.AddFrontendIndexedFile(filename, ToHex(metahash[:]))
}

// WriteReceivedData - Write a received chunk at a file's end
func (fileIndex *FileIndex) WriteReceivedData(filename string, reply *messages.DataReply, chunkIndex uint64, isEmpty bool) {
	// Grab the file index mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	if shared, ok := fileIndex.index[filename]; ok { // We know the file

		// Open the file in write mode
		f, err := os.OpenFile(PathToDownloadedFiles+filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic("WriteReceivedData(): Failed to open file in write mode")
		}
		defer f.Close()

		// Only create the file if it is empty
		if isEmpty {
			return
		}

		// Write the chunk
		if nbBytesWrote, err := f.Write(reply.Data); err != nil || nbBytesWrote != len(reply.Data) {
			panic("WriteReceivedData(): Failed to write into file")
		}

		// Remember the hash
		fileIndex.knownHashes[ToHex(reply.HashValue[:])] = NewKnownHash(shared, false, chunkIndex)

	} else {
		panic("WriteReceivedData(): Trying to write to non-existent file")
	}

}

// SetMetafile - Sets the metafile for the file with the given filename
func (fileIndex *FileIndex) SetMetafile(filename string, reply *messages.DataReply) {
	// Grab the file index mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	if sharedFile, ok := fileIndex.index[filename]; ok { // We know this filename
		// Copy the data
		sharedFile.Metafile = make([]byte, len(reply.Data))
		copy(sharedFile.Metafile[:], reply.Data)
		// Remember the metahash
		fileIndex.knownHashes[ToHex(reply.HashValue[:])] = NewKnownHash(sharedFile, true, 0)
	}

}
