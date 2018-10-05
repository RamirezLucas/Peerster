package utils

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnBroadcastClient -
func OnBroadcastClient(g *Gossiper, channel *net.UDPConn, simpleMsg *SimpleMessage) {

	// Print to the console
	g.Network.Mux.Lock()
	fmt.Printf("CLIENT MESSAGE %s\n%s\n", simpleMsg.Contents, PeersToString(g.Network.Peers))
	g.Network.Mux.Unlock()

	// Modify the packet
	simpleMsg.OriginalName = g.Name
	simpleMsg.RelayPeerAddr = g.GossipAddr

	// Create the packet
	pkt := GossipPacket{SimpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone
	g.Network.Mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.Network.Mux.Unlock()

	for _, peer := range g.Network.Peers {
		if _, err = channel.WriteToUDP(buf, peer.UdpAddr); err != nil {
			return
		}
	}

}

// OnBroadcastNetwork -
func OnBroadcastNetwork(g *Gossiper, channel *net.UDPConn, simpleMsg *SimpleMessage) {

	// Print to the console
	g.Network.Mux.Lock()
	fmt.Printf("%s\n%s\n", SimpleMessageToString(simpleMsg), PeersToString(g.Network.Peers))
	g.Network.Mux.Unlock()

	// Modify the packet
	sender := simpleMsg.RelayPeerAddr
	simpleMsg.RelayPeerAddr = g.GossipAddr

	// Create the packet
	pkt := GossipPacket{SimpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone (except the sender)
	g.Network.Mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.Network.Mux.Unlock()

	isPeerKnown := false
	for _, peer := range g.Network.Peers {
		if sender == peer.RawAddr {
			isPeerKnown = true
		} else {
			if _, err = channel.WriteToUDP(buf, peer.UdpAddr); err != nil {
				return
			}
		}
	}

	if !isPeerKnown { // We need to add the sender to the peers list
		var newPeer Peer
		if err := newPeer.CreatePeer(sender); err != nil {
			g.Network.Peers = append(g.Network.Peers, newPeer)
		}
	}

}
