package network

import (
	"Peerster/entities"
	"Peerster/fail"
	"Peerster/messages"

	"github.com/dedis/protobuf"
)

// OnBroadcastClient -
func OnBroadcastClient(g *entities.Gossiper, simpleMsg *messages.SimpleMessage) {

	// Print to the console
	fail.LeveledPrint(0, "", "CLIENT MESSAGE %s", simpleMsg.Contents)
	fail.LeveledPrint(0, "", g.PeerIndex.PeersToString())

	// Modify the packet
	simpleMsg.OriginalName = g.Args.Name
	simpleMsg.RelayPeerAddr = g.Args.GossipAddr

	// Create the packet
	pkt := messages.GossipPacket{SimpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone
	g.PeerIndex.Broadcast(g.GossipChannel, buf, "")

}

// OnBroadcastNetwork -
func OnBroadcastNetwork(g *entities.Gossiper, simpleMsg *messages.SimpleMessage) {

	// Print to the console
	fail.LeveledPrint(0, "", simpleMsg.SimpleMessageToString())
	fail.LeveledPrint(0, "", g.PeerIndex.PeersToString())

	// Modify the structure
	sender := simpleMsg.RelayPeerAddr
	simpleMsg.RelayPeerAddr = g.Args.GossipAddr

	// Create the packet
	pkt := messages.GossipPacket{SimpleMsg: simpleMsg}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send to everyone (except the sender)
	g.PeerIndex.Broadcast(g.GossipChannel, buf, sender)

}
