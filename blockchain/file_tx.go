package blockchain

import (
	"Peerster/crypto_rsa"
	"Peerster/fail"
	"Peerster/messages"
	"crypto/rsa"
)

type Tx struct {
	Signature [256]byte
	File      *messages.File
	PublicKey *rsa.PublicKey
}

func NewTx(publish *messages.TxPublish) *Tx {
	publicKey, err := crypto_rsa.BytesToPublicKey(publish.PublicKey)
	if err != nil {
		fail.HandleError(err)
		return nil
	}
	return &Tx{
		Signature: publish.Signature,
		File:      publish.File,
		PublicKey: publicKey,
	}
}

func (tx *Tx) ToTxPublish(hopLimit uint32) *messages.TxPublish {
	keyAsBytes, err := crypto_rsa.PublicKeyToBytes(tx.PublicKey)
	if err != nil {
		fail.HandleError(err)
		return nil
	}
	return &messages.TxPublish{
		Signature: tx.Signature,
		File:      tx.File,
		PublicKey: keyAsBytes,
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
