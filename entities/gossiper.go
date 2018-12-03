package entities

import (
	"Peerster/files"
	"Peerster/peers"
	"fmt"
	"net"
)

// Gossiper - Represents a gossiper
type Gossiper struct {
	Args          *CLArgsGossiper                // CL arguments for the Gossiper (RO)
	ClientChannel *net.UDPConn                   // UDP channel to communicate with the client (Shared, thread-safe)
	GossipChannel *net.UDPConn                   // UDP channel to communicate with the network (Shared, thread-safe)
	NameIndex     *peers.NameIndex               // A dictionnary between peer names and received messages (Shared, thread-safe)
	PeerIndex     *peers.PeerIndex               // A dictionnary between <ip:port> and peer addresses (Shared, thread-safe)
	Router        *peers.RoutingTable            // A routing table associating names with next hop address (Shared, thread-safe)
	Timeouts      *peers.StatusResponseForwarder // Timeouts for RumorMessage's answer (Shared, thread-safe)

	/* File transfer */
	FileIndex       *files.FileIndex             // A file index containing all indexed files (Shared, thread-safe)
	DataTimeouts    *files.DataResponseForwarder // Timeouts for DataReplies (Shared, thread-safe)
	SReqTotalMatch  *files.SReqTotalMatch        // Keeps track of how many total matches were received for each SeachRequest (Shared, thread-safe)
	TOSearchRequest *files.TOSearchRequest       // Timeouts for received SearchRequest's (Shared, thread-safe)
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
	gossip.NameIndex = peers.NewNameIndex()
	gossip.PeerIndex = peers.NewPeerIndex()
	gossip.Router = peers.NewRoutingTable()
	gossip.Timeouts = peers.NewStatusResponseForwarder()

	/* File transfer */
	gossip.FileIndex = files.NewFileIndex()
	gossip.DataTimeouts = files.NewDataResponseForwarder()
	gossip.SReqTotalMatch = files.NewSReqTotalMatch()
	gossip.TOSearchRequest = files.NewTOSearchRequest()

	// Copy all the peers from the CLArgs to the PeerIndex
	for _, peer := range args.Peers {
		gossip.PeerIndex.AddPeerIfAbsent(peers.StringToUDPAddress(peer))
	}

	return &gossip
}

// GossiperToString - Returns the textual representation of a Gossiper
func (gossip *Gossiper) GossiperToString() string {
	return fmt.Sprintf("ClienAddr: %s\nGossipAddr: %s\nName: %s\nSimpleMode: %v\n",
		gossip.Args.ClientAddr, gossip.Args.GossipAddr, gossip.Args.Name, gossip.Args.SimpleMode)
}
