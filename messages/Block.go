package messages

import (
	"crypto/sha256"
	"encoding/binary"
)

// Block - A blockchain's block
type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

// Hash - Computes the hash of a Block
func (b *Block) Hash() [32]byte {
	var out [32]byte
	h := sha256.New()
	h.Write(b.PrevHash[:])
	h.Write(b.Nonce[:])
	binary.Write(h, binary.LittleEndian, uint32(len(b.Transactions)))
	for _, t := range b.Transactions {
		th := t.File.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return out
}
