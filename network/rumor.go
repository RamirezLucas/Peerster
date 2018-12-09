package network

import (
	"Peerster/entities"
	"Peerster/fail"
	"Peerster/messages"
	"Peerster/peers"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/dedis/protobuf"
)

// OnSendRouteRumor - Sends a route rumor
func OnSendRouteRumor(g *entities.Gossiper, threadID uint32) {

	// Pick a random target and send a RouteRumor message
	target := g.PeerIndex.GetRandomPeer(nil)

	// Create a RouteRumor message
	routeRumor := &messages.RumorMessage{Text: ""}

	if target != nil {
		// Store the new message
		g.NameIndex.FillInRumorAndSave(routeRumor, g.Args.Name)
		// Send the route rumor
		OnSendRumor(g, routeRumor, target, threadID)
	}
}

// OnSendRumor - Sends a rumor
func OnSendRumor(g *entities.Gossiper, rumor *messages.RumorMessage, target *net.UDPAddr, threadID uint32) error {

	// Create the packet
	pkt := messages.GossipPacket{Rumor: rumor}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &fail.CustomError{Fun: "OnSendRumor", Desc: "failed to encode RumorMessage"}
	}

	// Send the packet
	fail.LeveledPrint(0, "", "MONGERING with %s", fmt.Sprintf("%s", target))
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
			fail.LeveledPrint(0, "", "FLIPPED COIN sending rumor to %s", peers.UDPAddressToString(newTarget))
			OnSendRumor(g, rumor, newTarget, threadID)
		}

	} else { // We received a status response
		OnReceiveStatus(g, response, target, threadID)
	}

	return nil
}

// OnReceiveClientRumor - Called when a rumor is received from the client
func OnReceiveClientRumor(g *entities.Gossiper, rumor *messages.RumorMessage, threadID uint32) {

	// Print to console
	fail.LeveledPrint(0, "", "CLIENT MESSAGE %s", rumor.Text)
	fail.LeveledPrint(0, "", g.PeerIndex.PeersToString())

	// Store the new message
	g.NameIndex.FillInRumorAndSave(rumor, g.Args.Name)

	// There is no risk to propagate back to ourself
	target := g.PeerIndex.GetRandomPeer(nil)

	if target == nil { // There is no one to propagate too
		return
	}

	// Propagate rumor
	OnSendRumor(g, rumor, target, threadID)
}

// OnReceiveRumor - Called when a rumor is received
func OnReceiveRumor(g *entities.Gossiper, rumor *messages.RumorMessage, sender *net.UDPAddr, threadID uint32) {

	// Is the message a RouteRumor ?
	isRouteRumor := (rumor.Text == "")

	if !isRouteRumor {
		fail.LeveledPrint(0, "", rumor.RumorMessageToString(peers.UDPAddressToString(sender)))
		fail.LeveledPrint(0, "", g.PeerIndex.PeersToString())
	}

	// Update the routing table for private messages
	if rumor.Origin != g.Args.Name {
		g.Router.UpdateTableAndPrint(rumor.Origin, sender, rumor.ID)
	}

	// Store the new message
	g.NameIndex.AddMessageIfNext(rumor)

	// Reply with status message
	vectorClock := g.NameIndex.GetVectorClock()
	OnSendStatus(vectorClock, g.GossipChannel, sender)

	// Prevent the sender from being selected
	target := g.PeerIndex.GetRandomPeer(sender)

	if target == nil { // There is no one to propagate too
		return
	}

	// Propagate rumor
	OnSendRumor(g, rumor, target, threadID)
}
