package types

import (
	"encoding/json"
	"strings"
	"sync"
)

// FrontendBuffer - A buffer of updates for the frontend
type FrontendBuffer struct {
	updates []*FrontendUpdate // An array of FrontendUpdate
	mux     sync.Mutex        // Mutex to manipulate the structure from different threads
}

// FrontendRumor - A rumor for the frontend
type FrontendRumor struct {
	Name string // Peer's name
	Msg  string // Peer's message
}

// FrontendPrivateMessage - A private message for the frontend
type FrontendPrivateMessage struct {
	Name string // Peer's name
	Msg  string // Peer's message
}

// FrontendPeer - A peer for the frontend
type FrontendPeer struct {
	IP   string // Peer's IP
	Port string // Peer's port
}

// FrontendUpdate - An update for the
type FrontendUpdate struct {
	Rumor          *FrontendRumor          // A rumor
	PrivateMessage *FrontendPrivateMessage // A private message
	Peer           *FrontendPeer           // A peer
}

// FBuffer - A buffer of updates for the frontend
var FBuffer = NewFrontendBuffer()

// NewFrontendBuffer - Creates a new instance of FrontendBuffer
func NewFrontendBuffer() *FrontendBuffer {
	var buffer FrontendBuffer
	buffer.updates = nil
	return &buffer
}

// EmptyBuffer - Empty the buffer
func (buffer *FrontendBuffer) EmptyBuffer() {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()
	buffer.updates = nil
}

// AddFrontendRumor - Adds a rumor to the buffer
func (buffer *FrontendBuffer) AddFrontendRumor(name, msg string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Prevent Javascript injection
	name = strings.Replace(name, "<", " &lt ", -1)
	name = strings.Replace(name, ">", " &gt ", -1)
	msg = strings.Replace(msg, "<", " &lt ", -1)
	msg = strings.Replace(msg, ">", " &gt ", -1)

	// Create update
	newRumor := &FrontendRumor{Name: name, Msg: msg}
	newUpdate := &FrontendUpdate{Rumor: newRumor}
	buffer.updates = append(buffer.updates, newUpdate)
}

// AddFrontendPrivateMessage - Adds a private message to the buffer
func (buffer *FrontendBuffer) AddFrontendPrivateMessage(name, msg string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Prevent Javascript injection
	name = strings.Replace(name, "<", " &lt ", -1)
	name = strings.Replace(name, ">", " &gt ", -1)
	msg = strings.Replace(msg, "<", " &lt ", -1)
	msg = strings.Replace(msg, ">", " &gt ", -1)

	// Create update
	newPrivateMessage := &FrontendPrivateMessage{Name: name, Msg: msg}
	newUpdate := &FrontendUpdate{PrivateMessage: newPrivateMessage}
	buffer.updates = append(buffer.updates, newUpdate)
}

// AddFrontendPeer - Adds a peer to the buffer
func (buffer *FrontendBuffer) AddFrontendPeer(ip, port string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Create update
	newPeer := &FrontendPeer{IP: ip, Port: port}
	newUpdate := &FrontendUpdate{Peer: newPeer}
	buffer.updates = append(buffer.updates, newUpdate)
}

// GetDataAndEmpty - Empty the buffer and returns all its data as a byte slice
func (buffer *FrontendBuffer) GetDataAndEmpty() *[]byte {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	data := []byte("{}")

	// Check if the buffer isn't empty
	if buffer.updates == nil {
		return &data
	}

	// Collect last updates
	data, _ = json.Marshal(map[string][]*FrontendUpdate{
		"updates": buffer.updates,
	})

	// Erase the buffer
	buffer.updates = nil

	return &data

}
