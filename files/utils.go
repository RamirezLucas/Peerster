package files

import (
	"Peerster/fail"
	"crypto/sha256"
	"fmt"
	"os"
)

// ToHex returns the hexadecimal string representation of a hash
func ToHex(hash []byte) string {
	return fmt.Sprintf("%x", hash[:])
}

// ToHex32 returns the hexadecimal string representation of a 32-bytes hash
func ToHex32(hash [32]byte) string {
	return fmt.Sprintf("%x", hash[:])
}

// GetChunksNumberFromRawFile returns the number of chunks from the filesize
func GetChunksNumberFromRawFile(fileSize int) uint64 {
	nbChunks := uint64(fileSize / ChunkSizeBytes)
	if fileSize%ChunkSizeBytes != 0 {
		nbChunks++
	}
	return nbChunks
}

// IndexLocalFile indexes a new file named filename stored in the PathToSharedFiles folder.
func IndexLocalFile(filename string) (*SharedFile, int64) {

	// Open the file
	var f *os.File
	var err error
	if f, err = os.Open(PathToSharedFiles + filename); err != nil {
		fail.LeveledPrint(1, "IndexLocalFile", `Failed to open file %s`, filename)
		return nil, 0
	}
	defer f.Close()

	// Check the filesize (must not be too large)
	fi, err := f.Stat()
	if err != nil || fi.Size() > MaxFileSizeBytes {
		fail.LeveledPrint(1, "IndexLocalFile", `File %s is too large (%d bytes)`, filename, fi.Size())
		return nil, 0
	}

	// Compute total number of chunks
	nbChunks := GetChunksNumberFromRawFile(int(fi.Size()))

	// Create a new SharedFile
	shared := NewSharedFileLocal(filename, nbChunks)

	// Buffer for chunk
	chunkBuffer := make([]byte, ChunkSizeBytes)
	nbBytesRead := 0

	// Create the metafile
	for metafileIndex := uint64(0); metafileIndex < nbChunks; metafileIndex++ {

		// Read the next chunk
		if nbBytesRead, err = f.Read(chunkBuffer); err != nil {
			fail.CustomPanic("IndexLocalFile", "File %s could not be read correctly.", filename)
		}

		// Create hash and append it to metafile
		chunkHash := sha256.Sum256(chunkBuffer[:nbBytesRead])
		if nbCopy := copy(shared.Metafile[metafileIndex*HashSizeBytes:], chunkHash[:]); nbCopy != HashSizeBytes {
			fail.CustomPanic("IndexLocalFile", "Metafile could not be generated for file %s.", filename)
		}

	}

	// Create metahash
	metahash := sha256.Sum256(shared.Metafile[:])
	if nbCopy := copy(shared.Metahash[:], metahash[:]); nbCopy != HashSizeBytes {
		fail.CustomPanic("IndexLocalFile", "Metahash could not be generated for file %s.", filename)
	}

	// Return the created shared file
	return nil, fi.Size()
}
