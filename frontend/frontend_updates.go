package frontend

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

// FrontendPeer - A peer for the frontend
type FrontendPeer struct {
	IP   string // Peer's IP
	Port string // Peer's port
}

// FrontendPrivateMessage - A private message for the frontend
type FrontendPrivateMessage struct {
	Origin      string // Message's origin
	Destination string // Message's destination
	Msg         string // Peer's message
}

// FrontendPrivateContact - A private contact for the frontend
type FrontendPrivateContact struct {
	Name string // Peer's name
}

// FrontendIndexedFile - An indexed file for the frontend
type FrontendIndexedFile struct {
	Filename string // The filename
	Metahash string // The associated metahash
}

// FrontendConstructingFile - A constructing file for the frontend
type FrontendConstructingFile struct {
	Filename string // The filename
	Metahash string // The associated metahash
	Origin   string // The peer distributing the file
}

// FrontendUpdate - An update for the
type FrontendUpdate struct {
	Rumor            *FrontendRumor            // A rumor
	Peer             *FrontendPeer             // A peer
	PrivateMessage   *FrontendPrivateMessage   // A private message
	PrivateContact   *FrontendPrivateContact   // A private contact
	IndexedFile      *FrontendIndexedFile      // An indexed file
	ConstructingFile *FrontendConstructingFile // A constructing file
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

// AddFrontendPeer - Adds a peer to the buffer
func (buffer *FrontendBuffer) AddFrontendPeer(ip, port string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Create update
	newPeer := &FrontendPeer{IP: ip, Port: port}
	newUpdate := &FrontendUpdate{Peer: newPeer}
	buffer.updates = append(buffer.updates, newUpdate)
}

// AddFrontendPrivateMessage - Adds a private message to the buffer
func (buffer *FrontendBuffer) AddFrontendPrivateMessage(origin, destination, msg string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Prevent Javascript injection
	origin = strings.Replace(origin, "<", " &lt ", -1)
	origin = strings.Replace(origin, ">", " &gt ", -1)
	msg = strings.Replace(msg, "<", " &lt ", -1)
	msg = strings.Replace(msg, ">", " &gt ", -1)

	// Create update
	newPrivateMessage := &FrontendPrivateMessage{Origin: origin, Destination: destination, Msg: msg}
	newUpdate := &FrontendUpdate{PrivateMessage: newPrivateMessage}
	buffer.updates = append(buffer.updates, newUpdate)
}

// AddFrontendPrivateContact - Adds a private contact to the buffer
func (buffer *FrontendBuffer) AddFrontendPrivateContact(name string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Create update
	newPrivateContact := &FrontendPrivateContact{Name: name}
	newUpdate := &FrontendUpdate{PrivateContact: newPrivateContact}
	buffer.updates = append(buffer.updates, newUpdate)
}

// AddFrontendIndexedFile - Adds an indexed file to the buffer
func (buffer *FrontendBuffer) AddFrontendIndexedFile(filename, metahash string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Create update
	newIndexedFile := &FrontendIndexedFile{Filename: filename, Metahash: metahash}
	newUpdate := &FrontendUpdate{IndexedFile: newIndexedFile}
	buffer.updates = append(buffer.updates, newUpdate)
}

// AddFrontendConstructingFile - Adds a constructing file to the buffer
func (buffer *FrontendBuffer) AddFrontendConstructingFile(filename, metahash, origin string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Create update
	newConstructingFile := &FrontendConstructingFile{Filename: filename, Metahash: metahash, Origin: origin}
	newUpdate := &FrontendUpdate{ConstructingFile: newConstructingFile}
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
