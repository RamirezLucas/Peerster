package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func main() {

	// Initialize the client
	var client Client
	if err := client.parseArgumentsClient(); err != nil {
		fmt.Println(err)
		return
	}

	// Create the packet
	simpleMsg := SimpleMessage{originalName: "",
		relayPeerAddr: "",
		contents:      client.msg}
	pkt := GossipPacket{&simpleMsg}

	// Encode the packet
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Establish a UDP connection
	udpAddr, err := net.ResolveUDPAddr("udp4", client.addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Program a call to close the channel when we are done
	defer udpConn.Close()

	// Send to local gossiper
	if _, err = udpConn.WriteToUDP(buf, udpAddr); err != nil {
		fmt.Println(err)
		return
	}

}
