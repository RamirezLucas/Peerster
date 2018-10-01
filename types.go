package main

import (
	"fmt"
	"sync"
)

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

// SimpleMessage -- Represents a message
type SimpleMessage struct {
	originalName  string // Name of original sender
	relayPeerAddr string // Address of last relayer
	contents      string // Message content
}

// GossipPacket -- Represents a gossip packet
type GossipPacket struct {
	msg *SimpleMessage // A message
}

func (g *Gossiper) String() string {
	acc := fmt.Sprintf("clientAddr: %v\nname: %v\nsimpleMode: %v\ngossipAddr: %v\npeers:\n", g.clientAddr, g.name, g.simpleMode, g.gossipAddr)
	for _, x := range g.peers {
		acc = acc + fmt.Sprintf("\t%v\n", x)
	}
	return acc
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.fun, e.desc)
}
