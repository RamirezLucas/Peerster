package files

import (
	"Peerster/fail"
	"Peerster/messages"
	"crypto/sha256"
	"strings"
	"sync"
)

// TOSearchRequest represents the set of pending timeouts for received search requests
type TOSearchRequest struct {
	requests map[string]emptyStruct // An index of previously received SearchRequests (origin and keywords hashed)
	mux      sync.Mutex             // Mutex to manipulate the structure from different threads
}

type emptyStruct struct {
}

// NewTOSearchRequest creates a new instance of TOSearchRequest
func NewTOSearchRequest() *TOSearchRequest {
	var memory TOSearchRequest
	memory.requests = make(map[string]emptyStruct)
	return &memory
}

// AddSearchRequest adds a SearchRequest to a TOSearchRequest
func (memory *TOSearchRequest) AddSearchRequest(request *messages.SearchRequest) bool {
	// Grab the mutex
	memory.mux.Lock()
	defer memory.mux.Unlock()

	// Compute hash based on SearchRequest's origin and keywords
	hashStr := hashSearchRequest(request)

	if _, ok := memory.requests[hashStr]; ok { // We already know this SearchRequest
		return false
	}

	// Add the SearchRequest to the TOSearchRequest and return the hash
	memory.requests[hashStr] = emptyStruct{}
	return true
}

// RemoveSearchRequest removes the SearchRequest represented by the given hashStr
// from the TOSearchRequest.
func (memory *TOSearchRequest) RemoveSearchRequest(request *messages.SearchRequest) {
	// Grab the mutex
	memory.mux.Lock()
	defer memory.mux.Unlock()

	// Compute hash based on SearchRequest's origin and keywords
	hashStr := hashSearchRequest(request)

	if _, ok := memory.requests[hashStr]; ok { // We know this hash
		delete(memory.requests, hashStr)
		return
	}

	fail.CustomPanic("TOSearchRequest.RemoveSearchRequest", "Trying to remove inexistant SearchRequest %s.", hashStr)
}

// FindSearchRequest attempts to find a SearchRequest in the TOSearchRequest
func (memory *TOSearchRequest) FindSearchRequest(request *messages.SearchRequest) bool {
	// Grab the mutex
	memory.mux.Lock()
	defer memory.mux.Unlock()

	// Compute hash based on SearchRequest's origin and keywords
	hashStr := hashSearchRequest(request)

	// Returns the search result
	_, foundHash := memory.requests[hashStr]
	return foundHash
}

// hashSearchRequest returns the hash of a SearchRequest
func hashSearchRequest(request *messages.SearchRequest) string {
	data := []byte(request.Origin + strings.Join(request.Keywords, ","))
	hash := sha256.Sum256(data[:])
	return ToHex(hash[:])
}
