package peers

import (
	"Peerster/frontend"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
)

// PeerIndex represents a dictionnary between <ip:port> and peer addresses
type PeerIndex struct {
	index map[string]*Peer // A mapping from peer <ip:port> to Peer
	mux   sync.Mutex       // Mutex to manipulate the structure from different threads
}

// NewPeerIndex creates a new instance of PeerIndex
func NewPeerIndex() *PeerIndex {
	var peerIndex PeerIndex
	peerIndex.index = make(map[string]*Peer)
	return &peerIndex
}

// Broadcast sends a packet to every neighbor, possible exluding one
func (peerIndex *PeerIndex) Broadcast(channel *net.UDPConn, buf []byte, excludeMe string) {
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	for addr, peer := range peerIndex.index {
		if addr != excludeMe {
			channel.WriteToUDP(buf, &peer.udpAddr)
		}
	}

}

// AddPeerIfAbsent adds a peer to the index if it doesn't exist yet
func (peerIndex *PeerIndex) AddPeerIfAbsent(newPeerAddr *net.UDPAddr) {
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	addrStr := UDPAddressToString(newPeerAddr)

	if _, ok := peerIndex.index[addrStr]; !ok { // We don't know this peer
		peerIndex.index[addrStr] = NewPeer(addrStr, newPeerAddr)

		// Add new peer to server buffer
		slices := strings.Split(addrStr, ":")
		frontend.FBuffer.AddFrontendPeer(slices[0], slices[1])
	}
}

// GetRandomPeer gets a random neighbor from the index, possibly excluding one
func (peerIndex *PeerIndex) GetRandomPeer(excludeMe *net.UDPAddr) *net.UDPAddr {
	if ret := peerIndex.GetRandomNeighbors(1, excludeMe); ret != nil {
		return ret[0]
	}
	return nil
}

// GetRandomNeighbors returns a maximum of nb randomly picked neighbors among the current list,
// possibly exluding one if excludeMe represents a valid UDP address. If there aren't any neighbors
// the function returns nil. If nb is bigger that the number of neighbors all neighbors except the
// excludeMe one are returned.
func (peerIndex *PeerIndex) GetRandomNeighbors(nbMax int, excludeMe *net.UDPAddr) []*net.UDPAddr {

	if nbMax == 0 {
		return nil
	}

	// Grab the mutex
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	// Get current number of peers
	nbPeers := len(peerIndex.index)
	if nbPeers == 0 {
		return nil
	}

	// Create temporary list of peers
	peersList := make([]*net.UDPAddr, 0)
	i := 0
	excludeStr := ""
	if excludeMe != nil {
		excludeStr = UDPAddressToString(excludeMe)
	}
	for addr, peer := range peerIndex.index {
		if addr != excludeStr {
			peersList = append(peersList, &peer.udpAddr)
			i++
		}
	}

	// Check if the list is empty
	if len(peersList) == 0 {
		return nil
	}

	// If the list is smaller than the required number return everything
	if nbMax >= len(peersList) {
		return peersList
	}

	// Pick nbMax neighbors randomly
	retList := make([]*net.UDPAddr, 0)
	pickedNeighbors := make(map[int]bool, nbMax)
	for i := 0; i < nbMax; i++ {
		for {
			rNumber := rand.Intn(nbMax)
			if _, ok := pickedNeighbors[rNumber]; !ok { // We never picked this one
				retList = append(retList, peersList[rNumber])
				pickedNeighbors[rNumber] = true
				break
			}
		}
	}

	return retList
}

// PeersToString returns a textual representation of a peer index
func (peerIndex *PeerIndex) PeersToString() string {
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	s := "PEERS "
	for addr := range peerIndex.index {
		s = s + fmt.Sprintf("%s,", addr)
	}
	return s[:len(s)-1]
}
