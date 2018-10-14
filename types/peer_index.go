package types

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
)

// PeerIndex - Represents a dictionnary between <ip:port> and peer addresses
type PeerIndex struct {
	index map[string]*Peer // A mapping from peer <ip:port> to Peer
	mux   sync.Mutex       // Mutex to manipulate the structure from different threads
}

// Peer - Represents a peer
type Peer struct {
	udpAddr net.UDPAddr // The peer's UDP address
}

// NewPeerIndex - Creates a new instance of PeerIndex
func NewPeerIndex() *PeerIndex {
	var peerIndex PeerIndex
	peerIndex.index = make(map[string]*Peer)
	return &peerIndex
}

// Broadcast - Send a packet to everyone, possible exluding one peer
func (peerIndex *PeerIndex) Broadcast(channel *net.UDPConn, buf []byte, excludeMe string) {
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	for addr, peer := range peerIndex.index {
		if addr != excludeMe {
			channel.WriteToUDP(buf, &peer.udpAddr)
		}
	}

}

// AddPeerIfAbsent - Add a peer to the index if it doesn't yet exists
func (peerIndex *PeerIndex) AddPeerIfAbsent(newPeerAddr *net.UDPAddr) {
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	addrStr := UDPAddressToString(newPeerAddr)

	// Prevents the client from talking on the network port
	if addrStr == ":0" {
		return
	}

	if _, ok := peerIndex.index[addrStr]; !ok { // We don't know this peer
		peerIndex.index[addrStr] = &Peer{udpAddr: *newPeerAddr}

		// Add new peer to server buffer
		slices := strings.Split(addrStr, ":")
		if port, err := strconv.ParseInt(slices[1], 10, 32); err == nil {
			BufferPeers.AddServerPeer(slices[0], int(port))
		}
	}
}

// GetRandomPeer - Get a random peer from the index, possibly excluding one
func (peerIndex *PeerIndex) GetRandomPeer(excludeMe *net.UDPAddr) *net.UDPAddr {
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	// Get number of known peers
	nbPeers := len(peerIndex.index)

	// Check that the list isn't empty
	if nbPeers == 0 {
		return nil
	}

	// Temporary variables
	tmp := make([]*Peer, nbPeers)
	i := 0

	if excludeMe != nil {

		excludeStr := UDPAddressToString(excludeMe)

		// Check that the excludeMe peer is not the only one in the list
		if nbPeers == 1 {
			if _, ok := peerIndex.index[excludeStr]; ok {
				return nil
			}
			for _, peer := range peerIndex.index {
				return &peer.udpAddr
			}

		}

		for addr, peer := range peerIndex.index {
			if addr != excludeStr {
				tmp[i] = peer
				i++
			}
		}

		// Pick a random peer in the list
		randomIndex := rand.Intn(nbPeers - 1)
		return &tmp[randomIndex].udpAddr

	}

	// == Normal case ==
	for _, peer := range peerIndex.index {
		tmp[i] = peer
		i++
	}

	// Pick a random peer in the list
	randomIndex := rand.Intn(nbPeers)
	return &tmp[randomIndex].udpAddr

}

// PeersToString - Returns a textual representation of a peer index
func (peerIndex *PeerIndex) PeersToString() string {
	peerIndex.mux.Lock()
	defer peerIndex.mux.Unlock()

	s := "PEERS "
	for addr := range peerIndex.index {
		s = s + fmt.Sprintf("%s,", addr)
	}
	return s[:len(s)-1]
}
