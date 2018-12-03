package files

import (
	"Peerster/fail"
	"Peerster/frontend"
	"Peerster/messages"
	"strings"
	"sync"
)

const (
	// PathToSharedFiles is the path to the folder where shared files are stored.
	PathToSharedFiles = "_SharedFiles/"
	// PathToDownloadedFiles is the path to the folder where downloaded files are stored.
	PathToDownloadedFiles = "_Downloads/"
)

// FileIndex represents a file index.
type FileIndex struct {
	index  map[string]*SharedFile // A mapping from metahash to SharedFile structures
	hashes map[string]*HashRef    // A mapping from a known hash to its corresponding file index
	mux    sync.Mutex             // Mutex to manipulate the structure from different threads
}

// NewFileIndex creates a new instance of FileIndex.
func NewFileIndex() *FileIndex {
	var fileIndex FileIndex
	fileIndex.index = make(map[string]*SharedFile)
	fileIndex.hashes = make(map[string]*HashRef)
	return &fileIndex
}

// AddMonoSourceFile adds a monosourced file named filename to the index with a given metahash.
// The file will be exlusively fetched from the peer origin.
func (fileIndex *FileIndex) AddMonoSourceFile(filename, origin string, metahash []byte) *SharedFile {

	newFile := NewSharedFileMonoSource(filename, metahash[:])
	hash := ToHex(metahash[:])

	// Grab the mutex on the index
	fileIndex.mux.Lock()
	if _, ok := fileIndex.index[hash]; ok { // We already have a file indexed with this metahash
		return nil
	}
	// Index the new file and unlock the mutex
	fileIndex.index[hash] = newFile
	fileIndex.mux.Unlock()

	// Send update to frontend
	frontend.FBuffer.AddFrontendConstructingFile(filename, hash, origin)

	return newFile
}

// AddMultiSourceFile adds a multisoured file with chunkCount chunks to the index with a given metahash.
// The file will be able to be fetched from multiple peers on the network.
func (fileIndex *FileIndex) AddMultiSourceFile(filename string, chunkCount uint64, metahash []byte) *SharedFile {

	newFile := NewSharedFileMultiSource(filename, chunkCount, metahash)
	hash := ToHex(metahash[:])

	// Grab the mutex on the index
	fileIndex.mux.Lock()
	if _, ok := fileIndex.index[hash]; ok { // We already have a file indexed with this metahash
		return nil
	}
	// Index the new file and unlock the mutex
	fileIndex.index[hash] = newFile
	fileIndex.mux.Unlock()

	// @]TODO Send update to frontend

	return newFile
}

// AddLocalFile indexes a new local file with the given filename. The file must
// located in the PathToDownloadedFiles folder.
func (fileIndex *FileIndex) AddLocalFile(filename string) {

	// Create new shared file
	shared := IndexLocalFile(filename)
	if shared == nil {
		fail.LeveledPrint(1, "IndexFile", `File %s could not be parsed`, filename)
		return
	}

	// Grab the mutex on the index
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	// Check if a file with the same metahash already exists in the database
	if _, ok := fileIndex.index[ToHex32(shared.Metahash)]; ok { // We already have a file with the same metahash
		fail.LeveledPrint(1, "IndexFile", `File %s could not be added to the index`, filename)
		return
	}

	// Add the chunk hashes to the set of known hashes
	for i := uint64(0); i < shared.ChunkCount; i++ {
		chunkHash := shared.Metafile[i*ChunkSizeBytes : (i+1)*ChunkSizeBytes]
		fileIndex.hashes[ToHex(chunkHash[:])] = NewHashRef(shared, i+1)
	}

	// Add the metahash to the set of known hashes
	fileIndex.hashes[ToHex32(shared.Metahash)] = NewHashRef(shared, 0)

	// Add the new indexed file to the index
	fileIndex.index[ToHex32(shared.Metahash)] = shared

	// Send update to frontend
	shared.AcknowledgeFileReconstructed()
}

