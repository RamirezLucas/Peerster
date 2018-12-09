package files

import (
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

/*FileIndex represents the set of files indexed or known by the gossiper. The object contains an index
mapping each known metahash to its corresponding `SharedFile` (`index`). The object also contains a mapping
from every known hash (metahash or chunk hash) to its corresponding `SharedFile` (`hashes`).

A FileIndex object should be created by calling `NewFileIndex()`. Once created, the object is thread-safe,
meaning that several threads may manipulate the object through its API simultaneously.*/
type FileIndex struct {
	index  map[string]*SharedFile // A mapping from metahash to SharedFile structures
	hashes map[string]*HashRef    // A mapping from a known hash to its corresponding SharedFile
	mux    sync.Mutex             // Mutex to manipulate the structure from different threads
}

/*NewFileIndex creates a new instance of NewFileIndex.*/
func NewFileIndex() *FileIndex {
	var fileIndex FileIndex
	fileIndex.index = make(map[string]*SharedFile)
	fileIndex.hashes = make(map[string]*HashRef)
	return &fileIndex
}

/*AddMonoSourceFile adds a monosourced file to the `FileIndex`. This file must be fetched from a single
source `origin` that is determined by the user.

`filename` The filename under which the file will be written to disk.

`origin` The peer from whom to download the file.

`metahash` The file's metahash.

The function returns a pointer to the created `SharedFile` on success, or `nil` if a file with the same
metahash already exists. */
func (fileIndex *FileIndex) AddMonoSourceFile(filename, origin string, metahash []byte) *SharedFile {

	newFile := NewSharedFileMonoSource(filename, metahash[:])
	hash := ToHex(metahash[:])

	// Grab the mutex on the index
	fileIndex.mux.Lock()
	if shared, ok := fileIndex.index[hash]; ok { // We already have a file indexed with this metahash
		// Unlock the mutex and attempts to change the file status
		fileIndex.mux.Unlock()
		if shared.SwitchMultiToMonoSource(filename) {
			return shared
		}

		// Failed to change the file status
		return nil
	}
	// Index the new file
	fileIndex.index[hash] = newFile

	// Unlock the mutex
	fileIndex.mux.Unlock()

	return newFile
}

/*AddLocalFile adds a locally stored file to the `FileIndex`. This file must be stored in the
`PathToSharedFiles` directory. All of the hashes (metahash and chunk hashes) generated from this
file are stored in the `FileIndex`'s `hashes` map.

`filename` The file to index's filename.

The functions returs a pointer to a `File` (for the blockchain).*/
func (fileIndex *FileIndex) AddLocalFile(filename string) *messages.File {

	// Create new shared file
	shared, filesize := IndexLocalFile(filename)
	if shared == nil {
		return nil
	}

	// Grab the mutex on the index
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	// Check if a file with the same metahash already exists in the database
	if _, ok := fileIndex.index[ToHex32(shared.Metahash)]; ok { // We already have a file with the same metahash
		return nil
	}

	// Add the chunk hashes to the set of known hashes
	for i := uint64(0); i < shared.ChunkCount; i++ {
		chunkHash := shared.Metafile[i*HashSizeBytes : (i+1)*HashSizeBytes]
		fileIndex.hashes[ToHex(chunkHash[:])] = NewHashRef(shared, i+1)
	}

	// Add the metahash to the set of known hashes
	fileIndex.hashes[ToHex32(shared.Metahash)] = NewHashRef(shared, 0)

	// Add the new indexed file to the index
	fileIndex.index[ToHex32(shared.Metahash)] = shared

	// Send update to frontend
	shared.AcknowledgeFileReconstructed()

	// Create return value
	tmp := make([]byte, HashSizeBytes)
	copy(tmp[:], shared.Metahash[:])
	return &messages.File{
		Name:         filename,
		Size:         filesize,
		MetafileHash: tmp,
	}
}

/*GetDataFromHash reads the bytes corresponding to a provided hash (metafile or file chunk).
The `hash` is looked for in the `FileIndex`'s `hashes` map.

`hash` The hash to look for.

The function returns a slice of bytes containing the requested data on success, or `nil` on failure.*/
func (fileIndex *FileIndex) GetDataFromHash(hash []byte) []byte {
	// Grab the file index mutex
	fileIndex.mux.Lock()

	if ref, ok := fileIndex.hashes[ToHex(hash[:])]; ok { // We know this hash
		// Unlock mutex and get the chunk
		fileIndex.mux.Unlock()
		return ref.File.GetChunk(ref.ChunkIndex)
	}

	// Unlock mutex and return
	fileIndex.mux.Unlock()
	return nil
}

/*HandleDataReply handles an incoming `DataReply` for which the gossiper found the associated `DataRequest`
that originated it. The file concerned by this reply is passed in a `HashRef`. Depending on the value of
`ref.ChunkID` either the metafile or one of the file's chunk is written.

`ref` A `HashRef` referencing the file concerned by this reply.

`reply` A received `DataReply` that originated from a known `DataRequest`

If the corresponding file is still incomplete after taking into account the new `DataReply` the function
returns a tuple indicating which chunk ID to request next and to whom. If `reply` caused file reconstruction
to complete then (0, "") is returned.
*/
func (fileIndex *FileIndex) HandleDataReply(ref *HashRef, reply *messages.DataReply) (uint64, string) {

	shared := ref.File
	if ref.ChunkIndex == 0 { // Metafile in reply.Data
		if shared.SetMetafile(reply) { // Reconstruction complete (empty file)
			ref := NewHashRef(shared, 0)
			fileIndex.addHashRef(ToHex(reply.HashValue[:]), ref)
			return 0, "" // Stop requesting
		}

		// Decide to whom to request the next chunk
		return 1, shared.GetChunkTarget(1, reply.Origin)
	}

	// Chunk in reply.Data
	if shared.WriteChunk(ref.ChunkIndex, reply.Data) {
		ref := NewHashRef(shared, ref.ChunkIndex)
		fileIndex.addHashRef(ToHex(reply.HashValue[:]), ref)
		return 0, "" // Stop requesting
	}

	// Decide to whom to request the next chunk
	return ref.ChunkIndex + 1, shared.GetChunkTarget(ref.ChunkIndex+1, reply.Origin)
}

/*HandleSearchRequest handles an incoming `SearchRequest` by searching the `FileIndex` for any
filename containing any of the keywords contained in the `SearchRequest.Keywords` slice. This function
is destined to be used when sending a `SearchReply`.

`search` The SearchRequest to evaluate.

The function returns a slice of `SearchResult`'s containing all matching files according to the
above criteria. The function returns `nil` if no matches were found. */
func (fileIndex *FileIndex) HandleSearchRequest(search *messages.SearchRequest) []*messages.SearchResult {
	// Grab the mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	results := make([]*messages.SearchResult, 0)

	// Iterate over all known files
	for _, shared := range fileIndex.index {

		// Search for a keyword in the filename
		for _, k := range search.Keywords {
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

/*HandleSearchResult handles a `SearchResult` contained in an incoming `SearchReply`. The function
updates the remote chunk mappings for the concerned file (or create them if they don't exist yet)
with the newest mappings contained in `result`.

`result` The `SearchResult` from which new mappings are taken.

`origin` The `SearchResult`'s origin.

The function returns true if the added mappings just triggered a complete match, or false otherwise.*/
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

	// Update the shared file with new remote chunk mappings
	return newFile.UpdateChunkMappings(result.ChunkMap, origin)
}

/*GetMetafileTargetMultisource returns the name of one of the peers possessing at least one
chunk of a multi-sourced file. Such peers necessarily have the file's metafile and so the gossiper
can request it from any one of them. The function panics if `metahash` is unknown

`metahash` The metahash of the file for which we wish to obtain the metafile.

The function returns a tuple where the first element is the peer's name or an empty string if the file
isn't in the `CompleteMatch` state. The second element points to a valid `SharedFile` whose metahash
is `metahash` or nil if such a file does not exists.
*/
func (fileIndex *FileIndex) GetMetafileTargetMultisource(metahash []byte) (string, *SharedFile) {
	// Grab the mutex
	fileIndex.mux.Lock()

	if shared, ok := fileIndex.index[ToHex(metahash[:])]; ok { // We know this metahash
		// Unlock the mutex and update the shared file with new remote chunk mappings
		fileIndex.mux.Unlock()
		return shared.GetMetafileQueryPeer(), shared
	}

	fileIndex.mux.Unlock()
	return "", nil
}

/*CheckHashPresent checks whether a hash is already known by the `FileIndex`.

`hash` The hash to check.

The function returns a `HashRef` containing information about the hash if it exists,
or nil if the hash is unknown.
*/
func (fileIndex *FileIndex) CheckHashPresent(hash []byte) *HashRef {
	// Grab the mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	// Check if a hash exists
	if ref, ok := fileIndex.hashes[ToHex(hash[:])]; ok {
		return ref
	}
	return nil
}

func (fileIndex *FileIndex) addHashRef(hash string, ref *HashRef) {
	// Grab the mutex
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()

	fileIndex.hashes[hash] = ref
}
