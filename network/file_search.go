package network

import (
	"Peerster/entities"
	"Peerster/fail"
	"Peerster/files"
	"Peerster/messages"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/dedis/protobuf"
)

const (
	// InitialBudget represents the initial budget when a new SearchRequest is created
	InitialBudget = uint64(2)
	// MaximumBudget represents the maximum budget any SearchRequest can reach
	MaximumBudget = uint64(32)
	// SearchRepeatIntervalSec represents the interval of time between two consecutive SearchRequest's
	SearchRepeatIntervalSec = 1
	// ThresholdTotalMatches represents the number of total matches required to stop a SearchRequest
	ThresholdTotalMatches = 1
)

/* ================ SEARCH REQUEST ================ */

// OnInitiateFileSearch initiates a file search on the network.
func OnInitiateFileSearch(gossiper *entities.Gossiper, defaultBudget uint64, keywords []string) {

	// Set budget
	initBudget := InitialBudget
	budgetMultiplication := (defaultBudget == 0)
	if budgetMultiplication {
		initBudget = defaultBudget
	}

	// Create a SearchRequest
	search := &messages.SearchRequest{
		Origin:   gossiper.Args.Name,
		Budget:   initBudget,
		Keywords: keywords,
	}

	// Register the SearchRequest in the gossiper to count the number of total matches
	gossiper.SReqTotalMatch.AddSearchRequest(search)

	for search.Budget <= MaximumBudget {
		// Pick a random neighbor and send it the SearchRequest
		if neighbor := gossiper.PeerIndex.GetRandomPeer(nil); neighbor != nil {
			OnSendSearchRequest(gossiper.GossipChannel, search, neighbor)

			// Wait some time and check the number of total matches
			time.Sleep(SearchRepeatIntervalSec * time.Second)
			if gossiper.SReqTotalMatch.CheckThresholdAndDelete(search, ThresholdTotalMatches) {
				fail.LeveledPrint(0, "", "SEARCH FINISHED")
				return
			}

			// Double the budget and resend
			search.Budget *= 2
		} else {
			break
		}

		// Stop after the first request if the budget was specified by the user
		if !budgetMultiplication {
			break
		}
	}

	// Delete the handler
	gossiper.SReqTotalMatch.DeleteSearchRequest(search)
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
		Results:     gossiper.FileIndex.HandleSearchRequest(search),
	}

	// Reply to sender
	if target := gossiper.Router.GetTarget(search.Origin); target != nil {
		OnSendSearchReply(gossiper.GossipChannel, reply, target)
	}

	// Set up timeout for particular request and delete it after 0.5 second
	if hash := gossiper.TOSearchRequest.AddSearchRequest(search); hash != "" {
		time.Sleep(500 * time.Millisecond)
		gossiper.TOSearchRequest.RemoveSearchRequest(hash)
	}
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
		for _, result := range reply.Results { // For each result

			// Create sorted list of chunk indices
			sort.Slice(result.ChunkMap, func(i, j int) bool { return result.ChunkMap[i] < result.ChunkMap[j] })
			strChunkMap := ""
			for _, index := range result.ChunkMap {
				strChunkMap += strconv.FormatUint(index, 10) + ","
			}
			if len(strChunkMap) > 0 {
				strChunkMap = strChunkMap[:len(strChunkMap)-1]
			}

			// Print to the console
			fail.LeveledPrint(0, "", "FOUND match %s at %s metafile=%s chunks=%s",
				result.Filename, reply.Origin, files.ToHex(result.MetafileHash[:]), strChunkMap)

			// Handle the SearchResult
			if gossiper.FileIndex.HandleSearchResult(result, reply.Origin) {
				// We just had a total match
				gossiper.SReqTotalMatch.UpdateIndexOnTotalMatch(result.Filename)
			}
		}

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
