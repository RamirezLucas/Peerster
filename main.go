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

func (g *Gossiper) listeningUDP(addrPort string, callback func(*net.UDPConn, *Gossiper, *GossipPacket) error) {

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
		if _, _, err := udpChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		var pkt *GossipPacket
		if err := protobuf.Decode(buf, pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		if err := callback(udpChannel, g, pkt); err != nil {
			// Error: ignore the packet
			continue
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

	// Launch 2 threads for client and peer communication
	go gossiper.listeningUDP(gossiper.clientAddr, callbackClient)
	go gossiper.listeningUDP(gossiper.gossipAddr, callbackPeer)

}
