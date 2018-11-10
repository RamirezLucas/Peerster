package types

import (
	"sync"
)

// DataResponseForwarder - Represents the set of pending timeouts for data requests
type DataResponseForwarder struct {
	responses map[string]*DataTimeoutHandler // An index of timeout handlers
	mux       sync.Mutex                     // Mutex to manipulate the structure from different threads
}

// DataTimeoutHandler - Represents a timeout handler for data requests
type DataTimeoutHandler struct {
	Origin string     // The peer's name that should be contained in the Origin field of the DataReply
	Done   bool       // Acknowledges that the handler has been executed
	Hash   *KnownHash // Information regarding this particular hash
}

// NewDataResponseForwarder - Creates a new instance of DataResponseForwarder
func NewDataResponseForwarder() *DataResponseForwarder {
	var forwarder DataResponseForwarder
	forwarder.responses = make(map[string]*DataTimeoutHandler)
	return &forwarder
}

// NewDataTimeoutHandler - Creates a new instance of DataTimeoutHandler
func NewDataTimeoutHandler(origin string, file *SharedFile, isMetahash bool, chunkIndex uint32) *DataTimeoutHandler {
	var handler DataTimeoutHandler
	handler.Origin = origin
	handler.Done = false
	handler.Hash = NewKnownHash(file, isMetahash, chunkIndex)
	return &handler
}

// AddDataTimeoutHandler - Adds a DataTimeoutHandler to the forwarder
func (forwarder *DataResponseForwarder) AddDataTimeoutHandler(hash []byte, origin string,
	file *SharedFile, isMetahash bool, chunkIndex uint32) *DataTimeoutHandler {

	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	handler := NewDataTimeoutHandler(origin, file, isMetahash, chunkIndex)
	forwarder.responses[string(hash[:])] = handler
	return handler
}

// DeleteDataTimeoutHandler - Deletes a DataTimeoutHandler from the forwarder
func (forwarder *DataResponseForwarder) DeleteDataTimeoutHandler(hash []byte) {
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	strHash := string(hash[:])

	if _, ok := forwarder.responses[strHash]; ok {
		delete(forwarder.responses, strHash)
	} else {
		panic("DeleteDataTimeoutHandler(): Trying to delete non-existing data handler")
	}
}

// SearchHashAndForward - Searches the set of handlers for a given hash. Accept the reply on match
func (forwarder *DataResponseForwarder) SearchHashAndForward(hash []byte, origin string) *KnownHash {
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	if match, ok := forwarder.responses[string(hash[:])]; ok { // We were waiting for this hash
		if !match.Done && match.Origin == origin { // Check that data was sent from the correct peer
			// Acknowledges to the sender thread
			match.Done = true
			// Return the information concerning this hash
			return match.Hash
		}
	}
	return nil
}

// CheckResponseReceived - Checks whether the response to the given hash was received
func (forwarder *DataResponseForwarder) CheckResponseReceived(hash []byte) bool {
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	if handler, ok := forwarder.responses[string(hash[:])]; ok {
		return handler.Done
	}

	panic("CheckResponseReceived(): Trying to check an inexistant handler")
}
