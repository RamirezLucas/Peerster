package main

import "fmt"

// PeersToString - Returns a textual representation of a []Peer
func PeersToString(peers []Peer) string {
	s := "PEERS"
	for _, peer := range peers {
		s = s + fmt.Sprintf(" %s", peer.rawAddr)
	}
	return s
}

// SimpleMessageToString - Returns a textual representation of a SimpleMessage
func SimpleMessageToString(msg *SimpleMessage) string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s",
		msg.originalName, msg.relayPeerAddr, msg.contents)
}

// RumorMessageToString - Returns a textual representation of a RumorMessage
func RumorMessageToString(rumor *RumorMessage, relayAddr string) string {
	return fmt.Sprintf("RUMOR origin %s from %s ID %d contents %s",
		rumor.origin, relayAddr, rumor.id, rumor.text)
}

// StatusPacketToString - Returns a textual representation of a StatusPacket
func StatusPacketToString(status *StatusPacket, relayAddr string) string {
	s := fmt.Sprintf("STATUS from %s", relayAddr)
	for _, peer := range status.want {
		s = s + fmt.Sprintf(" peer %s nextID %d", peer.identifier, peer.nextID)
	}
	return s
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.fun, e.desc)
}
