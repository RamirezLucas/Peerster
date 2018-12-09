package peers

import (
	"Peerster/fail"
	"Peerster/frontend"
	"Peerster/messages"
	"sync"
)

// NameIndex - Represents a dictionnary between peer names and received messages
type NameIndex struct {
	index map[string]*Messages // A mapping from peer name to messages
	mux   sync.Mutex           // Mutex to manipulate the structure from different threads
}

// Messages - Represents the list of public and private messages received by a peer
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

	nameIndex.addNameUnsafe(name)
}

// addNameUnsafe - Adds a named peer to the index (not thread-safe)
func (nameIndex *NameIndex) addNameUnsafe(name string) {
	if _, ok := nameIndex.index[name]; !ok { // The name should not already exists
		nameIndex.index[name] = NewMessages()
	} else {
		fail.CustomPanic("NameIndex.addNameUnsafe", "Trying to add existing name %s to the index.", name)
	}
}

// AddPrivateMessage - Adds a private message to the name index (no order imposed)
func (nameIndex *NameIndex) AddPrivateMessage(private *messages.PrivateMessage) {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	if messages, ok := nameIndex.index[private.Origin]; ok { // We know this name
		messages.private = append(messages.private, private.Text)

	} else { // We don't know this name
		nameIndex.addNameUnsafe(private.Origin)
		messages := nameIndex.index[private.Origin]
		messages.private = append(messages.private, private.Text)
	}

	// Forward to frontend
	frontend.FBuffer.AddFrontendPrivateMessage(private.Origin, private.Destination, private.Text)
}

// AddMessageIfNext - Adds a message to the name index if we got all the preceding ones
func (nameIndex *NameIndex) AddMessageIfNext(rumor *messages.RumorMessage) bool {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	// Is the message a RouteRumor ?
	isRouteRumor := (rumor.Text == "")

	if messages, ok := nameIndex.index[rumor.Origin]; ok { // We know this name
		if uint32(len(messages.public))+1 == rumor.ID { // Ensure message ordering
			messages.public = append(messages.public, rumor.Text)

			// Don't forward route rumors to the server
			if !isRouteRumor {
				frontend.FBuffer.AddFrontendRumor(rumor.Origin, rumor.Text)
			}

			return true
		}
	} else { // We don't know this name
		if rumor.ID == 1 { // Must be the first message
			nameIndex.addNameUnsafe(rumor.Origin)
			messages := nameIndex.index[rumor.Origin]
			messages.public = append(messages.public, rumor.Text)

			// Don't forward route rumors to the server
			if !isRouteRumor {
				frontend.FBuffer.AddFrontendRumor(rumor.Origin, rumor.Text)
			}

			return true
		}
	}
	return false
}

// FillInRumorAndSave - Fills in a rumor coming from the client and store it
func (nameIndex *NameIndex) FillInRumorAndSave(rumor *messages.RumorMessage, origin string) {
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
			frontend.FBuffer.AddFrontendRumor(rumor.Origin, rumor.Text)
		}

		return
	}

	// Should not happen
	fail.CustomPanic("NameIndex.FillInRumorAndSave", "Trying to get last message ID of non-client peer %s.", origin)
}

// GetUnknownMessageTarget - Tries to find a message that we have but that the other doesn't
func (nameIndex *NameIndex) GetUnknownMessageTarget(targetStatus *messages.StatusPacket) *messages.RumorMessage {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	for localName, msgs := range nameIndex.index {

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
			if 0 < nextIDWanted && nextIDWanted < uint32(len(msgs.public))+1 { // We have something the other doesn't have
				return &messages.RumorMessage{Origin: localName, ID: nextIDWanted, Text: msgs.public[nextIDWanted-1]}
			}
		} else if len(msgs.public) > 0 { // The other doesn't know the peer, we have at least one message from him
			return &messages.RumorMessage{Origin: localName, ID: 1, Text: msgs.public[0]}
		}
	}
	return nil
}

// IsLocalStatusComplete - Checks whether we are not aware of some message that was received by the other
func (nameIndex *NameIndex) IsLocalStatusComplete(status *messages.StatusPacket) bool {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	for _, distantPeer := range status.Want {

		if messages, ok := nameIndex.index[distantPeer.Identifier]; ok { // We know this name
			nextIDWanted := uint32(len(messages.public)) + 1
			if nextIDWanted < distantPeer.NextID { // The other has something we don't have
				return false
			}
		} else { // We don't know the peer
			nameIndex.addNameUnsafe(distantPeer.Identifier)
			if distantPeer.NextID > 1 { // The other has something we don't have
				return false
			}
		}
	}

	return true
}

// GetVectorClock - Get the vector clock derived from the name index
func (nameIndex *NameIndex) GetVectorClock() *messages.StatusPacket {
	nameIndex.mux.Lock()
	defer nameIndex.mux.Unlock()

	var status messages.StatusPacket
	status.Want = make([]messages.PeerStatus, len(nameIndex.index))
	i := 0
	for name, msgs := range nameIndex.index {
		status.Want[i] = messages.PeerStatus{
			Identifier: name,
			NextID:     uint32(len(msgs.public)) + 1}
		i++
	}

	return &status
}
