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

func (g *Gossiper) listeningUDP(addrPort string, callbackBroadcast func(*Gossiper, *net.UDPConn, *GossipPacket)) {

	var err error

	// Open an UDP connection
	var udpChannel *net.UDPConn
	if udpChannel, err = openUDPChannel(addrPort); err != nil {
		return
	}

	// Program a call to close the channel when we are done
	defer udpChannel.Close()

	// Create a buffer to store arriving data
	buf := make([]byte, BufSize)

	for {

		var sender *net.UDPAddr
		if _, sender, err = udpChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt *GossipPacket
		if err := protobuf.Decode(buf, pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Exactly one of the field of the GossipPacket can be non-nil
		if (pkt.simpleMsg != nil && pkt.rumor != nil) ||
			(pkt.simpleMsg != nil && pkt.status != nil) ||
			(pkt.rumor != nil && pkt.status != nil) ||
			(pkt.simpleMsg == nil && pkt.rumor == nil && pkt.status == nil) {
			// Error: ignore the packet
			continue
		}

		// Create another thread to do the work and select the right callback
		switch {
		case pkt.simpleMsg != nil:
			go callbackBroadcast(g, udpChannel, pkt)
		case pkt.rumor != nil:
			go callbackRumor(g, udpChannel, pkt, sender)
		case pkt.status != nil:
			go callbackStatus(g, udpChannel, pkt, sender)
		}

	}
}

func callbackRumor(g *Gossiper, udpChannel *net.UDPConn, pkt *GossipPacket, sender *net.UDPAddr) {

}

func callbackStatus(g *Gossiper, udpChannel *net.UDPConn, pkt *GossipPacket, sender *net.UDPAddr) {

}

func main() {

	// Argument parsing
	var gossiper Gossiper
	if err := gossiper.parseArgumentsGossiper(); err != nil {
		fmt.Println(err)
		return
	}

	// Launch 2 threads for client and peer communication
	go gossiper.listeningUDP(gossiper.clientAddr, callbackClient)
	go gossiper.listeningUDP(gossiper.gossipAddr, callbackPeer)

}
