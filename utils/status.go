package utils

import (
	"fmt"
	"net"
	"time"

	"github.com/dedis/protobuf"
)

// AntiEntropy -
func AntiEntropy(g *Gossiper, channel *net.UDPConn) {

	// Create a timeout timer
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			// Pick a random target and send a StatusPacket
			g.Network.Mux.Lock()
			target := g.Network.GetRandomPeer(nil)
			vectorClock := g.Network.VectorClock
			g.Network.Mux.Unlock()
			if target != nil {
				OnSendStatus(&vectorClock, channel, target)
			}
		}
	}
}

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
func OnReceiveStatus(g *Gossiper, channel *net.UDPConn, status *StatusPacket, sender *net.UDPAddr) {

	g.Network.Mux.Lock()

	// Print to the console
	fmt.Printf("%s\n%s\n", StatusPacketToString(status, fmt.Sprintf("%s", sender)), PeersToString(g.Network.Peers))

	rumorToPropagate := g.Network.GetUnknownMessageTarget(status)

	if rumorToPropagate == nil { // We don't have anything to propagate
		if g.Network.IsLocalStatusComplete(status) { // We are in sync with the other
			fmt.Printf("IN SYNC WITH %s\n", fmt.Sprintf("%s", sender))
			g.Network.Mux.Unlock()
		} else { // We must send back our own Status
			vectorClock := g.Network.VectorClock
			g.Network.Mux.Unlock()
			OnSendStatus(&vectorClock, channel, sender)
		}
	} else { // We must propagate a rumor to the sender
		g.Network.Mux.Unlock()
		OnSendRumor(g, rumorToPropagate, channel, sender)
	}

}
