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

func TestBlockChainGetBlock(t *testing.T) {
	_, tx := newTx()
	bcf := blockchain.NewBCF()
	go bcf.MiningRoutine()
	bcf.AddTx(tx)
	block := bcf.MineChan.Get()

	assert.True(t, block.Hash == bcf.GetBlock(block.Hash).Block.Hash())
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
	go bcf.MiningRoutine()

	// now mining normally twice
	// 1st block
	bcf.AddTx(tx0)
	bcf.MineChan.Get()

	// 2nd block
	_, tx1 := newTx()
	bcf.AddTx(tx1)
	block := bcf.MineChan.Get()

	assert.True(t, block.Transactions[0].Equals(tx1), "latest block holds the latest transaction")
	assert.Equal(t, 2, bcf.ChainLength, "chain has 2 blocks")

	// this simulates a user having his own blockchain
	bcf2 := blockchain.NewBCF()
	go bcf2.MiningRoutine()
	_, txFork0 := newTx()
	bcf2.AddTx(txFork0)           // adding a new transaction (1st)
	block0 := bcf2.MineChan.Get() // 1st block
	_, txFork1 := newTx()
	bcf2.AddTx(txFork1) // adding a new transaction (2nd)
	block1 := bcf2.MineChan.Get()
	bcf2.AddTx(tx1) // adding a tx that the other user also has
	block2 := bcf2.MineChan.Get()

	assert.Equal(t, 3, bcf2.ChainLength, "this user chain has 3 blocks (one more than the other)")

	// now we start receiving the blocks that the other user was mining (but in reverse order to simulate asking the missing ones)
	block2_ := bcf2.Head.Previous
	assert.True(t, block2.Hash == block2_.Hash)
	block2Added := bcf.AddBlock(block2.ToBlock(0)) // adding the 3rd block before its parent
	//assert.Equal(t, 1, len(pBlocks0), "1 pending block (the next (2nd) one)")
	assert.False(t, block2Added, "block2 is pending")
	assert.Equal(t, 2, bcf.ChainLength, "it is a fork shorter")

	block1_ := block2.Previous
	assert.True(t, block1.Hash == block1_.Hash)
	block1Added := bcf.AddBlock(block1.ToBlock(0)) // adding the 2nd block before its parent
	//assert.Equal(t, 1, len(pBlocks1), "1 pending blocks (the next (1st) one)")
	assert.False(t, block1Added, "block1 is pending")
	assert.Equal(t, 2, bcf.ChainLength, "it is a fork shorter")

	block0_ := block1.Previous
	assert.True(t, block0.Hash == block0_.Hash)
	block0Added := bcf.AddBlock(block0.ToBlock(0)) // adding the 1st block (the parent)
	//assert.Equal(t, 0, len(pBlocks2), "no more pending blocks (all are added)")
	assert.True(t, block0Added, "block0 is added and the two others at the same time")
	assert.Equal(t, 3, bcf.ChainLength, "now it is a fork longer!")

	assert.Equal(t, 4, len(bcf.Head.Filenames), "we have the 4 transactions (tx0, tx1, txFork0 and txFork1")
	assert.Equal(t, 4, len(bcf.Head.Hashes), "we have the 4 transactions (tx0, tx1, txFork0 and txFork1")
}
