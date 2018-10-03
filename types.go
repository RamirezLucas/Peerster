package main

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
	fun  string // Function's name
	desc string // Error description
}

// Gossiper - Represents a gossiper
type Gossiper struct {
	clientAddr string        // IP/Port on which the client talks (RO)
	gossipAddr string        // IP/Port on which to listen to other gossips (RO)
	name       string        // Name of that gossiper (RO)
	simpleMode bool          // Indicate whether the gossiper operated in simple broadcast mode (RO)
	network    GossipNetwork // The gossip network (Shared)
	mux        sync.Mutex    // Mutex to manipulate the structure from different threads
}

// Client - Represents a gossiper
type Client struct {
	addr string // IP/Port on which to talk
	msg  string // Message to send
}

// Peer - Represents another peer
type Peer struct {
	rawAddr string       // An IP/Port pair <ip:port>
	udpAddr *net.UDPAddr // A corresponding UDP address
}

// NamedPeer - Represents a named peer as well as the list of messages received from him
type NamedPeer struct {
	name     string
	messages []string
}

// GossipNetwork - Represents the known status of a gossip network
type GossipNetwork struct {
	peers       []Peer       // A list of peers
	history     []NamedPeer  // A list of named peers with their history
	vectorClock StatusPacket // A vector clock
}

// SimpleMessage - Represents a simple user message (from client to local gossiper)
type SimpleMessage struct {
	originalName  string // Name of original sender
	relayPeerAddr string // Address of last relayer
	contents      string // Message content
}

// RumorMessage - Represents a rumor
type RumorMessage struct {
	origin string // Name of original sender
	id     uint32 // Message id (sequential)
	text   string // Message content
}

// PeerStatus - Represent the status of a particular peer for a given gossiper
type PeerStatus struct {
	identifier string // Name of original sender
	nextID     uint32 // Next expected message ID for this sender
}

// StatusPacket - Represents the status of all known peers of a given gossiper (vector clock)
type StatusPacket struct {
	want []PeerStatus // Vector clock
}

// GossipPacket -- Structure that is exchanged between gossipers (only one of the 3-fields is non-nil)
type GossipPacket struct {
	simpleMsg *SimpleMessage // A plain message (part 1)
	rumor     *RumorMessage  // A rumor (part 2)
	status    *StatusPacket  // A vector clock (part 2)
}
