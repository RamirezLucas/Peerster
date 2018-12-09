package blockchain

import (
	"Peerster/fail"
	"Peerster/files"
	"Peerster/messages"
	"sync"
	"time"
)

// factorSleepingAfterMining is the number by which we multiply the mining time to get
// the waiting time after a block has been mined.
const factorSleepingAfterMining = 2

// factorSleepingGenesis is the amount of time to wait (in seconds) before publishing the genesis block
const sleepingGenesisSec = 5

/*Blockchain */
type Blockchain struct {
	root         *NodeBlock                 // The blockchain's root block (empty, no transactions contained)
	head         *NodeBlock                 // The blockchain's current head
	transactions map[string]string          // A filename-to-metahash mapping
	blocks       map[string]*NodeBlock      // A block's PrevHash-to-NodeBlock mapping
	invalid      map[string]*messages.Block // A set of invalid blocks

	pendingTxs map[string]*messages.File // The set of pending transactions (filename to *File)

	mux sync.Mutex // Mutex to manipulate the structure from different threads
}

/*NewBlockchain */
func NewBlockchain() *Blockchain {
	var blockchain Blockchain

	blockchain.root = NewNodeBlock(nil, createRootBlock(), 0)
	blockchain.head = blockchain.root

	blockchain.transactions = make(map[string]string)
	blockchain.blocks = make(map[string]*NodeBlock)
	blockchain.invalid = make(map[string]*messages.Block)
	blockchain.pendingTxs = make(map[string]*messages.File)

	// Add the "root" block to the list of known blocks
	blockchain.blocks[files.ToHex32(blockchain.root.block.Hash())] = blockchain.root

	return &blockchain
}

/*AddBlock */
func (blockchain *Blockchain) AddBlock(newBlock *messages.Block) bool {

	tmpHash := newBlock.Hash()
	fail.LeveledPrint(1, "Blockchain.AddBlock", "Attempting to add new block with hash %s",
		files.ToHex(tmpHash[:]))

	// Check that the block has valid hash
	if !newBlock.CheckHashValid() {
		return false
	}

	// Grab the mutex
	blockchain.mux.Lock()
	defer blockchain.mux.Unlock()

	// Add the block
	return blockchain.addBlockUnsafe(newBlock)
}

/*AddPendingTransaction */
func (blockchain *Blockchain) AddPendingTransaction(newTX *messages.File) bool {

	fail.LeveledPrint(1, "Blockchain.AddPendingTransaction",
		"Attempting to add pending transaction %s", newTX.Name)

	// Grab the mutex
	blockchain.mux.Lock()
	defer blockchain.mux.Unlock()

	if _, ok := blockchain.pendingTxs[newTX.Name]; ok { // We already have the transaction pending
		return false
	}
	if _, ok := blockchain.transactions[newTX.Name]; ok { // The association is already claimed
		return false
	}

	// Add the pending transaction
	blockchain.pendingTxs[newTX.Name] = newTX

	fail.LeveledPrint(1, "Blockchain.AddPendingTransaction",
		"Added pending transaction %s", newTX.Name)

	return true
}

