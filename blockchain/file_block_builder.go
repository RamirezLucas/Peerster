package blockchain

import (
	"Peerster/crypto_rsa"
	"Peerster/fail"
	"Peerster/logger"
	"Peerster/messages"
	"Peerster/utils"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
)

const (
	NumOfZeroBytes = 2
)

type FileBlockBuilder struct {
	Length int

	Previous *FileBlock
	prevHash [32]byte // needed only if previous is nil
	nonce    [32]byte

	Transactions []*Tx
	Filenames    map[string]bool
	Hashes       map[string]*Tx

	sync.RWMutex
}

func NewFileBlockBuilder(previousBlock *FileBlock) *FileBlockBuilder {
	fbb := &FileBlockBuilder{
		Length: 1,

		Previous: previousBlock,
		prevHash: [32]byte{}, //only us
		nonce:    [32]byte{},

		Transactions: []*Tx{},
		Filenames:    map[string]bool{},
		Hashes:       map[string]*Tx{},
	}
	if previousBlock != nil {
		fbb.Length = previousBlock.Length + 1
	}
	fbb.readPreviousTx()
	return fbb
}

func (fbb *FileBlockBuilder) SetNonce(nonce [32]byte) {
	fbb.Lock()
	defer fbb.Unlock()

	fbb.nonce = nonce
}

func (fbb *FileBlockBuilder) AddTxIfValid(newTx *Tx) bool {
	fbb.Lock()
	defer fbb.Unlock()

	return fbb.addTxIfValid(newTx)
}

func (fbb *FileBlockBuilder) SetBlockAndBuild(block *messages.Block) (*FileBlock, error) {
	fbb.Lock()

	if fbb.Previous != nil && fbb.Previous.Hash != block.PrevHash {
		return nil, fmt.Errorf("trying to add a block over a mismatching previous file-block")
	}

	fbb.Transactions = []*Tx{} // clear previous entries in transactions if they were some
	for _, txPublish := range block.Transactions {
		tx := NewTx(txPublish)
		if !fbb.addTxIfValid(tx) { // one tx contradicts another
			return nil, fmt.Errorf("one tx (%s) contradicts another previous tx", tx.File.String())
		}
	}
	fbb.prevHash = block.PrevHash // in case previous is nil when computing hash (prevHash needed)
	fbb.nonce = block.Nonce

	fbb.Unlock()
	return fbb.Build()
}

func (fbb *FileBlockBuilder) Build() (*FileBlock, error) {
	fbb.RLock()
	defer fbb.RUnlock()

	hash := fbb.Hash()
	if !utils.FirstNZero(NumOfZeroBytes, hash[:]) { // checking if hash is truly starting with `NumOfZeroBytes` bytes
		return nil, fmt.Errorf("hash needs to have %d leading bits set to zeros (%d bytes)", NumOfZeroBytes*8, NumOfZeroBytes)
	}

	return &FileBlock{
		Length:       fbb.Length,
		id:           utils.HashToHex(hash[:]),
		Previous:     fbb.Previous,
		Hash:         hash,
		Nonce:        fbb.nonce,
		Transactions: fbb.Transactions,
	}, nil
}

func (fbb *FileBlockBuilder) Hash() (out [32]byte) {
	fbb.RLock()
	defer fbb.RUnlock()

	previousHash := fbb.prevHash
	if fbb.Previous != nil {
		previousHash = fbb.Previous.Hash
	}

	h := sha256.New()
	h.Write(previousHash[:])
	h.Write(fbb.nonce[:])
	err := binary.Write(h, binary.LittleEndian, uint32(len(fbb.Transactions)))
	if err != nil {
		fail.HandleAbort("unexpected error when computing hash of block", err)
		return
	}
	for _, t := range fbb.Transactions {
		th := t.File.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

// private functions without locks

func (fbb *FileBlockBuilder) readPreviousTx() {
	previousBlock := fbb.Previous
	for previousBlock != nil {
		for _, tx := range previousBlock.Transactions {
			fbb.Filenames[tx.File.Name] = true

			// adding latest public keys (owners)
			fileHashString := tx.File.HashString()
			if _, ok := fbb.Hashes[fileHashString]; !ok {
				fbb.Hashes[fileHashString] = tx
			}
		}
		previousBlock = previousBlock.Previous
	}
}

func (fbb *FileBlockBuilder) addTxIfValid(newTx *Tx) bool {
	fileHashString := newTx.File.HashString()
	prevTx, ok := fbb.Hashes[fileHashString]

	if !ok {
		// if new hash (file), we check if filename is not already used
		if _, ok := fbb.Filenames[newTx.File.Name]; ok {
			logger.Printlnf("IGNORING TX: filename <%s> already used", newTx.File.Name)
			return false
		}
	} else if crypto_rsa.Verify(prevTx.Signature[:], newTx.Signature, prevTx.PublicKey) != nil {
		// check if changing ownership is legal here (i.e. if owner is the one starting the change)
		logger.Printlnf("IGNORING TX: there is already an owner of file <%s>", newTx.File.String())
		return false
	}

	// printing the transaction result
	if !ok {
		logger.Printlnf("ADDING TX: new owner of file <%s>", newTx.File.String())
	} else {
		logger.Printlnf("ADDING TX: owner of file <%s> changed", newTx.File.String())
	}

	fbb.Filenames[newTx.File.Name] = true
	fbb.Hashes[fileHashString] = newTx
	fbb.Transactions = append(fbb.Transactions, newTx)
	return true
}
