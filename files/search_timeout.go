package files

import (
	"Peerster/messages"
	"fmt"
	"sync"
)

// Unique ID to identify a SearchRequestEntry
var uniqueID = uint64(0)

// SearchRequestMemory represents the set of pending timeouts for received search requests
type SearchRequestMemory struct {
	requests map[string][]*SearchRequestEntry // An index of peer name to a list of SearchRequestEntry
	mux      sync.Mutex                       // Mutex to manipulate the structure from different threads
}

// SearchRequestEntry represents an old SearchRequest
type SearchRequestEntry struct {
	keywords *[]string // A list of keywords
	uid      uint64    // Unique ID for this entry
}

// NewSearchRequestMemory creates a new instance of SearchRequestMemory
func NewSearchRequestMemory() *SearchRequestMemory {
	var memory SearchRequestMemory
	memory.requests = make(map[string][]*SearchRequestEntry)
	return &memory
}

// NewSearchRequestEntry creates a new instance of SearchRequestEntry The function must be called
// while holding the mutex on the SearchRequestMemory object.
func NewSearchRequestEntry(keywords *[]string) *SearchRequestEntry {
	var entry SearchRequestEntry
	entry.keywords = keywords
	entry.uid = uniqueID
	uniqueID++
	return &entry
}

// AddSearchRequestEntry adds a SearchRequest to a SearchRequestMemory
func (memory *SearchRequestMemory) AddSearchRequestEntry(request *messages.SearchRequest) uint64 {
	// Grab the mutex
	memory.mux.Lock()
	defer memory.mux.Unlock()

	// Create a new entry
	newRequest := NewSearchRequestEntry(&request.Keywords)

	if entries, ok := memory.requests[request.Origin]; ok { // We know this name
		entries = append(entries, newRequest)
	} else { // We don't know this name
		var oldRequests []*SearchRequestEntry
		oldRequests = append(oldRequests, newRequest)
		memory.requests[request.Origin] = oldRequests
	}

	return newRequest.uid
}

// RemoveSearchRequestEntry removes a SearchRequest from a SearchRequestMemory. The function
// panics if uidRemove doesn't coresspond to any known SearchRequestEntry.
func (memory *SearchRequestMemory) RemoveSearchRequestEntry(origin string, uidRemove uint64) {
	// Grab the mutex
	memory.mux.Lock()
	defer memory.mux.Unlock()

	if entries, ok := memory.requests[origin]; ok { // We know this name
		for _, entry := range entries {
			if entry.uid == uidRemove {

			}
		}
	}

	panic(fmt.Sprintf("RemoveSearchRequestEntry(): Trying to remove inexistant handler %d", uid))

}