// GetDataFromHash gets data corresponding to a given hash. Returns nil if the hash is unknown.
func (fileIndex *FileIndex) GetDataFromHash(hash []byte) []byte {
	// Grab the file index mutex
	fileIndex.mux.Lock()

	if HashRef, ok := fileIndex.hashes[ToHex(hash[:])]; ok { // We know this hash
		// Unlock mutex and get the chunk
		fileIndex.mux.Unlock()
		return HashRef.File.GetChunk(HashRef.ChunkIndex)
	}

	// Unlock mutex and return
	fileIndex.mux.Unlock()
	return nil
}

// HandleDataReply handles a DataReply for a DataRequest represented by HashRef. Depending on the
// DataRequest either the metafile or a chunk is written. The function returns the next chunk to fetch
// for this file if there is one (indices start at 1), as well as the peer to fetch it from. If there
// is no next chunk to fetch the function returns (0, "").
func (fileIndex *FileIndex) HandleDataReply(ref *HashRef, reply *messages.DataReply) (uint64, string) {

	shared := ref.File
	if ref.ChunkIndex == 0 { // Metafile in reply.Data
		if shared.SetMetafile(reply) { // Reconstruction complete (empty file)
			fileIndex.AddHashRef(ToHex(reply.HashValue[:]), &HashRef{File: shared, ChunkIndex: 0})
			return 0, "" // Stop requesting
		}

		// Decide to whom to request the next chunk
		if shared.IsMonosource {
			return 1, reply.Origin // Request first chunk
		}
		if target, ok := shared.RemoteChunks[1]; ok {
			return 1, target // Request first chunk
		}
		fail.CustomPanic("HandleDataReply", "No known target for file %s chunk 1",
			shared.Filename)
	}

	// Chunk in reply.Data
	if shared.WriteChunk(ref.ChunkIndex, reply.Data) {
		fileIndex.AddHashRef(ToHex(reply.HashValue[:]), &HashRef{File: shared, ChunkIndex: ref.ChunkIndex})
		return 0, "" // Stop requesting
	}

	// Decide to whom to request the next chunk
	if shared.IsMonosource {
		return ref.ChunkIndex + 1, reply.Origin // Request next chunk
	}
	if target, ok := shared.RemoteChunks[1]; ok {
		return ref.ChunkIndex + 1, target // Request next chunk
	}
	fail.CustomPanic("HandleDataReply", "No known target for file %s chunk %d",
		shared.Filename, ref.ChunkIndex+1)

	// Unreachable
	return 0, ""
}

// HandleSearchRequest returns the list of SearchResult's corresponding to files whose
// filename contains at least one of the keyword contained in the keywords slice.
func (fileIndex *FileIndex) HandleSearchRequest(keywords []string) []*messages.SearchResult {
	// Grab the mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	results := make([]*messages.SearchResult, 0)

	// Iterate over all known files
	for _, shared := range fileIndex.index {

		// Search for a keyword in the filename
		for _, k := range keywords {
			if strings.Contains(shared.Filename, k) {
				// We have a match
				if ret := shared.GetFileSearchInfo(); ret != nil {
					results = append(results, ret)
				}
				break
			}
		}
	}

	return results
}

/*HandleSearchResult @TODO*/
func (fileIndex *FileIndex) HandleSearchResult(result *messages.SearchResult, origin string) bool {
	// Grab the mutex
	fileIndex.mux.Lock()

	if shared, ok := fileIndex.index[ToHex(result.MetafileHash[:])]; ok { // We know this metahash
		// Unlock the mutex and update the shared file with new remote chunk mappings
		fileIndex.mux.Unlock()
		return shared.UpdateChunkMappings(result.ChunkMap, origin)
	}

	// Create a new multisource shared file and unlock the mutex
	newFile := NewSharedFileMultiSource(result.Filename, result.ChunkCount, result.MetafileHash)
	fileIndex.index[ToHex(result.MetafileHash[:])] = newFile
	fileIndex.mux.Unlock()

	// @TODO what to do with frontend ?

	// Update the shared file with new remote chunk mappings
	return newFile.UpdateChunkMappings(result.ChunkMap, origin)
}

// AddHashRef adds a hash to the index of known hashes.
func (fileIndex *FileIndex) AddHashRef(hash string, ref *HashRef) {
	// Grab the mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	fileIndex.hashes[hash] = ref
}
