package types

import (
	"fmt"
	"net"
	"os"
	"sync"
)

// RoutingTable - Represents a routing table between peers name and next hop <ip:port>
type RoutingTable struct {
	table map[string]*NextHop // A mapping from peer name to <ip:port>
	mux   sync.Mutex          // Mutex to manipulate the structure from different threads
}

// NextHop - Represents a next hop along a route
type NextHop struct {
	nextPeer     *Peer  // The next peer along the route
	lastUpdateID uint32 // The last ID used to update the route
}

// NewRoutingTable - Creates a new instance of RoutingTable
func NewRoutingTable() *RoutingTable {
	var routing RoutingTable
	routing.table = make(map[string]*NextHop)
	return &routing
}

// NewNextHop - Creates a new instance of NextHop
func NewNextHop(rawAddr string, udpAddr *net.UDPAddr, updateID uint32) *NextHop {
	var nextHop NextHop
	nextHop.nextPeer = NewPeer(rawAddr, udpAddr)
	nextHop.lastUpdateID = updateID
	return &nextHop
}

// AddContactIfAbsent - Adds a new contact to the routing table if it doesn't exist yet
func (routing *RoutingTable) AddContactIfAbsent(name string, sender *net.UDPAddr) {
	routing.mux.Lock()
	defer routing.mux.Unlock()

	if _, ok := routing.table[name]; !ok { // We don't know the sender

		// Add the contact
		addrStr := UDPAddressToString(sender)
		routing.table[name] = NewNextHop(addrStr, sender, 0)

	}
}

// UpdateTableAndPrint - Updates the table with a new/updated record and prints it
func (routing *RoutingTable) UpdateTableAndPrint(name string, sender *net.UDPAddr, updateID uint32) {
	routing.mux.Lock()
	defer routing.mux.Unlock()

	addrStr := UDPAddressToString(sender)

	if nextHop, ok := routing.table[name]; !ok { // We don't know the sender
		routing.table[name] = NewNextHop(addrStr, sender, updateID)
		fmt.Printf("%s\n", routing.RouterEntryToStringUnsafe(name))

		// Send the new name to the server
		FBuffer.AddFrontendPrivateContact(name)

	} else { // We know the sender

		// Update the route if the sequence ID is higher or equal
		if updateID >= nextHop.lastUpdateID && nextHop.nextPeer.rawAddr != addrStr {
			routing.table[name] = NewNextHop(addrStr, sender, updateID)
			fmt.Printf("%s\n", routing.RouterEntryToStringUnsafe(name))
		}

	}

}

// GetTarget - Get the next-hop target in the routing table for a particular destination
func (routing *RoutingTable) GetTarget(name string) *net.UDPAddr {

	if nextHop, ok := routing.table[name]; ok { // We have a next-hop
		return &nextHop.nextPeer.udpAddr
	}

	return nil
}

// RouterEntryToStringUnsafe - Returns the textual representation of a router entry
func (routing *RoutingTable) RouterEntryToStringUnsafe(name string) string {

	if nextHop, ok := routing.table[name]; ok { // We know the sender
		return fmt.Sprintf("DSDV %s %s", name, nextHop.nextPeer.PeerToString())
	}

	fmt.Printf("ERROR: Trying to print non-existent entry %s in the routing table\n", name)
	os.Exit(1)
	return ""
}
