package messages

import (
	"Peerster/utils"
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
func (block *Block) Hash() [32]byte {
	var out [32]byte
	h := sha256.New()
	h.Write(block.PrevHash[:])
	h.Write(block.Nonce[:])
	binary.Write(h, binary.LittleEndian, uint32(len(block.Transactions)))
	for _, t := range block.Transactions {
		th := t.File.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return out
}

func (block *Block) HashString() string {
	hash := block.Hash()
	return utils.HashToHex(hash[:])
}

/*ChangeNonceRandomly genereates a new random nonce for a `Block` and
writes it in the `Nonce` field of the receiver block. */
func (block *Block) ChangeNonceRandomly() {
	rand.Read(block.Nonce[:])
}

/*CheckHashValid checks whether the block's hash (according the `Hash()`)
is valid for the blockchain, i.e. starts with enough zeroes.

The function returns true if the hash is valid, or false otherwise. */
func (block *Block) CheckHashValid() bool {

	hash := block.Hash()
	for i := 0; i < nbBytesZero; i++ {
		if hash[i] != 0 {
			return false // Invalid hash
		}
	}
	return true
}

/*ToString returns a textual representation of a `Block`.*/
func (block *Block) ToString() string {

	hash := block.Hash()

	str := fmt.Sprintf("%x", hash[:]) + ":"
	str += fmt.Sprintf("%x", block.PrevHash[:]) + ":"
	for _, tx := range block.Transactions {
		str += tx.File.Name + ","
	}

	return str[:len(str)-1]

}

/*IsGenesis returns true if this is a genesis block, i.e. a block with
`PrevHash` == 0. The function returns false otherwise.*/
func (block *Block) IsGenesis() bool {
	for i := 0; i < len(block.PrevHash); i++ {
		if block.PrevHash[i] != 0 {
			return false
		}
	}
	return true
}
