package messages

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Number of zero bytes that every Block hash must start with
const nbBytesZero = 2

// Block - A blockchain's block
type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []*TxPublish
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

/*ChangeNonceRandomly genereates a new random nonce for a `Block` and
writes it in the `Nonce` field of the receiver block. */
func (b *Block) ChangeNonceRandomly() {
	rand.Read(b.Nonce[:])
}

/*CheckHashValid checks whether the block's hash (according the `Hash()`)
is valid for the blockchain, i.e. starts with enough zeroes.

The function returns true if the hash is valid, or false otherwise. */
func (b *Block) CheckHashValid() bool {

	hash := b.Hash()
	for i := 0; i < nbBytesZero; i++ {
		if hash[i] != 0 {
			return false // Invalid hash
		}
	}
	return true
}

/*ToString returns a textual representation of a `Block`.*/
func (b *Block) ToString() string {

	hash := b.Hash()

	str := fmt.Sprintf("%x", hash[:]) + ":"
	str += fmt.Sprintf("%x", b.PrevHash[:]) + ":"
	for _, tx := range b.Transactions {
		str += tx.File.Name + ","
	}

	return str[:len(str)-1]

}

/*IsGenesis returns true if this is a genesis block, i.e. a block with
`PrevHash` == 0. The function returns false otherwise.*/
func (b *Block) IsGenesis() bool {
	for i := 0; i < len(b.PrevHash); i++ {
		if b.PrevHash[i] != 0 {
			return false
		}
	}
	return true
}
