package messages

import (
	"Peerster/utils"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// File - A file identificator
type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

// Hash - Computes the hash of a File
func (t *File) Hash() [32]byte {
	var out [32]byte
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(t.Name)))
	h.Write([]byte(t.Name))
	h.Write(t.MetafileHash)
	copy(out[:], h.Sum(nil))
	return out
}

func (file *File) String() string {
	return fmt.Sprintf("FILE named %s size %d with metafile hash %s",
		file.Name, file.Size, utils.HashToHex(file.MetafileHash))
}
