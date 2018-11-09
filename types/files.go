package types

import (
	"crypto/sha256"
	"os"
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
)

// FileIndex - Represents a file index
type FileIndex struct {
	index map[string]*SharedFile // A mapping from peer name to messages
	mux   sync.Mutex             // Mutex to manipulate the structure from different threads
}

// SharedFile - Represents a file indexed by the gossiper
type SharedFile struct {
	Filename      string     // Filename
	FilesizeBytes uint32     // Filesize in bytes
	Metafile      []byte     // Metafile (in RAM or disk)
	Metahash      []byte     // 32-bytes SHA-256 hash of metafile
	mux           sync.Mutex // Mutex to manipulate the structure from different threads
}

// NewFileIndex - Creates a new instance of FileIndex
func NewFileIndex() *FileIndex {
	var fileIndex FileIndex
	fileIndex.index = make(map[string]*SharedFile)
	return &fileIndex
}

// NewSharedFile - Creates a new instance of NewSharedFile
func NewSharedFile(filename string, filesize int64) *SharedFile {
	var shared SharedFile
	shared.Filename = filename
	shared.FilesizeBytes = uint32(filesize)
	shared.Metafile = nil
	shared.Metahash = make([]byte, HashSizeBytes)
	return &shared
}

// IndexNewFile - Indexes a new file
func (fileIndex *FileIndex) IndexNewFile(filename string) {

	// Open the file
	var f *os.File
	var err error
	if f, err = os.Open(PathToSharedFiles + filename); err != nil {
		return
	}
	defer f.Close()

	// Check the filesize (must not be too large)
	fi, err := f.Stat()
	if err != nil || fi.Size() > GetMaxFileSizeBytes() {
		return
	}

	// Create new SharedFile
	shared := NewSharedFile(filename, fi.Size())

	// Compute total number of chunks
	nbChunks := shared.FilesizeBytes / ChunkSizeBytes
	shared.Metafile = make([]byte, nbChunks*HashSizeBytes)

	// Buffer for chunk
	chunkBuffer := make([]byte, ChunkSizeBytes)
	nbBytesRead := 0

	// Create the metafile
	var metafileIndex uint32
	for metafileIndex = 0; metafileIndex < nbChunks; metafileIndex++ {

		// Read the next chunk
		if nbBytesRead, err = f.Read(chunkBuffer); err != nil {
			panic("IndexNewFile(): Indexed file could not be read correctly.")
		}

		// Create hash and append it to metafile
		chunkHash := sha256.Sum256(chunkBuffer[:nbBytesRead])
		if nbCopy := copy(shared.Metafile[metafileIndex*HashSizeBytes:], chunkHash[:]); nbCopy != HashSizeBytes {
			panic("IndexNewFile(): Metafile could not be generated.")
		}
	}

	// Create metahash
	metahash := sha256.Sum256(shared.Metafile[:])
	if nbCopy := copy(shared.Metahash[0:], metahash[:]); nbCopy != HashSizeBytes {
		panic("IndexNewFile(): Metafile could not be generated.")
	}

	// TODO: write the metafile to disk ?

	// Append the newly indexed file to the file index
	fileIndex.mux.Lock()
	defer fileIndex.mux.Unlock()
	if _, ok := fileIndex.index[filename]; ok { // We already have a file with the same name
		return
	}
	// We don't know this name
	fileIndex.index[filename] = shared

}

// GetMaxFileSizeBytes - Returns the maximum allowed size in bytes for a file
func GetMaxFileSizeBytes() int64 {
	return int64(ChunkSizeBytes * ChunkSizeBytes / HashSizeBytes)
}
