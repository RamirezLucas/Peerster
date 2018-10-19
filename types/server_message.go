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
	Name      string // Peer's name
	Msg       string // Peer's message
	IsPrivate bool   // Set to true is the message is private
}

// BufferMessages - A buffer of messages
var BufferMessages = NewMessageBuffer()

// NewMessageBuffer - Creates a new instance of MessageBuffer
func NewMessageBuffer() *MessageBuffer {
	var buffer MessageBuffer
	buffer.messages = nil
	return &buffer
}

// EmptyBuffer - Empty the buffer
func (buffer *MessageBuffer) EmptyBuffer() {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()
	buffer.messages = nil
}

// AddServerMessage - Adds a message to the buffer
func (buffer *MessageBuffer) AddServerMessage(name, msg string, isPrivate bool) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()

	// Prevent Javascript injection
	msg = strings.Replace(msg, "<", " &lt ", -1)
	msg = strings.Replace(msg, ">", " &gt ", -1)

	buffer.messages = append(buffer.messages, ServerMessage{Name: name, Msg: msg, IsPrivate: isPrivate})
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
