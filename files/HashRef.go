package files

/*HashRef is used as a reference to a `SharedFile` hash. This can either be the metahash,
in that case `ChunkIndex` is 0, or a chunk hash, with ID `ChunkIndex`.  */
type HashRef struct {
	File       *SharedFile // A pointer to the shared file concerned by this hash
	ChunkIndex uint64      // Indicates the chunk index in case (0 in case of metahash)
}

/*NewHashRef creates a new instance of HashRef.

`file` The `SharedFile` to be referenced.

`chunkIndex` The chunk ID corresponding to the hash (0 for a metahash).

A reference to the created object.*/
func NewHashRef(file *SharedFile, chunkIndex uint64) *HashRef {
	var HashRef HashRef
	HashRef.File = file
	HashRef.ChunkIndex = chunkIndex
	return &HashRef
}
