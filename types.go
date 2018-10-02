package main

import (
	"fmt"
	"sync"
)

// BufSize - Size of the UDP buffer
const BufSize = 1024

// CustomError -- Represents a custom error
type CustomError struct {
	fun  string // Function's name
	desc string // Error description
}

// Gossiper -- Represents a gossiper
type Gossiper struct {
	clientAddr string     // IP/Port on which the client talks
	gossipAddr string     // IP/Port on which to listen to other gossips
	name       string     // Name of that gossiper
	peers      []string   // List of known peers
	simpleMode bool       // Indicate whether the gossiper operated in simple mode (broadcast)
	mux        sync.Mutex // Mutex to manipulate the structure from different threads
}

// Client -- Represents a gossiper
type Client struct {
	addr string // IP/Port on which to talk
	msg  string // Message to send
}

// SimpleMessage -- Represents a simple user message (from client to local gossiper)
type SimpleMessage struct {
	originalName  string // Name of original sender
	relayPeerAddr string // Address of last relayer
	contents      string // Message content
}

// RumorMessage - Represents a rumor
type RumorMessage struct {
	origin string // Name of original sender
	iD     uint32 // Message id (sequential)
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
	msg    *SimpleMessage // A plain message (part 1)
	rumor  *RumorMessage  // A rumor (part 2)
	status *StatusPacket  // A vector clock(part 2)
}

func (g *Gossiper) String() string {
	acc := fmt.Sprintf("clientAddr: %v\nname: %v\nsimpleMode: %v\ngossipAddr: %v\npeers:\n", g.clientAddr, g.name, g.simpleMode, g.gossipAddr)
	for _, x := range g.peers {
		acc = acc + fmt.Sprintf("\t%v\n", x)
	}
	return acc
}

func (msg *SimpleMessage) String() string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s", msg.originalName, msg.relayPeerAddr, msg.contents)
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.fun, e.desc)
}
