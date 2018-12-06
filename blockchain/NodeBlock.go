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
