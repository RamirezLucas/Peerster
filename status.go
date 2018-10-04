package main

import (
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
			g.mux.Lock()
			target := g.network.GetRandomPeer(nil)
			g.mux.Unlock()
			if target != nil {
				OnSendStatus(g, channel, target)
			}
		}
	}
}

// OnSendStatus -
func OnSendStatus(g *Gossiper, channel *net.UDPConn, target *net.UDPAddr) error {

	// Retreive our vector clock
	g.mux.Lock()
	vectorClock := g.network.vectorClock
	g.mux.Unlock()

	// Create the packet
	pkt := GossipPacket{status: &vectorClock}
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
func OnReceiveStatus(g *Gossiper, channel *net.UDPConn, status *StatusPacket, sender *net.UDPAddr, target *net.UDPAddr) {

}
