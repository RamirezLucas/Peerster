package entities

import (
	"Peerster/blockchain"
	"Peerster/files"
	"Peerster/peers"
	"fmt"
	"net"
)

// Gossiper - Represents a gossiper
type Gossiper struct {
	Args          *CLArgsGossiper // CL arguments for the Gossiper (RO)
	ClientChannel *net.UDPConn    // UDP channel to communicate with the client (Shared, thread-safe)
	GossipChannel *net.UDPConn    // UDP channel to communicate with the network (Shared, thread-safe)

	/* Rumors and private messages */
	NameIndex *peers.NameIndex               // A dictionnary between peer names and received messages (Shared, thread-safe)
	PeerIndex *peers.PeerIndex               // A dictionnary between <ip:port> and peer addresses (Shared, thread-safe)
	Router    *peers.RoutingTable            // A routing table associating names with next hop address (Shared, thread-safe)
	Timeouts  *peers.StatusResponseForwarder // Timeouts for RumorMessage's answer (Shared, thread-safe)

	/* File transfer */
	FileIndex       *files.FileIndex       // A file index containing all indexed files (Shared, thread-safe)
	TODataRequest   *files.TODataRequest   // Timeouts for DataReplies (Shared, thread-safe)
	SReqTotalMatch  *files.SReqTotalMatch  // Keeps track of how many total matches were received for each SeachRequest (Shared, thread-safe)
	TOSearchRequest *files.TOSearchRequest // Timeouts for received SearchRequest's (Shared, thread-safe)

	/* Blockchain */
	Blockchain *blockchain.BCF // A blockchain for filename-to-metahash claiming (Shared, thread-safe)
}

// CLArgsGossiper - Command line arguments for the gossiper
type CLArgsGossiper struct {
	ClientAddr string   // IP/Port on which the client talks
	GossipAddr string   // IP/Port on which to listen to other gossips
	Name       string   // Name of that gossiper
	ServerPort string   // Port to launch the server on
	SimpleMode bool     // Indicates whether the gossiper operates in simple broadcast mode
	RTimer     uint     // Timer for RouteRumor messages
	Peers      []string // Original list of peers
}

// NewGossiper - Creates a new instance of Gossiper
func NewGossiper(args *CLArgsGossiper) *Gossiper {
	var gossip Gossiper
	gossip.Args = args

	/* Rumors and private messages */
	gossip.NameIndex = peers.NewNameIndex()
	gossip.PeerIndex = peers.NewPeerIndex()
	gossip.Router = peers.NewRoutingTable()
	gossip.Timeouts = peers.NewStatusResponseForwarder()

	// Copy all the peers from the CLArgs to the PeerIndex
	for _, peer := range args.Peers {
		gossip.PeerIndex.AddPeerIfAbsent(peers.StringToUDPAddress(peer))
	}

	/* File transfer */
	gossip.FileIndex = files.NewFileIndex()
	gossip.TODataRequest = files.NewTODataRequest()
	gossip.SReqTotalMatch = files.NewSReqTotalMatch()
	gossip.TOSearchRequest = files.NewTOSearchRequest()

	/* Blockchain */
	gossip.Blockchain = blockchain.NewBCF()

	return &gossip
}

// GossiperToString - Returns the textual representation of a Gossiper
func (gossip *Gossiper) GossiperToString() string {
	return fmt.Sprintf("ClienAddr: %s\nGossipAddr: %s\nName: %s\nSimpleMode: %v\n",
		gossip.Args.ClientAddr, gossip.Args.GossipAddr, gossip.Args.Name, gossip.Args.SimpleMode)
}
