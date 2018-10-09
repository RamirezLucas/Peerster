package types

import (
	"fmt"
)

// SimpleMessage - Represents a simple message
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

// SimpleMessageToString - Returns a textual representation of a SimpleMessage
func (pkt *SimpleMessage) SimpleMessageToString() string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s",
		pkt.OriginalName, pkt.RelayPeerAddr, pkt.Contents)
}

// RumorMessageToString - Returns a textual representation of a RumorMessage
func (pkt *RumorMessage) RumorMessageToString(relayAddr string) string {
	return fmt.Sprintf("RUMOR origin %s from %s ID %d contents %s",
		pkt.Origin, relayAddr, pkt.ID, pkt.Text)
}

// StatusPacketToString - Returns a textual representation of a StatusPacket
func (pkt *StatusPacket) StatusPacketToString(relayAddr string) string {
	s := fmt.Sprintf("STATUS from %s", relayAddr)
	for _, peer := range pkt.Want {
		s = s + fmt.Sprintf(" peer %s nextID %d", peer.Identifier, peer.NextID)
	}
	return s
}
