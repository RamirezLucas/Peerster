package network

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnSendPrivate - Sends a private message
func OnSendPrivate(g *types.Gossiper, private *types.PrivateMessage, target *net.UDPAddr) error {

	// Create the packet
	pkt := types.GossipPacket{Private: private}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &fail.CustomError{Fun: "OnSendPrivate", Desc: "failed to encode GossipPacket"}
	}

	// Send the packet
	if _, err = g.GossipChannel.WriteToUDP(buf, target); err != nil {
		return &fail.CustomError{Fun: "OnSendPrivate", Desc: "failed to send PrivateMessage"}
	}
	return nil

}

// OnReceiveClientPrivate - Called when a private message is received from the client
func OnReceiveClientPrivate(g *types.Gossiper, private *types.PrivateMessage) {

	// Fill in remaining fields
	private.Origin = g.Args.Name
	private.ID = 0
	private.HopLimit = 16

	// Add the message
	g.NameIndex.AddPrivateMessage(private)

	// Pick the target (should exist) and send
	target := g.Router.GetTarget(private.Destination)
	if target != nil {
		OnSendPrivate(g, private, target)
	}

}

// OnReceivePrivate - Called when a private message is received
func OnReceivePrivate(g *types.Gossiper, private *types.PrivateMessage, sender *net.UDPAddr) {

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
		target := g.Router.GetTarget(private.Destination)
		if target != nil {
			OnSendPrivate(g, private, target)
		}
	}

}
