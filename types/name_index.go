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

// Messages - Represents a public of messages
type Messages struct {
	public  []string // A list of public messages
	private []string // A list of (unordered) private messages
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
	messages.public = make([]string, 0)
	return &messages
}

// AddName - Adds a named peer to the index (thread-safe)
func (nameIndex *NameIndex) AddName(name string) {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	nameIndex.AddNameUnsafe(name)
}

// AddNameUnsafe - Adds a named peer to the index (not thread-safe)
func (nameIndex *NameIndex) AddNameUnsafe(name string) {
	if _, ok := nameIndex.index[name]; !ok { // The name should not already exists
		nameIndex.index[name] = NewMessages()
	} else {
		fmt.Printf("ERROR: Trying to add existing name %s to the name index", name)
		os.Exit(1)
	}
}

// AddPrivateMessage - Adds a private message to the name index (no order imposed)
func (nameIndex *NameIndex) AddPrivateMessage(private *PrivateMessage) {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	if messages, ok := nameIndex.index[private.Origin]; ok { // We know this name
		messages.private = append(messages.private, private.Text)

	} else { // We don't know this name
		nameIndex.AddNameUnsafe(private.Origin)
		messages := nameIndex.index[private.Origin]
		messages.private = append(messages.private, private.Text)
	}

	// Forward to frontend
	FBuffer.AddFrontendPrivateMessage(private.Origin, private.Destination, private.Text)
}

// AddMessageIfNext - Adds a message to the name index if we got all the preceding ones
func (nameIndex *NameIndex) AddMessageIfNext(rumor *RumorMessage) bool {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	// Is the message a RouteRumor ?
	isRouteRumor := (rumor.Text == "")

	if messages, ok := nameIndex.index[rumor.Origin]; ok { // We know this name
		if uint32(len(messages.public))+1 == rumor.ID { // Ensure message ordering
			messages.public = append(messages.public, rumor.Text)

			// Don't forward route rumors to the server
			if !isRouteRumor {
				FBuffer.AddFrontendRumor(rumor.Origin, rumor.Text)
			}

			return true
		}
	} else { // We don't know this name
		if rumor.ID == 1 { // Must be the first message
			nameIndex.AddNameUnsafe(rumor.Origin)
			messages := nameIndex.index[rumor.Origin]
			messages.public = append(messages.public, rumor.Text)

			// Don't forward route rumors to the server
			if !isRouteRumor {
				FBuffer.AddFrontendRumor(rumor.Origin, rumor.Text)
			}

			return true
		}
	}
	return false
}

// FillInRumorAndSave - Fills in a rumor coming from the client and store it
func (nameIndex *NameIndex) FillInRumorAndSave(rumor *RumorMessage, origin string) {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	// Is the message a RouteRumor ?
	isRouteRumor := (rumor.Text == "")

	if messages, ok := nameIndex.index[origin]; ok { // We know this name
		// Fill in the rumor
		rumor.Origin = origin
		rumor.ID = uint32(len(messages.public)) + 1
		// Store it
		messages.public = append(messages.public, rumor.Text)

		// Don't forward route rumors to the server
		if !isRouteRumor {
			FBuffer.AddFrontendRumor(rumor.Origin, rumor.Text)
		}

		return
	}

	// Should not happen
	fmt.Printf("ERROR: Trying to get the last message ID of non-client peer %s\n", origin)
	os.Exit(1)
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
			if 0 < nextIDWanted && nextIDWanted < uint32(len(messages.public))+1 { // We have something the other doesn't have
				return &RumorMessage{localName, nextIDWanted, messages.public[nextIDWanted-1]}
			}
		} else if len(messages.public) > 0 { // The other doesn't know the peer, we have at least one message from him
			return &RumorMessage{localName, 1, messages.public[0]}
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
			nextIDWanted := uint32(len(messages.public)) + 1
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
			NextID:     uint32(len(messages.public)) + 1}
		i++
	}

	return &status
}