/*MineNewBlock */
func (blockchain *Blockchain) MineNewBlock(isGenesis bool) *messages.Block {

	fail.LeveledPrint(1, "Blockchain.MineNewBlock", "Entering MineNewBlock")

	// Grab the mutex
	blockchain.mux.Lock()

	// Get list of valid pending transactions
	newTx := make([]messages.TxPublish, 0)
	for filename, file := range blockchain.pendingTxs {
		if _, ok := blockchain.transactions[filename]; !ok {
			fail.LeveledPrint(1, "Blockchain.MineNewBlock", "%s not in transactions", filename)

			// Append the transaction if it isn't in the blockchain
			newTx = append(newTx, messages.TxPublish{
				File:     *file,
				HopLimit: 0,
			})
		} else {
			fail.LeveledPrint(1, "Blockchain.MineNewBlock", "%s in transactions", filename)
			// Remove the transaction from the pending buffer
			delete(blockchain.pendingTxs, filename)
		}
	}

	// If there are no new transactions abort
	if !isGenesis && (newTx == nil || len(newTx) == 0) {
		// Release the mutex
		blockchain.mux.Unlock()
		return nil
	}

	// Generate block
	var hash, nonce [32]byte
	if !isGenesis {
		hash = blockchain.head.block.Hash()
	}
	newBlock := &messages.Block{
		PrevHash:     hash,
		Nonce:        nonce,
		Transactions: newTx,
	}

	// Release the mutex
	blockchain.mux.Unlock()

	/* **************** Change the nonce until the hash is valid **************** */
	fail.LeveledPrint(1, "Blockchain.MineNewBlock", "Mining new block for %d transactions", len(newTx))
	start := time.Now()
	for !newBlock.CheckHashValid() {
		newBlock.ChangeNonceRandomly()
	}
	elapsed := time.Since(start)
	fail.LeveledPrint(1, "Blockchain.MineNewBlock", "Done mining")
	/* **************** Change the nonce until the hash is valid **************** */

	// Grab the mutex
	blockchain.mux.Lock()

	// Check that the blockchain head is still the same
	if !isGenesis && (files.ToHex32(newBlock.PrevHash) != files.ToHex32(blockchain.head.block.Hash())) {
		blockchain.mux.Unlock()
		return blockchain.MineNewBlock(false)
	}

	// Check that the new transactions are still valid
	for _, tx := range newTx {
		if _, ok := blockchain.transactions[tx.File.Name]; ok {
			// One of the pending transaction is now in the blockchain, abort
			blockchain.mux.Unlock()
			return blockchain.MineNewBlock(false)
		}
	}

	fail.LeveledPrint(0, "", "FOUND-BLOCK %s", files.ToHex32(newBlock.Hash()))

	// Append the block to the blockchain
	blockchain.addBlockUnsafe(newBlock)
	blockchain.mux.Unlock()

	// Wait some time before publishing
	if isGenesis {
		time.Sleep(5 * time.Second)
	} else {
		time.Sleep(2 * elapsed)
	}
	return newBlock
}

/*AddBlockUnsafe */
func (blockchain *Blockchain) addBlockUnsafe(newBlock *messages.Block) bool {

	// Determine if this is a genesis block
	hashCompare := newBlock.PrevHash
	if newBlock.IsGenesis() {
		hashCompare = blockchain.root.block.Hash()
	}

	if prevBlock, ok := blockchain.blocks[files.ToHex(hashCompare[:])]; ok { // The previous block is known

		// Create a new NodeBlock and append it to the previous node's list of next nodes
		newNode := NewNodeBlock(prevBlock, newBlock, prevBlock.length+1)
		blockHash := files.ToHex32(newBlock.Hash())
		prevBlock.next = append(prevBlock.next, newNode)
		blockchain.blocks[blockHash] = newNode

		if prevBlock == blockchain.head { // The new block is the new head
			blockchain.head = newNode
			blockchain.addTransactions(newNode)

			// Print to the console
			fail.LeveledPrint(0, "", blockchain.printChain(newNode))

		} else if blockchain.head.length < newNode.length { // We have a new longest chain

			if cntRewind := blockchain.fork(newNode); cntRewind != 0 {
				// Print to the console
				fail.LeveledPrint(0, "", "FORK-LONGER rewind %d blocks", cntRewind)
				fail.LeveledPrint(0, "", blockchain.printChain(newNode))
			}
		}

		if len(prevBlock.next) != 1 { // We just discovered a new fork
			fail.LeveledPrint(0, "", "FORK-SHORTER %s", files.ToHex32(prevBlock.block.Hash()))
		}

		// See if the block we just added is the predecessor of some previously received one
		if invalidBlock, ok := blockchain.invalid[blockHash]; ok {
			blockchain.addBlockUnsafe(invalidBlock)
		}

		return true
	}

	// The previous block is not known
	fail.LeveledPrint(1, "Blockchain.addBlockUnsafe", "New invalid block with prevhash %s", files.ToHex32(newBlock.PrevHash))
	blockchain.invalid[files.ToHex32(newBlock.PrevHash)] = newBlock
	return false
}

