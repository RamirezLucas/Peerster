package files

import (
	"Peerster/fail"
	"Peerster/messages"
	"sync"
)

// TODataRequest - Represents the set of pending timeouts for data requests
type TODataRequest struct {
	responses map[string]*DataRequestHandler // An index of timeout handlers
	mux       sync.Mutex                     // Mutex to manipulate the structure from different threads
}

// DataRequestHandler - Represents a timeout handler for data requests
type DataRequestHandler struct {
	Origin string   // The peer's name that should be contained in the Origin field of the DataReply
	Done   bool     // Acknowledges that the handler has been executed
	Hash   *HashRef // Information regarding this particular hash
}

// NewTODataRequest - Creates a new instance of TODataRequest
func NewTODataRequest() *TODataRequest {
	var forwarder TODataRequest
	forwarder.responses = make(map[string]*DataRequestHandler)
	return &forwarder
}

// NewDataRequestHandler - Creates a new instance of TODataRequestHandler
func NewDataRequestHandler(origin string, hash *HashRef) *DataRequestHandler {
	var handler DataRequestHandler
	handler.Origin = origin
	handler.Hash = hash
	return &handler
}

// AddDataRequest - Adds a TODataRequestHandler to the forwarder
func (forwarder *TODataRequest) AddDataRequest(request *messages.DataRequest, hash *HashRef) bool {
	// Grab the mutex
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	// Attempt to add a new handler
	strHash := ToHex(request.HashValue[:])
	handler := NewDataRequestHandler(request.Origin, hash)
	if _, ok := forwarder.responses[strHash]; ok { // We already have a request pending for this hash
		return false
	}
	forwarder.responses[strHash] = handler
	return true
}

// CheckResponseAndDelete - Checks whether the response to the given hash was received
func (forwarder *TODataRequest) CheckResponseAndDelete(hash []byte) bool {
	// Grab the mutex
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	strHash := ToHex(hash[:])

	// Check if a response was received
	if handler, ok := forwarder.responses[strHash]; ok {
		if handler.Done {
			// Delete the handler
			delete(forwarder.responses, strHash)
			return true
		}
		return false
	}

	// Should not happen
	fail.CustomPanic("CheckResponseAndDelete", "Attempting to delete an inexistant handler with hash %s", strHash)
	return false
}

// SearchHashAndAcknowledge - Searches the set of handlers for a given hash. Accept the reply on match
func (forwarder *TODataRequest) SearchHashAndAcknowledge(reply *messages.DataReply) *HashRef {
	// Grab the mutex
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	if match, ok := forwarder.responses[ToHex(reply.HashValue[:])]; ok { // We were waiting for this hash
		if !match.Done && match.Origin == reply.Origin { // Check that data was sent from the correct peer
			// Acknowledges to the sender thread and return
			match.Done = true
			return match.Hash
		}
	}
	return nil
}
