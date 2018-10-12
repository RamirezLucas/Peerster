package types

import (
	"fmt"
	"net"
	"os"
	"sync"
)

// StatusResponseForwarder - Represents the set of pending timeouts
type StatusResponseForwarder struct {
	responses map[uint32]*TimeoutHandler // An index of timeout handlers
	mux       sync.Mutex                 // Mutex to manipulate the structure from different threads
}

// TimeoutHandler - Represents a StatusPacket answer to a RumorMessage
type TimeoutHandler struct {
	addr net.UDPAddr        // A peer's address
	com  chan *StatusPacket // A channel to communicate the status answer between threads
	done bool               // Indicated whether a packet was already forwarded using this handler
}

// NewStatusResponseForwarder - Creates a new instance of StatusResponseForwarder
func NewStatusResponseForwarder() *StatusResponseForwarder {
	var forwarder StatusResponseForwarder
	forwarder.responses = make(map[uint32]*TimeoutHandler)
	return &forwarder
}

// NewTimeoutHandler - Creates a new instance of TimeoutHandler
func NewTimeoutHandler(udpAddr *net.UDPAddr) *TimeoutHandler {
	var handler TimeoutHandler
	handler.addr = *udpAddr
	handler.com = make(chan *StatusPacket, 1)
	return &handler
}

// AddTimeoutHandler - Adds a new timeout handler to the forwarder
func (forwarder *StatusResponseForwarder) AddTimeoutHandler(threadID uint32, sender *net.UDPAddr) {
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	if _, ok := forwarder.responses[threadID]; !ok {
		forwarder.responses[threadID] = NewTimeoutHandler(sender)
	} else {
		fmt.Printf("ERROR: Trying to add existing threadID %d to the forwarder", threadID)
		os.Exit(1)
	}
}

// DeleteTimeoutHandler - Deletes a timeout handler from the forwarder (last chance pickup)
func (forwarder *StatusResponseForwarder) DeleteTimeoutHandler(threadID uint32) *StatusPacket {
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	if handler, ok := forwarder.responses[threadID]; ok {

		var status *StatusPacket

		// Last chance to look at the channel
		select {
		case status = <-handler.com:
		default: // Do nothing
		}

		close(handler.com)
		delete(forwarder.responses, threadID)
		return status
	}

	fmt.Printf("ERROR: Trying to delete non-existing threadID %d from the forwarder", threadID)
	os.Exit(1)
	return nil
}

// SearchAndForward - Searches the list of handlers for a given sender address. Forwards the packet on match
func (forwarder *StatusResponseForwarder) SearchAndForward(sender *net.UDPAddr, status *StatusPacket) bool {
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	minThreadID := uint32(0)

	for threadID, handler := range forwarder.responses {
		if !handler.done && CompareUDPAddress(&handler.addr, sender) { // Match!
			if minThreadID == 0 || threadID < minThreadID {
				minThreadID = threadID // Find the "oldest" thread
			}
		}
	}

	if minThreadID != 0 { // We got a match!
		handler := forwarder.responses[minThreadID]
		handler.com <- status
		handler.done = true
		return true
	}

	return false
}

// LookForData - Look for data on the channel dedicated to a particular thread
func (forwarder *StatusResponseForwarder) LookForData(threadID uint32) *StatusPacket {
	forwarder.mux.Lock()
	defer forwarder.mux.Unlock()

	// TODO: Maybe locking is needed here
	if handler, ok := forwarder.responses[threadID]; ok {
		select {
		case response := <-handler.com:
			return response
		default:
			return nil
		}
	} else {
		fmt.Printf("ERROR: Trying to look for data on non-existing timeut handler. ID: %d", threadID)
		os.Exit(1)
	}

	return nil
}
