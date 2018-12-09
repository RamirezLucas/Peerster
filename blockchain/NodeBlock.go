package blockchain

import "Peerster/messages"

/*NodeBlock @TODO*/
type NodeBlock struct {
	prev   *NodeBlock
	next   []*NodeBlock
	block  *messages.Block
	length uint64
}

/*NewNodeBlock @TODO*/
func NewNodeBlock(prev *NodeBlock, block *messages.Block, length uint64) *NodeBlock {
	var node NodeBlock
	node.prev = prev
	node.next = nil
	node.block = block
	node.length = length
	return &node
}

/*createRootBlock creates an empty "root" `Block` meant to be the first block of every
instantiated `Blockchain`. This `Block` has a `PrevHash` of 0 and an empty list of
transactions.*/
func createRootBlock() *messages.Block {
	var hash, nonce [32]byte
	return &messages.Block{
		PrevHash:     hash,
		Nonce:        nonce,
		Transactions: nil,
	}
}
