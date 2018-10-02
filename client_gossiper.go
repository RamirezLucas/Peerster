package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func callbackClient(g *Gossiper, udpChannel *net.UDPConn, pkt *GossipPacket) {

	// Print to the console
	fmt.Printf("%s, %v", pkt.simpleMsg.contents, g.peers)

	// Modify the packet
	pkt.simpleMsg.originalName = g.name
	pkt.simpleMsg.relayPeerAddr = g.gossipAddr

	// Create the packet
	buf, err := protobuf.Encode(*pkt)
	if err != nil {
		return
	}

	// Send to everyone
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	for _, peer := range g.peers.list {
		if _, err = udpChannel.WriteToUDP(buf, peer.udpAddr); err != nil {
			return
		}
	}

}
