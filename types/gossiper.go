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
	ClientAddr    string                   // IP/Port on which the client talks (RO)
	GossipAddr    string                   // IP/Port on which to listen to other gossips (RO)
	Name          string                   // Name of that gossiper (RO)
	ServerPort    string                   // Port to launch the server on (RO)
	SimpleMode    bool                     // Indicates whether the gossiper operates in simple broadcast mode (RO)
	ClientChannel *net.UDPConn             // UDP channel to communicate with the client (Shared, thread-safe)
	GossipChannel *net.UDPConn             // UDP channel to communicate with the network (Shared, thread-safe)
	NameIndex     *NameIndex               // A dictionnary between peer names and received messages (Shared, thread-safe)
	PeerIndex     *PeerIndex               // A dictionnary between <ip:port> and peer addresses (Shared, thread-safe)
	Router        *RoutingTable            // A routing table associating names with next hop address (Shared, thread-safe)
	Timeouts      *StatusResponseForwarder // Timeouts for RumorMessage's answer (Shared, thread-safe)
}

// NewGossiper - Creates a new instance of Gossiper
func NewGossiper() *Gossiper {
	var gossip Gossiper
	gossip.NameIndex = NewNameIndex()
	gossip.PeerIndex = NewPeerIndex()
	gossip.Router = NewRoutingTable()
	gossip.Timeouts = NewStatusResponseForwarder()
	return &gossip
}

// GossiperToString -
func (ent *Gossiper) GossiperToString() string {
	return fmt.Sprintf("ClienAddr: %s\nGossipAddr: %s\nName: %s\nSimpleMode: %v\n",
		ent.ClientAddr, ent.GossipAddr, ent.Name, ent.SimpleMode)
}
