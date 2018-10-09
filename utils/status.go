package utils

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnSendStatus -
func OnSendStatus(vectorClock *StatusPacket, channel *net.UDPConn, target *net.UDPAddr) error {

	// Create the packet
	pkt := GossipPacket{Status: vectorClock}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &CustomError{"OnSendStatus", "failed to encode GossipPacket"}
	}

	// Send the packet
	if _, err = channel.WriteToUDP(buf, target); err != nil {
		return &CustomError{"OnSendStatus", "failed to send StatusPacket"}
	}
	return nil
}

// OnReceiveStatus -
func OnReceiveStatus(g *Gossiper, status *StatusPacket, sender *net.UDPAddr) {

	g.Network.Mux.Lock()

	// Attempt to add the sending peer to the list of neighbors
	g.Network.AddPeerIfAbsent(sender)

	// Print to the console
	fmt.Printf("%s\n%s\n", StatusPacketToString(status, fmt.Sprintf("%s", sender)), PeersToString(g.Network.Peers))

	// See if we must propagate a rumor
	rumorToPropagate := g.Network.GetUnknownMessageTarget(status)

	if rumorToPropagate == nil { // We don't have anything to propagate
		if g.Network.IsLocalStatusComplete(status) { // We are in sync with the other
			fmt.Printf("IN SYNC WITH %s\n", fmt.Sprintf("%s", sender))
			g.Network.Mux.Unlock()
		} else { // We must send back our own Status
			vectorClock := g.Network.VectorClock
			g.Network.Mux.Unlock()
			OnSendStatus(&vectorClock, g.GossipChannel, sender)
		}
	} else { // We must propagate a rumor to the sender
		g.Network.Mux.Unlock()
		OnSendRumor(g, rumorToPropagate, sender)
	}

}
