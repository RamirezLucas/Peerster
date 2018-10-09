package network

import (
	"Peerster/types"
	"Peerster/utils"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnBroadcastClient -
func OnBroadcastClient(g *types.Gossiper, simpleMsg *types.SimpleMessage) {

	// Print to the console
	g.Network.Mux.Lock()
	fmt.Printf("CLIENT MESSAGE %s\n%s\n", simpleMsg.Contents, utils.PeersToString(g.Network.Peers))
	g.Network.Mux.Unlock()

	// Modify the packet
	simpleMsg.OriginalName = g.Name
	simpleMsg.RelayPeerAddr = g.GossipAddr

	// Create the packet
	pkt := types.GossipPacket{SimpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone
	g.Network.Mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.Network.Mux.Unlock()

	for _, peer := range g.Network.Peers {
		if _, err = g.GossipChannel.WriteToUDP(buf, peer.UDPAddr); err != nil {
			return
		}
	}

}

// OnBroadcastNetwork -
func OnBroadcastNetwork(g *types.Gossiper, simpleMsg *types.SimpleMessage) {

	// Print to the console
	g.Network.Mux.Lock()

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", simpleMsg.RelayPeerAddr)
	if err != nil {
		return
	}
	g.Network.AddPeerIfAbsent(udpAddr)

	fmt.Printf("%s\n%s\n", utils.SimpleMessageToString(simpleMsg), utils.PeersToString(g.Network.Peers))
	g.Network.Mux.Unlock()

	// Modify the packet
	sender := simpleMsg.RelayPeerAddr
	simpleMsg.RelayPeerAddr = g.GossipAddr

	// Create the packet
	pkt := types.GossipPacket{SimpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone (except the sender)
	g.Network.Mux.Lock() // Lock the gossiper because we are accessing peers
	for _, peer := range g.Network.Peers {
		if sender != peer.RawAddr {
			if _, err = g.GossipChannel.WriteToUDP(buf, peer.UDPAddr); err != nil {
				return
			}
		}
	}
	g.Network.Mux.Unlock()

}
