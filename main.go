package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func openUDPChannel(s string) (*net.UDPConn, error) {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return nil, &CustomError{"openUDPChannel", "cannot resolve UDP address"}
	}

	// Open an UDP connection
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, &CustomError{"openUDPChannel", "cannot listen on UDP channel"}
	}
	return udpConn, nil
}

func isPacketValid(pkt *GossipPacket, isSimpleMode bool, isClientSide bool) bool {

	// Exactly one of the field of the GossipPacket must be non-nil
	if (pkt.simpleMsg != nil && pkt.rumor != nil) || (pkt.simpleMsg != nil && pkt.status != nil) ||
		(pkt.rumor != nil && pkt.status != nil) || (pkt.simpleMsg == nil && pkt.rumor == nil && pkt.status == nil) {
		return false
	}

	// If we are in simple mode then it must be a SimpleMessage
	if isSimpleMode && pkt.simpleMsg == nil {
		return false
	}

	/* If we are not in simple mode then it must be either a RumorMessage
	or a StatusPacket */
	if !isSimpleMode && pkt.simpleMsg != nil {
		return false
	}

	/* If the packet comes from the client and we are not in simple mode
	then it must be a RumorMessage (the client can't send a StatusPacket)*/
	if isClientSide && !isSimpleMode && pkt.rumor == nil {
		return false
	}

	return true
}

// UDPDispatcher --
func (g *Gossiper) UDPDispatcher(addrPort string, OnReceiveBroadcast func(*Gossiper, *net.UDPConn, *SimpleMessage), isClientSide bool) {

	var err error

	// Open an UDP connection
	var channel *net.UDPConn
	if channel, err = openUDPChannel(addrPort); err != nil {
		return
	}

	// Program a call to close the channel when we are done
	defer channel.Close()

	// Create a buffer to store arriving data
	buf := make([]byte, BufSize)

	for {

		var sender *net.UDPAddr
		if _, sender, err = channel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt *GossipPacket
		if err := protobuf.Decode(buf, pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(pkt, g.simpleMode, isClientSide) {
			// Error: ignore the packet
			continue
		}

		if !isClientSide {
			// TODO: handle rumor-status-response
		}

		// Select the right callback
		switch {
		case pkt.simpleMsg != nil:
			OnReceiveBroadcast(g, channel, pkt.simpleMsg)
		case pkt.rumor != nil:
			go OnReceiveRumor(g, channel, pkt.rumor, sender, nil, isClientSide)
		case pkt.status != nil:
			go OnReceiveStatus(g, channel, pkt.status, sender)
		}

	}
}

func main() {

	// Argument parsing
	var gossiper Gossiper
	if err := gossiper.parseArgumentsGossiper(); err != nil {
		fmt.Println(err)
		return
	}

	// Add myself to the named peer list
	gossiper.network.AddNamedPeer(gossiper.name)

	// Launch 2 threads for client-side and network-side communication
	go gossiper.UDPDispatcher(gossiper.clientAddr, OnBroadcastClient, true)
	go gossiper.UDPDispatcher(gossiper.gossipAddr, OnBroadcastNetwork, false)

	/*
		Question: display known peers taking into account potential new one?
		Question: message arriving not in sequence -> save or discard ?
		Question: print Peers on invalid message ?
	*/

}
