package blockchain

import (
	"Peerster/messages"
	"sync"
)

/*Blockchain @TODO*/
type Blockchain struct {
	root         *NodeBlock            // The blockchain's root block (empty, no transactions contained)
	head         *NodeBlock            // The blockchain's current head
	transactions map[string]string     // A filename-to-metahash mapping
	blocks       map[string]*NodeBlock // A block's PrevHash-to-
	mux          sync.Mutex            // Mutex to manipulate the structure from different threads
}

/*NewBlockchain @TODO*/
func NewBlockchain() *Blockchain {
	var blockchain Blockchain

	blockchain.root = NewNodeBlock(nil, createRootBlock(), 0)
	blockchain.head = blockchain.root

	blockchain.transactions = make(map[string]string)
	blockchain.blocks = make(map[string]*NodeBlock)

	return &blockchain
}

/*createRootBlock @TODO*/
func createRootBlock() *messages.Block {
	var hash, nounce [32]byte
	return &messages.Block{
		PrevHash:     hash,
		Nonce:        nounce,
		Transactions: nil,
	}
}
