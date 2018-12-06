package blockchain

import (
	"Peerster/fail"
	"Peerster/files"
	"Peerster/messages"
	"sync"
)

// Number of zero bytes that every Block hash must start with
const nbBytesZero = 4

/*Blockchain @TODO*/
type Blockchain struct {
	root         *NodeBlock            // The blockchain's root block (empty, no transactions contained)
	head         *NodeBlock            // The blockchain's current head
	transactions map[string]string     // A filename-to-metahash mapping
	blocks       map[string]*NodeBlock // A block's PrevHash-to-NodeBlock mapping

	pendingTxs map[string]*messages.File // The set of pending transactions (filename to *File)

	mux sync.Mutex // Mutex to manipulate the structure from different threads
}

/*NewBlockchain @TODO*/
func NewBlockchain() *Blockchain {
	var blockchain Blockchain

	blockchain.root = NewNodeBlock(nil, createRootBlock(), 0)
	blockchain.head = blockchain.root

	blockchain.transactions = make(map[string]string)
	blockchain.blocks = make(map[string]*NodeBlock)
	blockchain.pendingTxs = nil

	// Add the "root" block to the list of known blocks
	blockchain.blocks[files.ToHex32(blockchain.root.block.PrevHash)] = blockchain.root

	return &blockchain
}

/*AddBlock @TODO*/
func (blockchain *Blockchain) AddBlock(newBlock *messages.Block) bool {

	// Check that the block's hash is valid
	newBlockHash := newBlock.Hash()
	for i := 0; i < nbBytesZero; i++ {
		if newBlockHash[i] != 0 {
			return false // Invalid hash
		}
	}

	// Grab the mutex
	blockchain.mux.Lock()
	defer blockchain.mux.Unlock()

	// Check that the previous block is known
	if prevBlock, ok := blockchain.blocks[files.ToHex32(newBlock.PrevHash)]; ok {

		// Create a new NodeBlock and append it to the previous node's list of next nodes
		newNode := NewNodeBlock(prevBlock, newBlock, prevBlock.length+1)
		prevBlock.next = append(prevBlock.next, newNode)
		blockchain.blocks[files.ToHex32(newBlockHash)] = newNode

		if prevBlock == blockchain.head { // The new block is the new head
			blockchain.head = newNode
			blockchain.addTransactions(newNode)
		} else if blockchain.head.length < newNode.length { // We have a new longest chain
			blockchain.fork(newNode)
		}

		return true
	}

	// The previous block is unknown: do nothing
	return false
}

/*AddPendingTransaction @TODO*/
func (blockchain *Blockchain) AddPendingTransaction(newTX *messages.File) bool {

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
	return true
}

/*addTransactions @TODO*/
func (blockchain *Blockchain) addTransactions(node *NodeBlock) {
	for _, tx := range node.block.Transactions {
		if _, ok := blockchain.transactions[tx.File.Name]; !ok {
			files.ToHex(tx.File.MetafileHash[:])
		} else {
			// @TODO: what to do here
		}
	}
}

/*deleteTransactions @TODO*/
func (blockchain *Blockchain) deleteTransactions(node *NodeBlock) {
	for _, tx := range node.block.Transactions {
		if _, ok := blockchain.transactions[tx.File.Name]; ok {
			delete(blockchain.transactions, tx.File.Name)
		} else {
			// @TODO: what to do here (fixing add should solve this too)
		}
	}
}

/*fork @TODO*/
func (blockchain *Blockchain) fork(newHead *NodeBlock) {

	// Chexk chains length
	if newHead.length != blockchain.head.length+1 {
		fail.CustomPanic("Blockchain.fork",
			"New chain is not exactly one block longer than current longest chain.\n"+
				"\tCurrent chain length: %d\n\tNew chain length: %d",
			blockchain.head.length, newHead.length)
	}

	// Find the most recent common block between the current head and the new head
	newPath := newHead.prev
	oldPath := blockchain.head

	// Iterate backwards to the root
	for i := newPath.length; i >= 0; i-- {
		if oldPath == newPath { // Found a common block
			break
		} else {
			// Delete the block's transactions on the current chain
			blockchain.deleteTransactions(oldPath)

			// Go backward
			oldPath = oldPath.prev
			newPath = newPath.prev
		}
	}

	// Add all transactions from the new longest chain
	for node := newHead; node != oldPath; node = node.prev {
		blockchain.addTransactions(node)
	}

}

/*createRootBlock creates an empty "root" `Block` meant to be the first block of every
instantiated `Blockchain`. This `Block` has a `PrevHash` of 0 and an empty list of
transactions.*/
func createRootBlock() *messages.Block {
	var hash, nounce [32]byte
	return &messages.Block{
		PrevHash:     hash,
		Nonce:        nounce,
		Transactions: nil,
	}
}
