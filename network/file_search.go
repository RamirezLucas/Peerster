package network

import (
	"Peerster/entities"
	"Peerster/messages"
	"net"

	"github.com/dedis/protobuf"
)

/* ================ SEARCH REQUEST ================ */

// OnInitiateFileSearch initiates a file search on the network.
func OnInitiateFileSearch() {
	//@TODO
}

// OnSendSearchRequest sends a SearchRequest on the network.
func OnSendSearchRequest(channel *net.UDPConn, search *messages.SearchRequest, target *net.UDPAddr) {

	// Create the packet
	pkt := messages.GossipPacket{SearchRequest: search}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send the packet
	channel.WriteToUDP(buf, target)

}

// OnReceiveSearchRequest handles an incoming SearchRequest.
func OnReceiveSearchRequest(gossiper *entities.Gossiper, search *messages.SearchRequest, sender *net.UDPAddr) {

	// Update the routing table
	if search.Origin != gossiper.Args.Name {
		gossiper.Router.AddContactIfAbsent(search.Origin, sender)
	}

	// Spread the request to other peers
	search.Budget--
	if search.Budget > 0 {
		neighborsSpread := gossiper.PeerIndex.GetRandomNeighbors(int(search.Budget), sender)
		nbNeighbors := uint64(len(neighborsSpread)) // 0 <= * <= search.Budget

		baseBudget := uint64(1)
		excessBudget := uint64(0)

		// The budget is bigger than the number of neighbors
		if nbNeighbors < search.Budget {
			baseBudget = search.Budget / nbNeighbors
			excessBudget = search.Budget % nbNeighbors
		}

		// Send to all selected neighbors
		for i, target := range neighborsSpread {
			// Set correct budget
			search.Budget = baseBudget
			if uint64(i) < excessBudget {
				search.Budget++
			}
			// Spread SearchRequest
			OnSendSearchRequest(gossiper.GossipChannel, search, target)
		}

	}

	// Process the request locally
	reply := &messages.SearchReply{
		Origin:      gossiper.Args.Name,
		Destination: search.Origin,
		HopLimit:    10,
		Results:     gossiper.FileIndex.GetMatchingFiles(search.Keywords),
	}

	// Reply to sender
	if target := gossiper.Router.GetTarget(search.Origin); target != nil {
		OnSendSearchReply(gossiper.GossipChannel, reply, target)
	}

	// @TODO: set up timeout for particular request
}

/* ================ SEARCH REPLY ================ */

// OnSendSearchReply sends a SearchReply on the network.
func OnSendSearchReply(channel *net.UDPConn, reply *messages.SearchReply, target *net.UDPAddr) {

	// Create the packet
	pkt := messages.GossipPacket{SearchReply: reply}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send the packet
	channel.WriteToUDP(buf, target)

}

// OnReceiveSearchReply handles an incoming SearchReply.
func OnReceiveSearchReply(gossiper *entities.Gossiper, reply *messages.SearchReply, sender *net.UDPAddr) {

	// Update the routing table
	if reply.Origin != gossiper.Args.Name {
		gossiper.Router.AddContactIfAbsent(reply.Origin, sender)
	}

	if gossiper.Args.Name == reply.Destination { // Message is for me
		// TODO: process it !
		// TODO: process it !
		// TODO: process it !
		// TODO: process it !

	} else { // Message is for someone else

		// Decrement hop limit
		reply.HopLimit--

		// Relay SearchReply if hop-limit not exhausted
		if reply.HopLimit != 0 {

			// Pick the target and forward
			if target := gossiper.Router.GetTarget(reply.Destination); target != nil {
				OnSendSearchReply(gossiper.GossipChannel, reply, target)
			}
		}
	}

}
