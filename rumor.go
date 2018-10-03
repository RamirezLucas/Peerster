package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnReceiveRumor -
func OnReceiveRumor(g *Gossiper, channel *net.UDPConn, rumor *RumorMessage, sender *net.UDPAddr, target *net.UDPAddr, isClientSide bool) {

	// Print to the console
	g.mux.Lock()

	/* ==== THREAD SAFE - BEGIN ==== */

	fmt.Printf("%s\n%s\n", RumorMessageToString(rumor, fmt.Sprintf("%s", sender)), PeersToString(g.network.peers))

	if isClientSide {
		// Create the message name and ID
		rumor.origin = g.name
		rumor.id = g.network.GetLastMessageID(g.name)
	} else {
		// Attempt to add the sending peer to the list of neighbors
		g.network.AddPeerIfAbsent(sender)
	}

	// Store the new message
	// TODO: the message might not have been added (what to do ?)
	g.network.AddMessageIfNext(rumor)

	// Pick a random peer if target == nil
	if target == nil {
		if isClientSide {
			// There is no risk to propagate back to ourself
			target = g.network.GetRandomPeer(nil)
		} else {
			// Prevent the sender from being selected
			target = g.network.GetRandomPeer(sender)
		}
		if target == nil { // There is no one to propagate too
			return
		}
	}
	g.mux.Unlock()

	/* ==== THREAD SAFE - END ==== */

	// Create the packet
	pkt := GossipPacket{rumor: rumor}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send the packet
	if _, err = channel.WriteToUDP(buf, target); err != nil {
		return
	}

	// Create a timeout timer

}