/*addTransactions */
func (blockchain *Blockchain) addTransactions(node *NodeBlock) {
	for _, tx := range node.block.Transactions {
		if _, ok := blockchain.transactions[tx.File.Name]; !ok {
			blockchain.transactions[tx.File.Name] = files.ToHex(tx.File.MetafileHash[:])
		} else {
			fail.CustomPanic("Blockchain.fork", "Adding existing transaction %s -> %s",
				tx.File.Name, files.ToHex(tx.File.MetafileHash[:]))
		}
	}
}

/*fork */
func (blockchain *Blockchain) fork(newHead *NodeBlock) uint64 {

	// Chexk chains length
	if newHead.length != blockchain.head.length+1 {
		fail.CustomPanic("Blockchain.fork",
			"New chain is not exactly one block longer than current longest chain.\n"+
				"\tCurrent chain length: %d\n\tNew chain length: %d",
			blockchain.head.length, newHead.length)
		return 0
	}

	// Maintain some status
	newBlocksFork := make([]*NodeBlock, 0)
	deleteBuffer := make(map[string]*messages.File)
	addBuffer := make(map[string]string)
	cntRewind := uint64(0)

	// Find the most recent common block between the current head and the new head
	newBlocksFork = append(newBlocksFork, newHead)
	newPath := newHead.prev
	oldPath := blockchain.head

	// Iterate backwards to the root
	for i := newPath.length; i >= 0; i-- {
		if oldPath == newPath { // Found a common block
			break
		} else {
			// Buffer the blocks transactions on the current chain
			newBlocksFork = append(newBlocksFork, newPath)
			for _, tx := range oldPath.block.Transactions {
				deleteBuffer[tx.File.Name] = &tx.File
			}

			// Go backward
			oldPath = oldPath.prev
			newPath = newPath.prev
			cntRewind++
		}
	}

	// Add all transactions from the new longest chain to a buffer
	nbNewNodes := len(newBlocksFork)
	validChain := true
	i := nbNewNodes - 1
	for ; i >= 0; i-- { // For each block on the new path (from latest to most recent)
		for _, tx := range newBlocksFork[i].block.Transactions { // For each transaction
			// Double mapping on the forked chain
			if _, ok := addBuffer[tx.File.Name]; ok {
				validChain = false
				break
			}
			// Double mapping (one mapping on the forked chain, another one on the old mappings)
			if _, ok1 := blockchain.blocks[tx.File.Name]; ok1 {
				if _, ok2 := deleteBuffer[tx.File.Name]; !ok2 {
					validChain = false
					break
				}
			}
			addBuffer[tx.File.Name] = files.ToHex(tx.File.MetafileHash[:])
		}
	}

	if validChain { // The new chain is valid
		// Update the blockchain's head
		blockchain.head = newHead

		// Apply changes
		for filename, file := range deleteBuffer {
			// Delete everything from the delete buffer
			delete(blockchain.transactions, filename)

			// Transactions that are only deleted should go in the pending transaction buffer
			if _, ok := addBuffer[filename]; !ok {
				blockchain.pendingTxs[filename] = file
			}
		}
		for filename, metahash := range addBuffer {
			// Add everything from the add buffer
			blockchain.transactions[filename] = metahash
		}
		return cntRewind
	}

	// The new changes are invalid, abort fork and cut out the chain
	cutOutBlock := newBlocksFork[i].prev
	for indexFaultyBlock, nextBlock := range cutOutBlock.next {
		if nextBlock == newBlocksFork[i] { // Find faulty block
			// Delete node in next list
			if len(cutOutBlock.next) == 1 {
				cutOutBlock.next = nil
			} else {
				cutOutBlock.next[indexFaultyBlock] = cutOutBlock.next[len(cutOutBlock.next)-1]
				cutOutBlock.next = cutOutBlock.next[:len(cutOutBlock.next)-1]
			}
			break
		}
	}

	// Nothing was done
	return 0
}

/*printChain */
func (blockchain *Blockchain) printChain(head *NodeBlock) string {
	str := "CHAIN "
	for node := head; node != blockchain.root; node = node.prev {
		str += node.block.ToString() + " "
	}
	return str[:len(str)-1]
}
