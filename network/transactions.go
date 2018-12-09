package network

import (
	"Peerster/entities"
	"Peerster/messages"
	"net"

	"github.com/dedis/protobuf"
)

// TransactionHopLimit is the hop limit for TxPublish
const TransactionHopLimit = 10

// BlockHopLimit is the hop limit for BlochPublish
const BlockHopLimit = 20

/* ================ TRANSACTIONS ================ */

/*OnBroadcastTransaction is used to broadcast a `TxPublish` on the network. */
func OnBroadcastTransaction(gossiper *entities.Gossiper, tx *messages.TxPublish) {

	// Create the packet
	pkt := messages.GossipPacket{TxPublish: tx}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Broadcast the packet
	gossiper.PeerIndex.Broadcast(gossiper.GossipChannel, buf, "")
}

/*OnReceiveTransaction is called when a `TxPublish` is received.*/
func OnReceiveTransaction(gossiper *entities.Gossiper, tx *messages.TxPublish, sender *net.UDPAddr) {

	// Check if the transaction is valid and add it to the pending buffer
	if gossiper.Blockchain.AddPendingTransaction(&tx.File) {

		// Check the hop limit
		if tx.HopLimit == 0 {
			return
		}

		// Broadcast to other peers
		tx.HopLimit--
		OnBroadcastTransaction(gossiper, tx)

		// Start mining
		if newBlock := gossiper.Blockchain.MineNewBlock(false); newBlock != nil {
			OnBroadcastBlock(gossiper, newBlock)
		}
	}

}

/* ================ BLOCKS ================ */

/*OnBroadcastBlock is used to broadcast a `Block` on the network.*/
func OnBroadcastBlock(gossiper *entities.Gossiper, block *messages.Block) {

	// Create a BlockPublish
	publish := &messages.BlockPublish{
		Block:    *block,
		HopLimit: BlockHopLimit,
	}

	// Create the packet
	pkt := messages.GossipPacket{BlockPublish: publish}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Broadcast the packet
	gossiper.PeerIndex.Broadcast(gossiper.GossipChannel, buf, "")
}

/*OnReceiveBlock is called when a `BlockPublish` is received.*/
func OnReceiveBlock(gossiper *entities.Gossiper, block *messages.BlockPublish, sender *net.UDPAddr) {
	gossiper.Blockchain.AddBlock(&block.Block)
}
