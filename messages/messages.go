package messages

import (
	"fmt"
)

// SimpleMessage represents a simple message
type SimpleMessage struct {
	OriginalName  string // Name of original sender
	RelayPeerAddr string // Address of last relayer
	Contents      string // Message content
}

// RumorMessage represents a rumor message
type RumorMessage struct {
	Origin string // Name of original sender
	ID     uint32 // Message id (sequential)
	Text   string // Message content
}

// PeerStatus represent the status of a particular peer for a given gossiper
type PeerStatus struct {
	Identifier string // Name of original sender
	NextID     uint32 // Next expected message ID for this sender
}

// StatusPacket represents the status of all known peers of a given gossiper (vector clock)
type StatusPacket struct {
	Want []PeerStatus // Vector clock
}

// PrivateMessage represents a private message between 2 peers
type PrivateMessage struct {
	Origin      string // The sender's name
	ID          uint32 // The message ID (not important for now as sequence order isn't enforced)
	Text        string // The message's content
	Destination string // The destination's name
	HopLimit    uint32 // The maximum number of hops the message is allowed to go through
}

// DataRequest represents a data request
type DataRequest struct {
	Origin      string // The message's origin
	Destination string // The message's destination
	HopLimit    uint32 // The maximum number of hops the message is allowed to go through
	HashValue   []byte // The requeted hash value
}

// DataReply represents a data reply
type DataReply struct {
	Origin      string // The message's origin
	Destination string // The message's destination
	HopLimit    uint32 // The maximum number of hops the message is allowed to go through
	HashValue   []byte // The hash value computed from the Data field
	Data        []byte // Data
}

// SearchRequest represents a search request
type SearchRequest struct {
	Origin   string   // The message's origin
	Budget   uint64   // The budget allocated to this request
	Keywords []string // A list of keywords to research
}

// SearchReply represents a search reply
type SearchReply struct {
	Origin      string          // The message's origin
	Destination string          // The message's destination
	HopLimit    uint32          // The maximum number of hops the message is allowed to go through
	Results     []*SearchResult // A list of search results
}

// SearchResult represents a search result
type SearchResult struct {
	FileName     string   // The filename associated to the result
	MetafileHash []byte   // The file's metahash
	ChunkMap     []uint64 // The indices of the chunks that the replying peer contains locally
	ChunkCount   uint64   // Number of chunks for this file
}

// GossipPacket is the structure that is exchanged between gossipers (only one of the fields can be non-nil)
type GossipPacket struct {
	SimpleMsg     *SimpleMessage  // A plain message
	Rumor         *RumorMessage   // A rumor message
	Status        *StatusPacket   // A vector clock
	Private       *PrivateMessage // A private message
	DataRequest   *DataRequest    // A data request
	DataReply     *DataReply      // A data reply
	SearchRequest *SearchRequest  // A search request
	SearchReply   *SearchReply    // A search reply
}

// SimpleMessageToString returns a textual representation of a SimpleMessage
func (pkt *SimpleMessage) SimpleMessageToString() string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s",
		pkt.OriginalName, pkt.RelayPeerAddr, pkt.Contents)
}

// RumorMessageToString returns a textual representation of a RumorMessage
func (pkt *RumorMessage) RumorMessageToString(relayAddr string) string {
	return fmt.Sprintf("RUMOR origin %s from %s ID %d contents %s",
		pkt.Origin, relayAddr, pkt.ID, pkt.Text)
}

// StatusPacketToString returns a textual representation of a StatusPacket
func (pkt *StatusPacket) StatusPacketToString(relayAddr string) string {
	s := fmt.Sprintf("STATUS from %s", relayAddr)
	for _, peer := range pkt.Want {
		s = s + fmt.Sprintf(" peer %s nextID %d", peer.Identifier, peer.NextID)
	}
	return s
}

// PrivateMessageToString returns a textual representation of a PrivateMessage
func (pkt *PrivateMessage) PrivateMessageToString() string {
	return fmt.Sprintf("PRIVATE origin %s hop_limit %d contents %s",
		pkt.Origin, pkt.HopLimit, pkt.Text)
}
