package types

import (
	"net"
	"sync"
)

// RoutingTable - Represents a routing table between peers name and next hop <ip:port>
type RoutingTable struct {
	table map[string]*Peer // A mapping from peer name to <ip:port>
	mux   sync.Mutex       // Mutex to manipulate the structure from different threads
}

// NewRoutingTable - Creates a new instance of RoutingTable
func NewRoutingTable() *RoutingTable {
	var routing RoutingTable
	routing.table = make(map[string]*Peer)
	return &routing
}

// UpdateTable - Updates the table with a new/updated record
func (routing *RoutingTable) UpdateTable(name string, sender *net.UDPAddr) {
	routing.mux.Lock()
	defer routing.mux.Unlock()

	addrStr := UDPAddressToString(sender)

	if _, ok := routing.table[name]; !ok { // We don't know the sender
		routing.table[name] = NewPeer(addrStr, sender)

		// Add new peer to server buffer
		// slices := strings.Split(addrStr, ":")
		// BufferPeers.AddServerPeer(slices[0], slices[1])

	} else { // We know the sender

		routing.table[name] = NewPeer(addrStr, sender)

	}
}
