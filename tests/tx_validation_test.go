package tests

import (
	"Peerster/blockchain"
	"Peerster/crypto_rsa"
	"Peerster/messages"
	"Peerster/utils"
	"crypto/rsa"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockBuilder(t *testing.T) {
	fbb := blockchain.NewFileBlockBuilder(nil)

	assert.NotNil(t, fbb, "block builder constructor does not return nil")
	assert.Equal(t, 1, fbb.Length)
	assert.Equal(t, 0, len(fbb.Transactions))
}

func TestTxCreation(t *testing.T) {
	_, tx := newTx()
	assert.NotNil(t, tx, "tx should not be nil")
}

func TestAddOwner(t *testing.T) {
	_, tx := newTx()
	fbb := blockchain.NewFileBlockBuilder(nil)

	assert.True(t, fbb.AddTxIfValid(tx))
	assert.Equal(t, 1, len(fbb.Transactions))
}

func TestChangeOwner(t *testing.T) {
	ownerKey, tx := newTx()
	fbb := createFBB(t, tx)

	signature, err := crypto_rsa.Sign(tx.Signature[:], ownerKey)

	assert.Nil(t, err)

	newOwnerKey := crypto_rsa.GeneratePrivateKey()
	newTx := &blockchain.Tx{
		Signature: signature,
		File:      tx.File,
		PublicKey: &newOwnerKey.PublicKey,
	}

	assert.True(t, fbb.AddTxIfValid(newTx))
	assert.Equal(t, 1, len(fbb.Transactions))
	assert.Equal(t, 1, len(fbb.Filenames))
}

func TestTryChangeOwner(t *testing.T) {
	_, tx := newTx()
	fbb := createFBB(t, tx)

	newOwnerKey := crypto_rsa.GeneratePrivateKey()
	signature, err := crypto_rsa.Sign(tx.Signature[:], newOwnerKey)

	assert.Nil(t, err)

	newTx := &blockchain.Tx{
		Signature: signature,
		File:      tx.File,
		PublicKey: &newOwnerKey.PublicKey,
	}

	assert.False(t, fbb.AddTxIfValid(newTx))
	assert.Equal(t, 0, len(fbb.Transactions))
	assert.Equal(t, 1, len(fbb.Filenames))
}

func TestAddSecondTx(t *testing.T) {
	_, tx := newTx()
	fbb := createFBB(t, tx)

	_, newTx := newTx()

	assert.True(t, fbb.AddTxIfValid(newTx))
	assert.Equal(t, 1, len(fbb.Transactions))
	assert.Equal(t, 2, len(fbb.Filenames))
}

func TestTryAddSecondTx(t *testing.T) {
	_, tx := newTx()
	fbb := createFBB(t, tx)

	_, newTx := newTx()
	newTx.File.Name = tx.File.Name

	assert.False(t, fbb.AddTxIfValid(newTx))
	assert.Equal(t, 0, len(fbb.Transactions))
	assert.Equal(t, 1, len(fbb.Filenames))
}

// private functions

func createFBB(t *testing.T, tx *blockchain.Tx) *blockchain.FileBlockBuilder {
	genesis := blockchain.NewFileBlockBuilder(nil)
	genesis.AddTxIfValid(tx)

	fbb := mineAndGetNextBlock(genesis)

	assert.Equal(t, 2, fbb.Length)
	assert.Equal(t, 0, len(fbb.Transactions))
	assert.Equal(t, 1, len(fbb.Filenames))

	return fbb
}

func mineAndGetNextBlock(genesis *blockchain.FileBlockBuilder) *blockchain.FileBlockBuilder {
	var fb *blockchain.FileBlock
	for fb == nil {
		genesis.SetNonce(utils.Random32Bytes())
		fb, _ = genesis.Build()
	}
	return blockchain.NewFileBlockBuilder(fb)
}

func newTx() (*rsa.PrivateKey, *blockchain.Tx) {

	ownerKey := crypto_rsa.GeneratePrivateKey()
	someBytes := utils.Random32Bytes()
	file := &messages.File{
		Name:         fmt.Sprintf("%s.%s", utils.HashToHex(someBytes[:4]), utils.HashToHex(someBytes[4:5])),
		Size:         32,
		MetafileHash: someBytes[:],
	}
	fileHash := file.Hash()
	signature, err := crypto_rsa.Sign(fileHash[:], ownerKey)
	if err != nil {
		common.HandleError(err)
		return nil, nil
	}
	return ownerKey, &blockchain.Tx{
		Signature: signature,
		File:      file,
		PublicKey: &ownerKey.PublicKey,
	}
}
