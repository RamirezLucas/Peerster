package files

import (
	"Peerster/messages"
	"strings"
	"sync"
)

/*SReqTotalMatch is used to count the number of total matches associated to an emmited
SearchRequest. A SearchRequest is identified by its list of keywords. Each time a SearchReply is
received and triggers a total match, any SearchRequest including a keyword that is contained in the
SearchReply's `filename` field will see its number of total matches incremented by 1.

A TimeoutSearchReply object should be created by calling NewTimeoutSearchRequest(). Once created, the
object is thread-safe, meaning that several threads may manipulate the object through its API simultaneously.*/
type SReqTotalMatch struct {
	requests map[string]uint64 // An index associating a SearchRequest to its number of associated total matches
	mux      sync.Mutex        // Mutex to manipulate the structure from different threads
}

/*NewSReqTotalMatch creates a new instance of SReqTotalMatch.*/
func NewSReqTotalMatch() *SReqTotalMatch {
	var timeout SReqTotalMatch
	timeout.requests = make(map[string]uint64)
	return &timeout
}

/*AddEmittedSearchRequest adds a SearchRequest to the index. If an identical SearchRequest
was already registered simply reset the number of associated total matches to 0. */
func (timeout *SReqTotalMatch) AddEmittedSearchRequest(request *messages.SearchRequest) {
	// Grab the mutex
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	// Join the keywords
	keywordsJoin := strings.Join(request.Keywords, ",")

	/* If the index already contains these keywords then assume that the user wants to
	"reset" the search. In that case the number of associated total matches is reset to 0. */
	timeout.requests[keywordsJoin] = 0
}

/*CheckThresholdAndDelete checks whether at least `threshold` total matches have been associated
to the `request`. If that's the case, then the corresponding entry in the index is deleted and true
is returned.

If the entry corresponding to `request` does not exist true is returned. This can happen when two identical
SearchRequest are sent consecutively. Otherwise false is returned.

The caller should stop sending SearchRequest's with increased budget when true is returned.*/
func (timeout *SReqTotalMatch) CheckThresholdAndDelete(request *messages.SearchRequest, threshold uint64) bool {
	// Grab the mutex
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	// Join the keywords
	keywordsJoin := strings.Join(request.Keywords, ",")

	if matches, ok := timeout.requests[keywordsJoin]; ok { // We know these keywords
		if matches >= threshold {
			delete(timeout.requests, keywordsJoin)
			return true
		}
		return false
	}

	return true
}

/*UpdateIndexOnTotalMatch increments the number of total matches for any SearchRequest from which
the SearchReply's filename could have originated.*/
func (timeout *SReqTotalMatch) UpdateIndexOnTotalMatch(filename string) {
	// Grab the mutex
	timeout.mux.Lock()
	defer timeout.mux.Unlock()

	/* Increment the number of total matches for any SearchRequest from which the
	SearchReply's filename could have originated. */
	for keywordsJoin, matches := range timeout.requests {
		keywords := strings.Split(keywordsJoin, ",")
		for _, k := range keywords {
			if strings.Contains(filename, k) {
				timeout.requests[keywordsJoin] = matches + 1
			}
		}
	}
}
