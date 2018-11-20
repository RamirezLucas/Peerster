package files

import "fmt"

// ToHex - Returns the hexadecimal string representations of a hash
func ToHex(hash []byte) string {
	return fmt.Sprintf("%x", hash[:])
}

// GetChunksNumberFromMetafile - Returns the number of chunks from the size of the metafile
func GetChunksNumberFromMetafile(metafileSize int) uint64 {
	nbChunks := uint64(metafileSize / HashSizeBytes)
	if metafileSize%HashSizeBytes != 0 {
		nbChunks++
	}
	return nbChunks
}

// GetChunksNumberFromRawFile - Returns the number of chunks from the filesize
func GetChunksNumberFromRawFile(fileSize int) uint64 {
	nbChunks := uint64(fileSize / ChunkSizeBytes)
	if fileSize%ChunkSizeBytes != 0 {
		nbChunks++
	}
	return nbChunks
}
