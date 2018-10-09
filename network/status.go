package network

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

// OnSendStatus -
func OnSendStatus(vectorClock *types.StatusPacket, channel *net.UDPConn, target *net.UDPAddr) error {

	// Create the packet
	pkt := types.GossipPacket{Status: vectorClock}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &fail.CustomError{Fun: "OnSendStatus", Desc: "failed to encode GossipPacket"}
	}

	// Send the packet
	if _, err = channel.WriteToUDP(buf, target); err != nil {
		return &fail.CustomError{Fun: "OnSendStatus", Desc: "failed to send StatusPacket"}
	}
	return nil
}

// OnReceiveStatus -
func OnReceiveStatus(g *types.Gossiper, status *types.StatusPacket, sender *net.UDPAddr) {

	// Attempt to add the sending peer to the list of neighbors
	g.PeerIndex.AddPeerIfAbsent(sender)

	// Print to the console
	fmt.Printf("%s\n%s\n", status.StatusPacketToString(types.UDPAddressToString(sender)), g.PeerIndex.PeersToString())

	// See if we must propagate a rumor
	rumorToPropagate := g.NameIndex.GetUnknownMessageTarget(status)

	if rumorToPropagate == nil { // We don't have anything to propagate
		if g.NameIndex.IsLocalStatusComplete(status) { // We are in sync with the other
			fmt.Printf("IN SYNC WITH %s\n", fmt.Sprintf("%s", sender))
		} else { // We must send back our own Status
			vectorClock := g.NameIndex.GetVectorClock()
			OnSendStatus(vectorClock, g.GossipChannel, sender)
		}
	} else { // We must propagate a rumor to the sender
		OnSendRumor(g, rumorToPropagate, sender)
	}

}
