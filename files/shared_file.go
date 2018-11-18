package files

// KnownHash - Represents a known hash
type KnownHash struct {
	File       *SharedFile // A pointer to the shared file with this hash
	IsMetahash bool        // Indicates whether this is a metahash
	ChunkIndex uint32      // Indicates the chunk index in case this isn't a metahash
}

// SharedFile - Represents a file indexed by the gossiper
type SharedFile struct {
	Filename     string // The filename (read-only)
	Metafile     []byte // Metafile in RAM (read-only)
	Metahash     []byte // 32-bytes SHA-256 hash of metafile (read-only)
	IsDownloaded bool   // Indicates whether the file was indexed here first or dowloaded
}

// NewKnownHash - Creates a new instance of KnownHash
func NewKnownHash(file *SharedFile, isMetahash bool, chunkIndex uint32) *KnownHash {
	var knownHash KnownHash
	knownHash.File = file
	knownHash.IsMetahash = isMetahash
	knownHash.ChunkIndex = chunkIndex
	return &knownHash
}

// NewSharedFile - Creates a new instance of NewSharedFile
func NewSharedFile(filename string, metahash []byte, isDownloaded bool) *SharedFile {
	var shared SharedFile
	shared.Filename = filename
	shared.Metafile = nil
	if metahash != nil {
		shared.Metahash = metahash
	} else {
		shared.Metahash = make([]byte, HashSizeBytes)
	}
	shared.IsDownloaded = isDownloaded
	return &shared
}
