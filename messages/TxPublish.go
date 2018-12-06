package messages

import (
	"crypto/sha256"
	"encoding/binary"
)

// TxPublish - A name-to-methash mapping for the blockchain
type TxPublish struct {
	File     File
	HopLimit uint32
}

// File - A file identificator
type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

// Hash - Computes the hash of a TxPublish
func (t *TxPublish) Hash() [32]byte {
	var out [32]byte
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(t.File.Name)))
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return out
}
