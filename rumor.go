package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/dedis/protobuf"
)

// OnReceiveRumor -
func OnReceiveRumor(g *Gossiper, channel *net.UDPConn, rumor *RumorMessage, sender *net.UDPAddr,
	target *net.UDPAddr, timeouts *StatusResponseForwarder, isClientSide bool) {

	g.mux.Lock()
	/* ==== THREAD SAFE - BEGIN ==== */

	// Print to the console
	if isClientSide {
		fmt.Printf("CLIENT MESSAGE %s\n%s\n", rumor.text, PeersToString(g.network.peers))
	} else {
		fmt.Printf("%s\n%s\n", RumorMessageToString(rumor, fmt.Sprintf("%s", sender)), PeersToString(g.network.peers))
	}

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

	/* ==== THREAD SAFE - END ==== */
	g.mux.Unlock()

	// Create the packet
	pkt := GossipPacket{rumor: rumor}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send the packet
	fmt.Printf("MONGERING with %s\n", fmt.Sprintf("%s", target))
	if _, err = channel.WriteToUDP(buf, target); err != nil {
		return
	}

	/* Allocate a TimeoutHandler object that the UDPDispatcher will use
	to forward us the StatusPacket response */
	responseChan := make(chan StatusPacket)
	timeouts.mux.Lock()
	timeouts.responses = append(timeouts.responses, TimeoutHandler{target, responseChan, false})
	timeouts.mux.Unlock()

	// Create a timeout timer
	timer := time.NewTicker(time.Second)
	var response *StatusPacket
	stopWaiting := false

	for !stopWaiting {
		select {
		case <-timer.C: // Timeout expired
			stopWaiting = true
		case r := <-responseChan:
			response = &r
			stopWaiting = true
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Stop the timer
	timer.Stop()
	// TODO: remove the timeout handler

	if response == nil { // The response did not arrive on time
		if rand.Int()%2 == 0 { // Flip a coin
			return // Stop the thread
		}
		// Spread the rumor to someone else
		// TODO: prevent resending to the same peer
		// TODO: print flip coin
		OnReceiveRumor(g, channel, rumor, sender, nil, timeouts, true)

	} else { // We received a status response
		OnReceiveStatus(g, channel, response, target, target)
	}

}
