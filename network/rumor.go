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

// OnSendRumor -
func OnSendRumor(g *types.Gossiper, rumor *types.RumorMessage, target *net.UDPAddr) error {

	// Create the packet
	pkt := types.GossipPacket{Rumor: rumor}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &fail.CustomError{Fun: "OnSendRumor", Desc: "failed to encode RumorMessage"}
	}

	// Send the packet
	fmt.Printf("MONGERING with %s\n", fmt.Sprintf("%s", target))
	if _, err = g.GossipChannel.WriteToUDP(buf, target); err != nil {
		return &fail.CustomError{Fun: "OnSendRumor", Desc: "failed to send RumorMessage"}
	}

	/* Allocate a TimeoutHandler object that the UDPDispatcher will use
	to forward us the StatusPacket response */
	responseChan := make(chan types.StatusPacket)
	hash := rand.Int()
	g.Timeouts.Mux.Lock()
	g.Timeouts.Responses = append(g.Timeouts.Responses, types.TimeoutHandler{Addr: target, Com: responseChan, Hash: hash, Done: false})
	g.Timeouts.Mux.Unlock()

	// Create a timeout timer
	timer := time.NewTicker(time.Second)
	var response *types.StatusPacket
	stop := false

	// Wait for an answer or a timeout, whichever is first
	for !stop {
		select {
		case <-timer.C: // Timeout expired
			stop = true
		case r := <-responseChan:
			response = &r
			stop = true
		}
	}

	// Stop the timer
	timer.Stop()
	g.Timeouts.Mux.Lock()

	// Last chance to get the response status
	select {
	case r := <-responseChan:
		response = &r
	default:
		// Do nothing
	}

	for i, t := range g.Timeouts.Responses {
		if hash == t.Hash { // Found our timeout
			// Delete our handler
			len := len(g.Timeouts.Responses)
			g.Timeouts.Responses[i] = g.Timeouts.Responses[len-1]
			g.Timeouts.Responses = g.Timeouts.Responses[:len-1]
		}
	}
	g.Timeouts.Mux.Unlock()

	if response == nil { // The response did not arrive on time
		if rand.Int()%2 == 0 { // Flip a coin
			return nil // Stop the thread
		}
		// Spread the rumor to someone else
		newTarget := g.PeerIndex.GetRandomPeer(target)
		fmt.Printf("FLIPPED COIN sending rumor to %s\n", fmt.Sprintf("%s", newTarget))
		OnSendRumor(g, rumor, newTarget)
	} else { // We received a status response
		OnReceiveStatus(g, response, target)
	}

	return nil

}

// OnReceiveRumor -
func OnReceiveRumor(g *types.Gossiper, rumor *types.RumorMessage, sender *net.UDPAddr, isClientSide bool) {

	if isClientSide {
		// Create the message name and ID
		rumor.Origin = g.Name
		rumor.ID = g.NameIndex.GetLastMessageID(g.Name)
	} else {
		// Attempt to add the sending peer to the list of neighbors
		g.PeerIndex.AddPeerIfAbsent(sender)
	}

	// Print to the console
	if isClientSide {
		fmt.Printf("CLIENT MESSAGE %s\n%s\n", rumor.Text, g.PeerIndex.PeersToString())
	} else {
		fmt.Printf("%s\n%s\n", rumor.RumorMessageToString(types.UDPAddressToString(sender)), g.PeerIndex.PeersToString())
	}

	// Store the new message
	g.NameIndex.AddMessageIfNext(rumor)

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
	OnSendRumor(g, rumor, target)
}
