package types

import (
	"net"
	"sync"
)

// TimeoutHandler - Represents a StatusPacket answer to a RumorMessage
type TimeoutHandler struct {
	Addr *net.UDPAddr      // A peer's address
	Com  chan StatusPacket // A channel to communicate the status answer between threads
	Hash int               // A random number to identify the thread owning the timeout
	Done bool              // Indicated whether a packet was already forwarded using this handler
}

// StatusResponseForwarder - Represents the set of pending timeouts
type StatusResponseForwarder struct {
	Responses []TimeoutHandler // An array of timeout handlers
	Mux       sync.Mutex       // Mutex to manipulate the structure from different threads
}
