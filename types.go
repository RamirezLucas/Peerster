package main

import (
	"fmt"
	"net"
	"sync"
)

// BufSize - Size of the UDP buffer
const BufSize = 1024

// TimeoutSec - Length of generic timeout
const TimeoutSec = 1

// CustomError -- Represents a custom error
type CustomError struct {
	fun  string // Function's name
	desc string // Error description
}

// Gossiper -- Represents a gossiper
type Gossiper struct {
	clientAddr string        // IP/Port on which the client talks (RO)
	gossipAddr string        // IP/Port on which to listen to other gossips (RO)
	name       string        // Name of that gossiper (RO)
	simpleMode bool          // Indicate whether the gossiper operated in simple broadcast mode (RO)
	network    GossipNetwork // The gossip network (Shared)
	mux        sync.Mutex    // Mutex to manipulate the structure from different threads
}

// Client -- Represents a gossiper
type Client struct {
	addr string // IP/Port on which to talk
	msg  string // Message to send
}

// Peer - Represents another peer
type Peer struct {
	rawAddr string         // An IP/Port pair <ip:port>
	udpAddr *net.UDPAddr   // A corresponding UDP address
	name    string         // The peer's name
	msgs    []RumorMessage // The list of messages received from that peer
}

// GossipNetwork - Represents the known status of a gossip network
type GossipNetwork struct {
	peers       []Peer       // A list of peers
	vectorClock StatusPacket // A vectorclock
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
	status    *StatusPacket  // A vector clock(part 2)
}

// CreatePeer - Creates a peer
func (p *Peer) CreatePeer(addr string) error {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return &CustomError{"CreatePeer", "cannot resolve UDP address"}
	}

	p.rawAddr = addr
	p.udpAddr = udpAddr

	return nil
}

// PeersToString - Returns a textual representation of a []Peer
func PeersToString(peers []Peer) string {
	s := "PEERS"
	for _, peer := range peers {
		s = s + fmt.Sprintf(" %s", peer.rawAddr)
	}
	return s + "\n"
}

// SimpleMessageToString - Returns a textual representation of a SimpleMessage
func SimpleMessageToString(msg *SimpleMessage) string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		msg.originalName, msg.relayPeerAddr, msg.contents)
}

// RumorMessageToString -- Returns a textual representation of a RumorMessage
func RumorMessageToString(rumor *RumorMessage, relayAddr string) string {
	return fmt.Sprintf("RUMOR origin %s from %s ID %d contents %s\n",
		rumor.origin, relayAddr, rumor.id, rumor.text)
}

// StatusPacketToString -- Returns a textual representation of a StatusPacket
func StatusPacketToString(status *StatusPacket, relayAddr string) string {
	s := fmt.Sprintf("STATUS from %s", relayAddr)
	for _, peer := range status.want {
		s = s + fmt.Sprintf(" peer %s nextID %d", peer.identifier, peer.nextID)
	}
	return s + "\n"
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.fun, e.desc)
}
