package types

import (
	"fmt"
	"os"
	"sync"
)

// NameIndex - Represents a dictionnary between peer names and received messages
type NameIndex struct {
	index map[string]*Messages // A mapping from peer name to messages
	mux   sync.Mutex           // Mutex to manipulate the structure from different threads
}

// Messages - Represents a list of messages
type Messages struct {
	list []string // A list of messages
}

// NewNameIndex - Creates a new instance of NameIndex
func NewNameIndex() *NameIndex {
	var nameIndex NameIndex
	nameIndex.index = make(map[string]*Messages)
	return &nameIndex
}

// NewMessages - Creates a new instance of Messages
func NewMessages() *Messages {
	var messages Messages
	messages.list = make([]string, 0)
	return &messages
}

// AddName - Adds a named peer to the index (thread-safe)
func (nameIndex *NameIndex) AddName(name string) {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	if _, ok := nameIndex.index[name]; !ok { // The name should not already exists
		nameIndex.index[name] = NewMessages()
	} else {
		fmt.Printf("ERROR: Trying to add existing name %s to the name index", name)
		os.Exit(1)
	}
}

// AddNameUnsafe - Adds a named peer to the index (not thread-safe)
func (nameIndex *NameIndex) AddNameUnsafe(name string) {
	if _, ok := nameIndex.index[name]; !ok { // The name should not already exists
		nameIndex.index[name] = NewMessages()
	} else {
		fmt.Printf("ERROR: Trying to add existing name %s to the name index", name)
	}
}

// AddMessageIfNext - Adds a message to the name index if we got all the preceding ones
func (nameIndex *NameIndex) AddMessageIfNext(rumor *RumorMessage) {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	if messages, ok := nameIndex.index[rumor.Origin]; ok { // We know this name
		if uint32(len(messages.list))+1 == rumor.ID { // Ensure message ordering
			messages.list = append(messages.list, rumor.Text)
			BufferMessages.AddServerMessage(rumor.Origin, rumor.Text)
		}
	} else { // We don't know this name
		if rumor.ID == 1 { // Must be the first message
			nameIndex.AddNameUnsafe(rumor.Origin)
			messages := nameIndex.index[rumor.Origin]
			messages.list = append(messages.list, rumor.Text)
			BufferMessages.AddServerMessage(rumor.Origin, rumor.Text)
		}
	}
}

// GetLastMessageID - Get the next message expected for a given name
func (nameIndex *NameIndex) GetLastMessageID(name string) uint32 {
	if messages, ok := nameIndex.index[name]; ok { // We know this name
		return uint32(len(messages.list)) + 1
	}
	return 0
}

// GetUnknownMessageTarget - Tries to find a message that we have but that the other doesn't
func (nameIndex *NameIndex) GetUnknownMessageTarget(targetStatus *StatusPacket) *RumorMessage {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	for localName, messages := range nameIndex.index {

		// Look for name in status
		indexPeer := -1
		for i, x := range targetStatus.Want {
			if x.Identifier == localName {
				indexPeer = i
				break
			}
		}

		if indexPeer != -1 { // We both now the peer
			nextIDWanted := targetStatus.Want[indexPeer].NextID
			if 0 < nextIDWanted && nextIDWanted < uint32(len(messages.list))+1 { // We have something the other doesn't have
				return &RumorMessage{localName, nextIDWanted, messages.list[nextIDWanted-1]}
			}
		} else if len(messages.list) > 0 { // The other doesn't know the peer, we have at least one message from him
			return &RumorMessage{localName, 1, messages.list[0]}
		}
	}
	return nil
}

// IsLocalStatusComplete - Checks whether we are not aware of some message that was received by the other
func (nameIndex *NameIndex) IsLocalStatusComplete(status *StatusPacket) bool {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	for _, distantPeer := range status.Want {

		if messages, ok := nameIndex.index[distantPeer.Identifier]; ok { // We know this name
			nextIDWanted := uint32(len(messages.list)) + 1
			if nextIDWanted < distantPeer.NextID { // The other has something we don't have
				return false
			}
		} else { // We don't know the peer
			nameIndex.AddNameUnsafe(distantPeer.Identifier)
			if distantPeer.NextID > 1 { // The other has something we don't have
				return false
			}
		}
	}

	return true
}

// GetVectorClock - Get the vector clock derived from the name index
func (nameIndex *NameIndex) GetVectorClock() *StatusPacket {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	var status StatusPacket
	status.Want = make([]PeerStatus, len(nameIndex.index))
	i := 0
	for name, messages := range nameIndex.index {
		status.Want[i] = PeerStatus{
			Identifier: name,
			NextID:     uint32(len(messages.list)) + 1}
		i++
	}

	return &status
}

// GetEverything - Returns everything that was received this far
func (nameIndex *NameIndex) GetEverything() *[]byte {

	buffer := NewMessageBuffer()

	// Retrieve everything
	nameIndex.mux.Lock()
	for name, messages := range nameIndex.index {
		for _, m := range messages.list {
			buffer.AddServerMessage(name, m)
		}
	}

	// Empty the "normal" buffer (we already have everything in the local one)
	BufferMessages.EmptyBuffer()

	nameIndex.mux.Unlock()

	return buffer.GetDataAndEmpty()
}
