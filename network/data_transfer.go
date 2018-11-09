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

// OnSendDataRequest - Sends a data request
func OnSendDataRequest(g *types.Gossiper, request *types.DataRequest, target *net.UDPAddr, threadID uint32) error {

	// Create the packet
	pkt := types.GossipPacket{DataRequest: request}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &fail.CustomError{Fun: "OnSendDataRequest", Desc: "failed to encode GossipPacket"}
	}

	// Send the packet
	if _, err = g.GossipChannel.WriteToUDP(buf, target); err != nil {
		return &fail.CustomError{Fun: "OnSendDataRequest", Desc: "failed to send PrivateMessage"}
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

// OnRemoteFileRequest - Handles a remote file request
func OnRemoteFileRequest(g *types.Gossiper, metahash []byte, localFilename, remotePeer string, threadID uint32) {

	// Check that the remote peer exists
	target := g.Router.GetTarget(remotePeer)
	if target == nil {
		return
	}

	// Create metafile request and send it
	request := types.DataRequest{Origin: g.Args.Name,
		Destination: remotePeer,
		HopLimit:    16,
		HashValue:   metahash}

	OnSendDataRequest(g, &request, target, threadID)

	// Request each chunk one after the other

}
