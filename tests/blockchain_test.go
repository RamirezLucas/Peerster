package tests

import (
	"Peerster/blockchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createBCF(t *testing.T) *blockchain.BCF {
	bcf := blockchain.NewBCF()
	assert.NotNil(t, bcf)
	assert.Equal(t, 0, bcf.ChainLength)
	return bcf
}

func TestBlockChainMiningRoutine(t *testing.T) {
	_, tx := newTx()
	bcf := blockchain.NewBCF()
	go bcf.MiningRoutine()
	bcf.AddTx(tx)
	block := bcf.MineChan.Get()

	assert.True(t, block.Transactions[0].Equals(tx))
	assert.Equal(t, 1, bcf.ChainLength)
}

func TestBlockChainForkShorterAndLonger(t *testing.T) {
	_, tx := newTx()
	bcf := blockchain.NewBCF()
	initHead := bcf.Head

	// now mining normally
	go bcf.MiningRoutine()
	bcf.AddTx(tx)
	block := bcf.MineChan.Get()
	assert.True(t, block.Transactions[0].Equals(tx))
	assert.Equal(t, 1, bcf.ChainLength)

	// now mining from the initHead once
	nextHead := mineAndGetNextBlock(initHead) // fork shorter
	bcf.AddBlock(nextHead.Previous.ToBlock(0))
	assert.Equal(t, 1, bcf.ChainLength)

	// now mining from the initHead twice
	nextNextHead := mineAndGetNextBlock(nextHead) //fork longer
	bcf.AddBlock(nextNextHead.Previous.ToBlock(0))
	assert.Equal(t, 2, bcf.ChainLength)

	assert.Equal(t, 1, len(block.Transactions))
}

func TestBlockChainForkWithoutPrevious(t *testing.T) {
	_, tx0 := newTx()
	bcf := blockchain.NewBCF()
	bcf.AddTx(tx0)
	initHead := bcf.Head

	// now mining normally twice
	go bcf.MiningRoutine()
	bcf.MineChan.Get()
	_, tx1 := newTx()
	bcf.AddTx(tx1)
	block := bcf.MineChan.Get()
	assert.True(t, block.Transactions[0].Equals(tx1))
	assert.Equal(t, 2, bcf.ChainLength)

	// now mining from the initHead multiple times (creating a new fork)
	nextHead0 := mineAndGetNextBlock(initHead)
	_, txFork0 := newTx()
	nextHead0.AddTxIfValid(txFork0)
	nextHead1 := mineAndGetNextBlock(nextHead0)
	nextHead1.AddTxIfValid(tx1) // adding a tx already in the fork
	nextHead2 := mineAndGetNextBlock(nextHead1)

	bcf.AddBlock(nextHead2.Previous.ToBlock(0)) // adding the 3rd block before its parent
	assert.Equal(t, 2, bcf.ChainLength)

	bcf.AddBlock(nextHead1.Previous.ToBlock(0)) // adding the 2nd block before its parent

	assert.Equal(t, 2, bcf.ChainLength)

	bcf.AddBlock(nextHead0.Previous.ToBlock(0)) // adding the 1st block (the parent)
	assert.Equal(t, 3, bcf.ChainLength)

	assert.Equal(t, 3, len(bcf.Head.Filenames))
	assert.Equal(t, 3, len(bcf.Head.Hashes))
}
