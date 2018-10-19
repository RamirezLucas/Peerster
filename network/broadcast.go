package network

import (
	"Peerster/types"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnBroadcastClient -
func OnBroadcastClient(g *types.Gossiper, simpleMsg *types.SimpleMessage) {

	// Print to the console
	fmt.Printf("CLIENT MESSAGE %s\n%s\n", simpleMsg.Contents, g.PeerIndex.PeersToString())

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
	g.PeerIndex.Broadcast(g.GossipChannel, buf, "")

}

// OnBroadcastNetwork -
func OnBroadcastNetwork(g *types.Gossiper, simpleMsg *types.SimpleMessage) {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", simpleMsg.RelayPeerAddr)
	if err != nil {
		return
	}

	// Prevents the client from talking on the network port
	if simpleMsg.RelayPeerAddr == "" {
		return
	}

	// Add the peer to the index
	g.PeerIndex.AddPeerIfAbsent(udpAddr)

	// Print to the console
	fmt.Printf("%s\n%s\n", simpleMsg.SimpleMessageToString(), g.PeerIndex.PeersToString())

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
	g.PeerIndex.Broadcast(g.GossipChannel, buf, sender)

}
