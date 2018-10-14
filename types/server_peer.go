package types

import (
	"encoding/json"
	"sync"
)

// PeerBuffer - A buffer of peers for the frontend
type PeerBuffer struct {
	peers []ServerPeer // An array of PeerMessage
	mux   sync.Mutex   // Mutex to manipulate the structure from different threads
}

// ServerPeer - A peer for the frontend
type ServerPeer struct {
	IP   string // Peer's IP
	Port int    // Peer's port
}

// BufferPeers - A buffer of peers
var BufferPeers = NewPeerBuffer()

// NewPeerBuffer - Creates a new instance of PeerBuffer
func NewPeerBuffer() *PeerBuffer {
	var buffer PeerBuffer
	buffer.peers = nil
	return &buffer
}

// AddServerPeer - Adds a peer to the buffer
func (buffer *PeerBuffer) AddServerPeer(ip string, port int) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	buffer.peers = append(buffer.peers, ServerPeer{IP: ip, Port: port})
}

// GetDataAndEmpty - Empty the buffer and returns all its data as a byte slice
func (buffer *PeerBuffer) GetDataAndEmpty() []byte {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Check if the peer buffer isn't empty
	if buffer.peers == nil || len(buffer.peers) == 0 {
		return []byte("{}")
	}

	// Collect last messages
	data, _ := json.Marshal(map[string][]ServerPeer{
		"peers": buffer.peers,
	})
	buffer.peers = nil

	return data
}
