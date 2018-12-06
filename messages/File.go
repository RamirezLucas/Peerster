package messages

import (
	"crypto/sha256"
	"encoding/binary"
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
