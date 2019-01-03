package blockchain

import (
	"Peerster/messages"
	"Peerster/utils"
	"fmt"
	"sort"
	"strings"
)

type FileBlock struct {
	length int
	id     string

	previous     *FileBlock
	hash         [32]byte
	nonce        [32]byte
	transactions map[string]*Tx
}

func (fb *FileBlock) IsAfterGenesis() bool {
	return fb.previous == nil
}

func (fb *FileBlock) txIsValidWithThisBlock(newTx *Tx) bool {
	_, ok := fb.transactions[newTx.id]
	return !ok //is valid if not yet present
}

func (fb *FileBlock) ToBlock(hopLimit uint32) *messages.Block {
	transactions := []messages.TxPublish{}
	for _, tx := range fb.transactions {
		transactions = append(transactions, tx.ToTxPublish(hopLimit))
	}

	prevHash := [32]byte{}
	if fb.previous != nil {
		prevHash = fb.previous.hash
	}
	return &messages.Block{
		PrevHash:     prevHash,
		Nonce:        fb.nonce,
		Transactions: transactions,
	}
}

func (fb *FileBlock) ToBlockPublish(hopLimit uint32) *messages.BlockPublish {
	return &messages.BlockPublish{
		Block:    *fb.ToBlock(hopLimit),
		HopLimit: hopLimit,
	}
}

func (fb *FileBlock) String() string {
	prevHash := [32]byte{}
	if fb.previous != nil {
		prevHash = fb.previous.hash
	}
	txStrings := []string{}
	for _, tx := range fb.transactions {
		txStrings = append(txStrings, tx.File.Name)
	}
	sort.Slice(txStrings, func(i, j int) bool {
		return txStrings[i] < txStrings[j]
	})
	return fmt.Sprintf("%s:%s:%s", utils.HashToHex(fb.hash[:]), utils.HashToHex(prevHash[:]), strings.Join(txStrings, ","))
}

func (fb *FileBlock) ChainString() string {
	blockStrings := []string{}
	block := fb
	for block != nil {
		blockStrings = append(blockStrings, block.String())
		block = block.previous
	}
	return fmt.Sprintf("CHAIN %s", strings.Join(blockStrings, " "))
}
