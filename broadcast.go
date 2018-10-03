package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnBroadcastClient -
func OnBroadcastClient(g *Gossiper, channel *net.UDPConn, simpleMsg *SimpleMessage) {

	// Print to the console
	g.mux.Lock()
	fmt.Printf("CLIENT MESSAGE %s\n%s\n", simpleMsg.contents, PeersToString(g.network.peers))
	g.mux.Unlock()

	// Modify the packet
	simpleMsg.originalName = g.name
	simpleMsg.relayPeerAddr = g.gossipAddr

	// Create the packet
	pkt := GossipPacket{simpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	for _, peer := range g.network.peers {
		if _, err = channel.WriteToUDP(buf, peer.udpAddr); err != nil {
			return
		}
	}

}

// OnBroadcastNetwork -
func OnBroadcastNetwork(g *Gossiper, channel *net.UDPConn, simpleMsg *SimpleMessage) {

	// Print to the console
	g.mux.Lock()
	fmt.Printf("%s\n%s\n", SimpleMessageToString(simpleMsg), PeersToString(g.network.peers))
	g.mux.Unlock()

	// Modify the packet
	sender := simpleMsg.relayPeerAddr
	simpleMsg.relayPeerAddr = g.gossipAddr

	// Create the packet
	pkt := GossipPacket{simpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone (except the sender)
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	isPeerKnown := false
	for _, peer := range g.network.peers {
		if sender == peer.rawAddr {
			isPeerKnown = true
		} else {
			if _, err = channel.WriteToUDP(buf, peer.udpAddr); err != nil {
				return
			}
		}
	}

	if !isPeerKnown { // We need to add the sender to the peers list
		var newPeer Peer
		if err := newPeer.CreatePeer(sender); err != nil {
			g.network.peers = append(g.network.peers, newPeer)
		}
	}

}
