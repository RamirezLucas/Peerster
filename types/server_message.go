package types

import (
	"encoding/json"
	"strings"
	"sync"
)

// MessageBuffer - A buffer of messages for the frontend
type MessageBuffer struct {
	messages []ServerMessage // An array of ServerMessage
	mux      sync.Mutex      // Mutex to manipulate the structure from different threads
}

// ServerMessage - A message for the frontend
type ServerMessage struct {
	Name string // Peer's name
	Msg  string // Peer's message
}

// BufferMessages - A buffer of messages
var BufferMessages = NewMessageBuffer()

// NewMessageBuffer - Creates a new instance of MessageBuffer
func NewMessageBuffer() *MessageBuffer {
	var buffer MessageBuffer
	buffer.messages = nil
	return &buffer
}

// AddServerMessage - Adds a message to the buffer
func (buffer *MessageBuffer) AddServerMessage(name, msg string) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Prevent Javascript injection
	msg = strings.Replace(msg, "<", " &lt ", -1)
	msg = strings.Replace(msg, ">", " &gt ", -1)

	buffer.messages = append(buffer.messages, ServerMessage{Name: name, Msg: msg})
}

// GetDataAndEmpty - Empty the buffer and returns all its data as a byte slice
func (buffer *MessageBuffer) GetDataAndEmpty() *[]byte {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	data := []byte("{}")

	// Check if the peer buffer isn't empty
	if buffer.messages == nil || len(buffer.messages) == 0 {
		return &data
	}

	// Collect last messages
	data, _ = json.Marshal(map[string][]ServerMessage{
		"messages": buffer.messages,
	})
	buffer.messages = nil

	return &data

}
