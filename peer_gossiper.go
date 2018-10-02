package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func callbackPeer(udpChannel *net.UDPConn, g *Gossiper, pkt *GossipPacket) error {

	// Print the message on standard output
	fmt.Println()

	// Modify the packet
	sender := pkt.msg.relayPeerAddr
	pkt.msg.relayPeerAddr = g.gossipAddr

	// Create the packet
	buf, err := protobuf.Encode(*pkt)
	if err != nil {
		return &CustomError{"callbackPeer", "failed to encode packet"}
	}

	// Send to everyone (except the sender)
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	isPeerKnown := false
	for _, peer := range g.peers {
		if sender == peer {
			isPeerKnown = true
		} else {
			// TODO: remove code copy
			udpAddr, err := net.ResolveUDPAddr("udp4", peer)
			if err != nil {
				return &CustomError{"callbackPeer", "unable to resolve UDP address"}
			}
			if _, err = udpChannel.WriteToUDP(buf, udpAddr); err != nil {
				return &CustomError{"callbackPeer", "unable to write on UDP channel"}
			}
		}
	}
	if !isPeerKnown { // We need to add the sender to the peers list
		g.peers = append(g.peers, sender)
	}

	return nil
}
