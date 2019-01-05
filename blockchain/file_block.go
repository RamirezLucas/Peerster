package blockchain

import (
	"Peerster/messages"
	"Peerster/utils"
	"fmt"
	"sort"
	"strings"
)

type FileBlock struct {
	Length int
	id     string

	Previous     *FileBlock
	Hash         [32]byte
	Nonce        [32]byte
	Transactions []*Tx
}

func (fb *FileBlock) IsGenesis() bool {
	return fb.Previous == nil
}

func (fb *FileBlock) ToBlock(hopLimit uint32) *messages.Block {
	transactions := []*messages.TxPublish{}
	for _, tx := range fb.Transactions {
		transactions = append(transactions, tx.ToTxPublish(hopLimit))
	}

	prevHash := [32]byte{}
	if fb.Previous != nil {
		prevHash = fb.Previous.Hash
	}
	return &messages.Block{
		PrevHash:     prevHash,
		Nonce:        fb.Nonce,
		Transactions: transactions,
	}
}

func (fb *FileBlock) ToBlockPublish(hopLimit uint32) *messages.BlockPublish {
	return &messages.BlockPublish{
		Block:    fb.ToBlock(hopLimit),
		HopLimit: hopLimit,
	}
}

func (fb *FileBlock) String() string {
	prevHash := [32]byte{}
	if fb.Previous != nil {
		prevHash = fb.Previous.Hash
	}
	txStrings := []string{}
	for _, tx := range fb.Transactions {
		txStrings = append(txStrings, tx.File.Name)
	}
	sort.Slice(txStrings, func(i, j int) bool {
		return txStrings[i] < txStrings[j]
	})
	return fmt.Sprintf("%s:%s:%s", utils.HashToHex(fb.Hash[:]), utils.HashToHex(prevHash[:]), strings.Join(txStrings, ","))
}

func (fb *FileBlock) ChainString() string {
	blockStrings := []string{}
	block := fb
	for block != nil {
		blockStrings = append(blockStrings, block.String())
		block = block.Previous
	}
	return fmt.Sprintf("CHAIN %s", strings.Join(blockStrings, " "))
}
