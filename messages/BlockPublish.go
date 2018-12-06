package messages

import (
	"crypto/sha256"
	"encoding/binary"
)

// BlockPublish - A block for the blockchain
type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

// Block - A blockchain's block
type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

// Hash - Computes the hash of a Block
func (b *BlockPublish) Hash() [32]byte {
	var out [32]byte
	h := sha256.New()
	h.Write(b.Block.PrevHash[:])
	h.Write(b.Block.Nonce[:])
	binary.Write(h, binary.LittleEndian, uint32(len(b.Block.Transactions)))
	for _, t := range b.Block.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return out
}
