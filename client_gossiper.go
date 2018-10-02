package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func callbackClient(g *Gossiper, udpChannel *net.UDPConn, simpleMsg *SimpleMessage) {

	// Print to the console
	g.mux.Lock()
	fmt.Printf("%s%v", simpleMsg.contents, g.peers)
	g.mux.Unlock()

	// Modify the packet
	simpleMsg.originalName = g.name
	simpleMsg.relayPeerAddr = g.gossipAddr

	// Create the packet
	pkt := GossipPacket{simpleMsg: simpleMsg}
	buf, err := protobuf.Encode(pkt)
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
