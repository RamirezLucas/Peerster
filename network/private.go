package network

import (
	"Peerster/entities"
	"Peerster/messages"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnSendPrivate - Sends a private message
func OnSendPrivate(g *entities.Gossiper, private *messages.PrivateMessage, target *net.UDPAddr) {

	// Create the packet
	pkt := messages.GossipPacket{Private: private}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send the packet
	g.GossipChannel.WriteToUDP(buf, target)
}

// OnReceiveClientPrivate - Called when a private message is received from the client
func OnReceiveClientPrivate(g *entities.Gossiper, private *messages.PrivateMessage) {

	// Fill in remaining fields
	private.Origin = g.Args.Name
	private.ID = 0
	private.HopLimit = 16

	// Add the message
	g.NameIndex.AddPrivateMessage(private)

	// Pick the target (should exist) and send
	if target := g.Router.GetTarget(private.Destination); target != nil {
		OnSendPrivate(g, private, target)
	}

}

// OnReceivePrivate - Called when a private message is received
func OnReceivePrivate(g *entities.Gossiper, private *messages.PrivateMessage, sender *net.UDPAddr) {

	// Update the routing table for private messages
	if private.Origin != g.Args.Name {
		g.Router.AddContactIfAbsent(private.Origin, sender)
	}

	// Check if the message is for me
	if g.Args.Name == private.Destination {
		fmt.Printf("%s\n", private.PrivateMessageToString())
		g.Router.AddContactIfAbsent(private.Origin, sender)
		g.NameIndex.AddPrivateMessage(private)
		return
	}

	// Decrement hop limit
	private.HopLimit--

	// Send/Relay private message if hop-limit not exhausted
	if private.HopLimit != 0 {

		// Pick the target (should exist) and send
		if target := g.Router.GetTarget(private.Destination); target != nil {
			OnSendPrivate(g, private, target)
		}
	}

}
