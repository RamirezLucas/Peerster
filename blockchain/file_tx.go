package blockchain

import (
	"Peerster/messages"
	"crypto/rsa"
)

type Tx struct {
	Signature [256]byte
	File      *messages.File
	PublicKey *rsa.PublicKey
}

func NewTx(publish *messages.TxPublish) *Tx {
	return &Tx{
		Signature: publish.Signature,
		File:      publish.File,
		PublicKey: publish.PublicKey,
	}
}

func (tx *Tx) ToTxPublish(hopLimit uint32) *messages.TxPublish {
	return &messages.TxPublish{
		Signature: tx.Signature,
		File:      tx.File,
		PublicKey: tx.PublicKey,
		HopLimit:  hopLimit,
	}
}

func (this *Tx) Equals(that *Tx) bool {
	return that != nil &&
		this.Signature == that.Signature &&
		this.File.Hash() == that.File.Hash() &&
		this.PublicKey.E == that.PublicKey.E &&
		this.PublicKey.N.Cmp(that.PublicKey.N) == 0
}
