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
	bcf := blockchain.NewBCF()
	go bcf.MiningRoutine()
	bcf.AddTx(tx)
	block := bcf.MineChan.Get()

	assert.True(t, block.Transactions[0].Equals(tx))
	assert.Equal(t, 1, bcf.ChainLength)
}

func TestBlockChainForkShorterAndLonger(t *testing.T) {
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
	bcf := blockchain.NewBCF()
	bcf.AddTx(tx)
	initHead := bcf.Head

	// now mining normally
	go bcf.MiningRoutine()
	block := bcf.MineChan.Get()
	assert.True(t, block.Transactions[0].Equals(tx))
	assert.Equal(t, 1, bcf.ChainLength)

	// now mining from the initHead once
	nextHead := mineAndGetNextBlock(initHead)

	// now mining from the initHead twice
	nextNextHead := mineAndGetNextBlock(nextHead)

	bcf.AddBlock(nextNextHead.Previous.ToBlock(0)) // adding the 2nd block before its parent
	assert.Equal(t, 1, bcf.ChainLength)

	bcf.AddBlock(nextHead.Previous.ToBlock(0)) // adding the 1st block (the parent)
	assert.Equal(t, 2, bcf.ChainLength)

	assert.Equal(t, 1, len(block.Transactions))
}
