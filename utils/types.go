package utils

import (
	"net"
	"sync"
)

// BufSize - Size of the UDP buffer
const BufSize = 1024

// TimeoutSec - Length of generic timeout
const TimeoutSec = 1

// CustomError - Represents a custom error
type CustomError struct {
	Fun  string // Function's name
	Desc string // Error description
}

// Gossiper - Represents a gossiper
type Gossiper struct {
	ClientAddr string                  // IP/Port on which the client talks (RO)
	GossipAddr string                  // IP/Port on which to listen to other gossips (RO)
	Name       string                  // Name of that gossiper (RO)
	SimpleMode bool                    // Indicate whether the gossiper operated in simple broadcast mode (RO)
	Network    GossipNetwork           // The gossip network (Shared)
	Timeouts   StatusResponseForwarder // Timeouts for RumorMessage's answer (Shared)
}

// Client - Represents a gossiper
type Client struct {
	Addr string // IP/Port on which to talk
	Msg  string // Message to send
}

// Peer - Represents another peer
type Peer struct {
	RawAddr string       // An IP/Port pair <ip:port>
	UdpAddr *net.UDPAddr // A corresponding UDP address
}

// NamedPeer - Represents a named peer as well as the list of messages received from him
type NamedPeer struct {
	Name     string
	Messages []string
}

// GossipNetwork - Represents the known status of a gossip network
type GossipNetwork struct {
	Peers       []Peer       // A list of peers
	History     []NamedPeer  // A list of named peers with their history
	VectorClock StatusPacket // A vector clock
	Mux         sync.Mutex   // Mutex to manipulate the structure from different threads
}

// TimeoutHandler - Represents a StatusPacket answer to a RumorMessage
type TimeoutHandler struct {
	Addr *net.UDPAddr      // A peer's address
	Com  chan StatusPacket // A channel to communicate the status answer between threads
	Done bool              // Indicated whether a packet was already forwarded using this handler
}

// StatusResponseForwarder - Represents the set of pending timeouts
type StatusResponseForwarder struct {
	Responses []TimeoutHandler // An array of timeout handlers
	Mux       sync.Mutex       // Mutex to manipulate the structure from different threads
}

// SimpleMessage - Represents a simple user message (from client to local gossiper)
type SimpleMessage struct {
	OriginalName  string // Name of original sender
	RelayPeerAddr string // Address of last relayer
	Contents      string // Message content
}

// RumorMessage - Represents a rumor
type RumorMessage struct {
	Origin string // Name of original sender
	ID     uint32 // Message id (sequential)
	Text   string // Message content
}

// PeerStatus - Represent the status of a particular peer for a given gossiper
type PeerStatus struct {
	Identifier string // Name of original sender
	NextID     uint32 // Next expected message ID for this sender
}

// StatusPacket - Represents the status of all known peers of a given gossiper (vector clock)
type StatusPacket struct {
	Want []PeerStatus // Vector clock
}

// GossipPacket -- Structure that is exchanged between gossipers (only one of the 3-fields is non-nil)
type GossipPacket struct {
	SimpleMsg *SimpleMessage // A plain message (part 1)
	Rumor     *RumorMessage  // A rumor (part 2)
	Status    *StatusPacket  // A vector clock (part 2)
}
