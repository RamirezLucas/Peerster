package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func callbackClient(udpChannel *net.UDPConn, g *Gossiper, pkt *GossipPacket) error {

	// Print the message on standard output
	fmt.Println(pkt.msg)

	// Modify the packet
	pkt.msg.originalName = g.name
	pkt.msg.relayPeerAddr = g.gossipAddr

	// Create the packet
	buf, err := protobuf.Encode(*pkt)
	if err != nil {
		return &CustomError{"callbackClient", "failed to encode packet"}
	}

	// Send to everyone
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	for _, peer := range g.peers {
		// TODO: remove code copy
		udpAddr, err := net.ResolveUDPAddr("udp4", peer)
		if err != nil {
			return &CustomError{"callbackClient", "unable to resolve UDP address"}
		}
		if _, err = udpChannel.WriteToUDP(buf, udpAddr); err != nil {
			return &CustomError{"callbackClient", "unable to write on UDP channel"}
		}
	}

	return nil
}
