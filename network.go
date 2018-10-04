package main

import (
	"fmt"
	"math/rand"
	"net"
)

// CreatePeer - Creates a Peer
func (p *Peer) CreatePeer(addr string) error {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return &CustomError{"CreatePeer", "cannot resolve UDP address"}
	}

	p.rawAddr = addr
	p.udpAddr = udpAddr

	return nil
}

// AddNamedPeer - Adds a named peer to the network
func (network *GossipNetwork) AddNamedPeer(name string) {
	newPeer := NamedPeer{name: name}
	newPeerStatus := PeerStatus{name, 1}
	network.history = append(network.history, newPeer)
	network.vectorClock.want = append(network.vectorClock.want, newPeerStatus)
}

// AddMessageIfNext - Adds a message to the GossipNetwork given that we got all the preceding ones
func (network *GossipNetwork) AddMessageIfNext(r *RumorMessage) bool {
	for i, namedPeer := range network.history { // Find the named peer in the list
		if namedPeer.name == r.origin { // Match !
			if r.id == network.vectorClock.want[i].nextID { // Check the message order
				namedPeer.messages = append(namedPeer.messages, r.text) // Append the message to the list
				network.vectorClock.want[i].nextID++                    // Update the vector clock
				return true
			}
			return false
		}
	}

	// Could not find the named peer in the list: add it
	network.AddNamedPeer(r.origin)
	networkSize := len(network.history) - 1
	network.history[networkSize].messages = append(network.history[networkSize].messages, r.text)
	network.vectorClock.want[networkSize].nextID++
	return true
}

// GetMessage - Get a message stored in the GossipNetwork
func (network *GossipNetwork) GetMessage(name string, id uint32) (string, error) {
	for i, namedPeer := range network.history { // Find the named peer in the list
		if namedPeer.name == name { // Match !
			if id < network.vectorClock.want[i].nextID {
				return namedPeer.messages[id-1], nil
			}
			return "", &CustomError{"GetMessage", "index out of range"}
		}
	}
	return "", &CustomError{"GetMessage", "could not find specified peer"}
}

// GetLastMessageID - Get the next message expected IF for a given named peer
func (network *GossipNetwork) GetLastMessageID(name string) uint32 {
	for i, namedPeer := range network.history { // Find the named peer in the list
		if namedPeer.name == name { // Match !
			return network.vectorClock.want[i].nextID
		}
	}
	return 0
}

// AddPeerIfAbsent - Add a peer to the list if it is absent from it
func (network *GossipNetwork) AddPeerIfAbsent(newPeerAddr *net.UDPAddr) {
	for _, peer := range network.peers { // Iterate over the know peers
		if CompareUDPAddress(peer.udpAddr, newPeerAddr) { // Match !
			return
		}
	}

	// We must add the peer to the list
	newPeer := Peer{fmt.Sprintf("%s", newPeerAddr), newPeerAddr}
	network.peers = append(network.peers, newPeer)
}

// GetRandomPeer - Get a random peer from a list, excluding one
func (network *GossipNetwork) GetRandomPeer(excludeMe *net.UDPAddr) *net.UDPAddr {

	nbPeers := len(network.peers)

	// Check that the list isn't empty
	if nbPeers == 0 {
		return nil
	}

	// Check that the excludeMe peer is not the only one in the list
	if excludeMe != nil && nbPeers == 1 && CompareUDPAddress(excludeMe, network.peers[0].udpAddr) {
		return nil
	}

	/* This for loop ensures that we select a peer uniformy at random (as far as
	our pseudo-random generator goes) while preventing the selection of the
	"excludeMe" peer */
	for {
		randomPeer := network.peers[rand.Intn(nbPeers)].udpAddr
		if excludeMe == nil || !CompareUDPAddress(excludeMe, randomPeer) {
			return randomPeer
		}
	}
}
