package peers

import (
	"net"
)

// Peer - Represents a peer
type Peer struct {
	rawAddr string      // The peer's raw address <ip:port>
	udpAddr net.UDPAddr // The peer's UDP address
}

// NewPeer - Creates a new instance of Peer
func NewPeer(rawAddr string, udpAddr *net.UDPAddr) *Peer {
	return &Peer{rawAddr: rawAddr, udpAddr: *udpAddr}
}

// PeerToString - Returns the textual representation of a Peer
func (peer *Peer) PeerToString() string {
	return peer.rawAddr
}
