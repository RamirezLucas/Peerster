package files

// HashRef represents a known hash.
type HashRef struct {
	File       *SharedFile // A pointer to the shared file concerned by this hash
	ChunkIndex uint64      // Indicates the chunk index in case (0 in case of metahash)
}

// NewHashRef creates a new instance of HashRef
func NewHashRef(file *SharedFile, chunkIndex uint64) *HashRef {
	var HashRef HashRef
	HashRef.File = file
	HashRef.ChunkIndex = chunkIndex
	return &HashRef
}
