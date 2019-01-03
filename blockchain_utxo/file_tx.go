package blockchain

import (
	"Peerster/messages"
)

type Tx struct {
	id   string
	File messages.File
}

func NewTx(publish *messages.TxPublish) *Tx {
	//hash := publish.File.Hash()
	return &Tx{
		id:   publish.File.Name,
		File: publish.File,
	}
}

func (tx *Tx) IsValid(tx2 *Tx) bool {
	return tx.File.Name != tx2.File.Name
}

func (tx *Tx) ToTxPublish(hopLimit uint32) messages.TxPublish {
	return messages.TxPublish{
		File:     tx.File,
		HopLimit: hopLimit,
	}
}
