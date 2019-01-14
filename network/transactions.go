package network

import (
	"Peerster/blockchain"
	"Peerster/entities"
	"Peerster/messages"
	"Peerster/utils"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// TransactionHopLimit is the hop limit for TxPublish
const TransactionHopLimit = 10

// BlockHopLimit is the hop limit for BlockPublish and BlockReply
const BlockHopLimit = 20

// BlockRequestBudget is the default budget assigned to look for a block
const BlockRequestBudget = 20

/* ================ TRANSACTIONS ================ */

/*OnBroadcastTransaction is used to broadcast a `TxPublish` on the network. */
func OnBroadcastTransaction(gossiper *entities.Gossiper, tx *messages.TxPublish) {

	// Create the packet
	pkt := messages.GossipPacket{TxPublish: tx}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Broadcast the packet
	gossiper.PeerIndex.Broadcast(gossiper.GossipChannel, buf, "")
}

/*OnReceiveTransaction is called when a `TxPublish` is received.*/
func OnReceiveTransaction(gossiper *entities.Gossiper, tx *messages.TxPublish, sender *net.UDPAddr) {

	// Check if the transaction is valid and add it to the pending buffer
	if gossiper.Blockchain.AddTx(blockchain.NewTx(tx)) {

		// Check the hop limit
		if tx.HopLimit == 0 {
			return
		}

		// Broadcast to other peers
		tx.HopLimit--
		OnBroadcastTransaction(gossiper, tx)
	}
}

/* ================ BLOCKS ================ */

/*OnBroadcastBlock is used to broadcast a `Block` on the network.*/
func OnBroadcastBlock(gossiper *entities.Gossiper, block *messages.Block) {

	// Create a BlockPublish
	publish := &messages.BlockPublish{
		Block:    block,
		HopLimit: BlockHopLimit,
	}

	// Create the packet
	pkt := messages.GossipPacket{BlockPublish: publish}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Broadcast the packet
	gossiper.PeerIndex.Broadcast(gossiper.GossipChannel, buf, "")
}

/*OnReceiveBlock is called when a `BlockPublish` is received.*/
func OnReceiveBlock(gossiper *entities.Gossiper, block *messages.BlockPublish, sender *net.UDPAddr) {

	// Get missing blocks if any
	if !gossiper.Blockchain.AddBlock(block.Block) {
		return
	}

	var missingBlocks [][32]byte

	for hash, test := range gossiper.Blockchain.MissingBlocks {
		if test {
			var hashN [32]byte
			copy(hashN[:], utils.HexToHash(hash)[0:32])
			missingBlocks = append(missingBlocks, hashN)
		}
	}

	// If we have any missing block
	if len(missingBlocks) > 0 {

		// Create a block request
		request := &messages.BlockRequest{
			Origin:    gossiper.Args.Name,
			BlockHash: missingBlocks,
			Budget:    BlockRequestBudget,
		}

		// Broadcasting the packet to all neighbours with our budget
		gossiper.PeerIndex.BroadcastBlockRequest(gossiper.GossipChannel, request, "")
	}
}

/*OnReceiveBlockRequest is called when we receive a block request from anyone
The goal is to continue broadcasting this request if we don't have all the requested blocks
If we have some blocks, we create and send a DataReply back with all the blocks we know about
Then we continue broadcasting the request, but only with the remaining blocks (the one we don't have)*/
func OnReceiveBlockRequest(gossiper *entities.Gossiper, request *messages.BlockRequest, sender *net.UDPAddr) {

	// Holds the blocks found and matching the request
	var blocksFound []*messages.Block

	// If the block request is empty, it means the sender just connected to the network and is looking for an existing blockchain
	// Therefore we will just send the latest block mined
	if len(request.BlockHash) == 0 && gossiper.Blockchain.Head != nil && gossiper.Blockchain.Head.Previous != nil {

		fmt.Printf("RECEIVED CHAIN REQUEST from %s\n", request.Origin)

		blocksFound = append(blocksFound, gossiper.Blockchain.GetBlock(gossiper.Blockchain.Head.Previous.Hash).Block)

		// Create a block reply
		reply := &messages.BlockReply{
			Destination: request.Origin,
			Block:       blocksFound,
			HopLimit:    BlockHopLimit,
		}

		OnReceiveBlockReply(gossiper, reply, sender)

		// We end the function here
		return
	}

	fmt.Printf("RECEIVED BLOCK REQUEST from %s asking for %d blocks\n", request.Origin, len(request.BlockHash))

	// Holds the hashes of the blocks to broadcast (the ones we don't have)
	var newRequest [][32]byte

	for _, hash := range request.BlockHash {

		blockFound := gossiper.Blockchain.GetBlock(hash)

		// If we have the requested block, we add it to the blocks found so far, else we add the hash to our new request
		if blockFound != nil {
			blocksFound = append(blocksFound, blockFound.Block)
		} else {
			newRequest = append(newRequest, hash)
		}
	}

	// If we found at least one block, we have to create a reply
	if len(blocksFound) > 0 {

		// Create a block reply
		reply := &messages.BlockReply{
			Destination: request.Origin,
			Block:       blocksFound,
			HopLimit:    BlockHopLimit,
		}

		OnReceiveBlockReply(gossiper, reply, sender)
	}

	// Finally if we still have some requests we broadcast them (but we exclude the sender)
	if len(newRequest) > 0 {
		request.BlockHash = newRequest
		gossiper.PeerIndex.BroadcastBlockRequest(gossiper.GossipChannel, request, sender.String())
	}
}

/*OnReceiveBlockReply when receiving a block reply*/
func OnReceiveBlockReply(gossiper *entities.Gossiper, reply *messages.BlockReply, sender *net.UDPAddr) {

	// First we check if we are the destination, and if so, we have to take care of the new blocks
	// Otherwise we have to forward the reply
	if reply.Destination == gossiper.Args.Name {

		fmt.Printf("RECEIVED BLOCK REPLY from %s\n", sender.String())

		for _, block := range reply.Block {
			myBlock := gossiper.Blockchain.GetBlock(block.Hash())

			// If we don't know about this block yet, we have to take care of it
			if myBlock == nil {
				gossiper.Blockchain.AddBlockGen(block, false)
			}
		}

		var missingBlocks [][32]byte

		for hash, test := range gossiper.Blockchain.MissingBlocks {
			if test {
				var hashN [32]byte
				copy(hashN[:], utils.HexToHash(hash)[0:32])
				missingBlocks = append(missingBlocks, hashN)
			}
		}

		// If we have any missing block
		if len(missingBlocks) > 0 {

			// Create a block request
			request := &messages.BlockRequest{
				Origin:    gossiper.Args.Name,
				BlockHash: missingBlocks,
				Budget:    BlockRequestBudget,
			}

			// Broadcasting the packet to all neighbours with our budget
			gossiper.PeerIndex.BroadcastBlockRequest(gossiper.GossipChannel, request, "")
		}

	} else if reply.HopLimit > 0 {

		// Prepare the packet
		reply.HopLimit--
		pkt := messages.GossipPacket{BlockReply: reply}
		buf, err := protobuf.Encode(&pkt)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Setup the target
		target := gossiper.Router.GetTarget(reply.Destination)

		// Send the packet
		fmt.Printf("FORWARDING BLOCK REPLY to %s via %s\n", reply.Destination, target.String())
		gossiper.GossipChannel.WriteToUDP(buf, target)
	} else {
		fmt.Printf("HOPLIMIT for BLOCK REPLY is 0\n")
	}
}
