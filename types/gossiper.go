package types

import (
	"fmt"
	"net"
)

// BufSize - Size of the UDP buffer
const BufSize = 2048

// TimeoutSec - Length of generic timeout
const TimeoutSec = 1

// Gossiper - Represents a gossiper
type Gossiper struct {
	Args          *CLArgsGossiper          // CL arguments for the Gossiper (RO)
	ClientChannel *net.UDPConn             // UDP channel to communicate with the client (Shared, thread-safe)
	GossipChannel *net.UDPConn             // UDP channel to communicate with the network (Shared, thread-safe)
	NameIndex     *NameIndex               // A dictionnary between peer names and received messages (Shared, thread-safe)
	PeerIndex     *PeerIndex               // A dictionnary between <ip:port> and peer addresses (Shared, thread-safe)
	Router        *RoutingTable            // A routing table associating names with next hop address (Shared, thread-safe)
	FileIndex     *FileIndex               // A file index containing all indexed files (Shared, thread-safe)
	Timeouts      *StatusResponseForwarder // Timeouts for RumorMessage's answer (Shared, thread-safe)
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
	gossip.NameIndex = NewNameIndex()
	gossip.PeerIndex = NewPeerIndex()
	gossip.Router = NewRoutingTable()
	gossip.FileIndex = NewFileIndex()
	gossip.Timeouts = NewStatusResponseForwarder()

	// Copy all the peers from the CLArgs to the PeerIndex
	for _, peer := range args.Peers {
		gossip.PeerIndex.AddPeerIfAbsent(StringToUDPAddress(peer))
	}

	return &gossip
}

// GossiperToString - Returns the textual representation of a Gossiper
func (gossip *Gossiper) GossiperToString() string {
	return fmt.Sprintf("ClienAddr: %s\nGossipAddr: %s\nName: %s\nSimpleMode: %v\n",
		gossip.Args.ClientAddr, gossip.Args.GossipAddr, gossip.Args.Name, gossip.Args.SimpleMode)
}
