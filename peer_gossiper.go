package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func callbackPeer(g *Gossiper, udpChannel *net.UDPConn, pkt *GossipPacket) {

	// Print to the console
	fmt.Printf("%v, %v", pkt.simpleMsg, g.peers)

	// Modify the packet
	sender := pkt.simpleMsg.relayPeerAddr
	pkt.simpleMsg.relayPeerAddr = g.gossipAddr

	// Create the packet
	buf, err := protobuf.Encode(*pkt)
	if err != nil {
		return
	}

	// Send to everyone (except the sender)
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	isPeerKnown := false
	for _, peer := range g.peers.list {
		if sender == peer.rawAddr {
			isPeerKnown = true
		} else {
			if _, err = udpChannel.WriteToUDP(buf, peer.udpAddr); err != nil {
				return
			}
		}
	}

	if !isPeerKnown { // We need to add the sender to the peers list
		var newPeer Peer
		if err := newPeer.CreatePeer(sender); err != nil {
			g.peers.list = append(g.peers.list, newPeer)
		}
	}

}
