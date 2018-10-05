package utils

import (
	"fmt"
	"math/rand"
	"net"
)

// CompareUDPAddress - Compares 2 UDP addresses
func CompareUDPAddress(a, b *net.UDPAddr) bool {
	if a.Port == b.Port {
		for i := 0; i < 4; i++ {
			if a.IP[i] != b.IP[i] {
				return false
			}
		}
		return true
	}
	return false
}

// CreatePeer - Creates a Peer
func (p *Peer) CreatePeer(addr string) error {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return &CustomError{"CreatePeer", "cannot resolve UDP address"}
	}

	p.RawAddr = addr
	p.UdpAddr = udpAddr

	return nil
}

// AddNamedPeer - Adds a named peer to the network
func (network *GossipNetwork) AddNamedPeer(name string) {
	newPeer := NamedPeer{Name: name}
	newPeerStatus := PeerStatus{name, 1}
	network.History = append(network.History, newPeer)
	network.VectorClock.Want = append(network.VectorClock.Want, newPeerStatus)
}

// AddMessageIfNext - Adds a message to the GossipNetwork given that we got all the preceding ones
func (network *GossipNetwork) AddMessageIfNext(r *RumorMessage) bool {
	for i, namedPeer := range network.History { // Find the named peer in the list
		if namedPeer.Name == r.Origin { // Match !
			if r.ID == network.VectorClock.Want[i].NextID { // Check the message order
				namedPeer.Messages = append(namedPeer.Messages, r.Text) // Append the message to the list
				network.VectorClock.Want[i].NextID++                    // Update the vector clock
				return true
			}
			return false
		}
	}

	// Could not find the named peer in the list: add it
	network.AddNamedPeer(r.Origin)
	networkSize := len(network.History) - 1
	network.History[networkSize].Messages = append(network.History[networkSize].Messages, r.Text)
	network.VectorClock.Want[networkSize].NextID++
	return true
}

// GetMessage - Get a message stored in the GossipNetwork
func (network *GossipNetwork) GetMessage(name string, id uint32) (string, error) {
	for i, namedPeer := range network.History { // Find the named peer in the list
		if namedPeer.Name == name { // Match !
			if id < network.VectorClock.Want[i].NextID {
				return namedPeer.Messages[id-1], nil
			}
			return "", &CustomError{"GetMessage", "index out of range"}
		}
	}
	return "", &CustomError{"GetMessage", "could not find specified peer"}
}

// GetLastMessageID - Get the next message expected IF for a given named peer
func (network *GossipNetwork) GetLastMessageID(name string) uint32 {
	for i, namedPeer := range network.History { // Find the named peer in the list
		if namedPeer.Name == name { // Match !
			return network.VectorClock.Want[i].NextID
		}
	}
	return 0
}

// GetUnknownMessageTarget - Tries to find a message that we have but that the other doesn't
func (network *GossipNetwork) GetUnknownMessageTarget(status *StatusPacket) *RumorMessage {

	for i, knownPeer := range network.VectorClock.Want {
		indexPeer := status.IsPeerInStatus(knownPeer.Identifier)
		if indexPeer != -1 { // We both know the peer
			nextIDWanted := status.Want[indexPeer].NextID
			if 0 < nextIDWanted && nextIDWanted < knownPeer.NextID { // We have something the other doesn't have
				rumor := RumorMessage{knownPeer.Identifier, nextIDWanted, network.History[i].Messages[nextIDWanted-1]}
				return &rumor
			}
		} else if knownPeer.NextID != 1 { // The other doesn't know the peer, we have at least one message from him
			rumor := RumorMessage{knownPeer.Identifier, 1, network.History[i].Messages[0]}
			return &rumor
		}
	}
	return nil
}

// IsLocalStatusComplete - Checks whether we are not aware of some message that was received by the other
func (network *GossipNetwork) IsLocalStatusComplete(status *StatusPacket) bool {
	for _, distantPeer := range status.Want {
		indexPeer := network.VectorClock.IsPeerInStatus(distantPeer.Identifier)
		if indexPeer != -1 { // We know the peer
			nextIDWanted := network.VectorClock.Want[indexPeer].NextID
			if nextIDWanted < distantPeer.NextID { // The other has something we don't have
				return false
			}
		} else { // We don't know the peer
			return false
		}
	}
	return true
}

// IsPeerInStatus - Checks whether a peer with a given name exists in the StatusPacket
func (status *StatusPacket) IsPeerInStatus(name string) int {
	for i, peer := range status.Want {
		if peer.Identifier == name {
			return i
		}
	}
	return -1
}

// AddPeerIfAbsent - Add a peer to the list if it is absent from it
func (network *GossipNetwork) AddPeerIfAbsent(newPeerAddr *net.UDPAddr) {
	for _, peer := range network.Peers { // Iterate over the know peers
		if CompareUDPAddress(peer.UdpAddr, newPeerAddr) { // Match !
			return
		}
	}

	// We must add the peer to the list
	newPeer := Peer{fmt.Sprintf("%s", newPeerAddr), newPeerAddr}
	network.Peers = append(network.Peers, newPeer)
}

// GetRandomPeer - Get a random peer from a list, excluding one
func (network *GossipNetwork) GetRandomPeer(excludeMe *net.UDPAddr) *net.UDPAddr {

	nbPeers := len(network.Peers)

	// Check that the list isn't empty
	if nbPeers == 0 {
		return nil
	}

	// Check that the excludeMe peer is not the only one in the list
	if excludeMe != nil && nbPeers == 1 && CompareUDPAddress(excludeMe, network.Peers[0].UdpAddr) {
		return nil
	}

	/* This for loop ensures that we select a peer uniformy at random (as far as
	our pseudo-random generator goes) while preventing the selection of the
	"excludeMe" peer */
	for {
		randomPeer := network.Peers[rand.Intn(nbPeers)].UdpAddr
		if excludeMe == nil || !CompareUDPAddress(excludeMe, randomPeer) {
			return randomPeer
		}
	}
}
