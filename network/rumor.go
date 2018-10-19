package network

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/dedis/protobuf"
)

// OnSendRumor - Sends a rumor
func OnSendRumor(g *types.Gossiper, rumor *types.RumorMessage, target *net.UDPAddr, threadID uint32) error {

	// Create the packet
	pkt := types.GossipPacket{Rumor: rumor}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &fail.CustomError{Fun: "OnSendRumor", Desc: "failed to encode RumorMessage"}
	}

	// Send the packet
	fmt.Printf("MONGERING with %s\n", fmt.Sprintf("%s", target))
	if _, err = g.GossipChannel.WriteToUDP(buf, target); err != nil {
		g.Timeouts.DeleteTimeoutHandler(threadID)
		return &fail.CustomError{Fun: "OnSendRumor", Desc: "failed to send RumorMessage"}
	}

	// NOTE: chance to miss the packet here, although unlikely

	/* Allocate a TimeoutHandler object that the UDPDispatcher will use
	to forward us the StatusPacket response */
	g.Timeouts.AddTimeoutHandler(threadID, target)

	// Create a timeout timer
	timer := time.NewTicker(time.Second)

	// Wait for the timeout
	select {
	case <-timer.C: // Timeout expired
	}

	// Stop the timer
	timer.Stop()

	response := g.Timeouts.DeleteTimeoutHandler(threadID)
	if response == nil { // The response did not arrive on time

		if rand.Int()%2 == 0 { // Flip a coin
			return nil // Stop the thread
		}

		// Spread the rumor to someone else

		if newTarget := g.PeerIndex.GetRandomPeer(target); newTarget != nil {
			fmt.Printf("FLIPPED COIN sending rumor to %s\n", fmt.Sprintf("%s", newTarget))
			OnSendRumor(g, rumor, newTarget, threadID)
		}

	} else { // We received a status response
		OnReceiveStatus(g, response, target, threadID)
	}

	return nil

}

// OnReceiveRumor - Called when a rumor is received
func OnReceiveRumor(g *types.Gossiper, rumor *types.RumorMessage, sender *net.UDPAddr, isClientSide bool, threadID uint32) {

	// Is the message a RouteRumor ?
	isRouteRumor := (rumor.Text == "")

	if isClientSide {
		// Create the message name and ID
		rumor.Origin = g.Args.Name
		rumor.ID = g.NameIndex.GetLastMessageID(g.Args.Name)
	} else {
		// Attempt to add the sending peer to the list of neighbors
		g.PeerIndex.AddPeerIfAbsent(sender)
	}

	// Print to the console
	if isClientSide {
		fmt.Printf("CLIENT MESSAGE %s\n%s\n", rumor.Text, g.PeerIndex.PeersToString())
	} else {
		if !isRouteRumor {
			fmt.Printf("%s\n%s\n", rumor.RumorMessageToString(types.UDPAddressToString(sender)), g.PeerIndex.PeersToString())
		}
	}

	// Store the new message
	if g.NameIndex.AddMessageIfNext(rumor) { // TODO: unorder but higher ?
		// If the message was in sequence, update the routing table and print
		g.Router.UpdateTableAndPrint(rumor.Origin, sender)
	}

	// Reply with status message
	vectorClock := g.NameIndex.GetVectorClock()
	OnSendStatus(vectorClock, g.GossipChannel, sender)

	// Pick a random peer
	var target *net.UDPAddr
	if isClientSide {
		// There is no risk to propagate back to ourself
		target = g.PeerIndex.GetRandomPeer(nil)
	} else {
		// Prevent the sender from being selected
		target = g.PeerIndex.GetRandomPeer(sender)
	}

	if target == nil { // There is no one to propagate too
		return
	}

	// Propagate rumor
	OnSendRumor(g, rumor, target, threadID)
}
