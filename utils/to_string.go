package utils

import (
	"Peerster/types"
	"fmt"
)

// PeersToString - Returns a textual representation of a []Peer
func PeersToString(peers []types.Peer) string {
	s := "PEERS "
	for _, peer := range peers {
		s = s + fmt.Sprintf("%s,", peer.RawAddr)
	}
	return s[:len(s)-1]
}

// GossiperToString -
func GossiperToString(g *types.Gossiper) string {
	return fmt.Sprintf("ClienAddr: %s\nGossipAddr: %s\nName: %s\nSimpleMode: %v\n",
		g.ClientAddr, g.GossipAddr, g.Name, g.SimpleMode)
}

// SimpleMessageToString - Returns a textual representation of a SimpleMessage
func SimpleMessageToString(msg *types.SimpleMessage) string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s",
		msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
}

// RumorMessageToString - Returns a textual representation of a RumorMessage
func RumorMessageToString(rumor *types.RumorMessage, relayAddr string) string {
	return fmt.Sprintf("RUMOR origin %s from %s ID %d contents %s",
		rumor.Origin, relayAddr, rumor.ID, rumor.Text)
}

// StatusPacketToString - Returns a textual representation of a StatusPacket
func StatusPacketToString(status *types.StatusPacket, relayAddr string) string {
	s := fmt.Sprintf("STATUS from %s", relayAddr)
	for _, peer := range status.Want {
		s = s + fmt.Sprintf(" peer %s nextID %d", peer.Identifier, peer.NextID)
	}
	return s
}
