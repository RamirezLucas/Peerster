package main

import (
	"Peerster/utils"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func main() {

	// Initialize the client
	var client utils.Client
	if err := client.ParseArgumentsClient(); err != nil {
		fmt.Println(err)
		return
	}

	// Create the packet
	simpleMsg := utils.SimpleMessage{Contents: client.Msg}
	pkt := utils.GossipPacket{SimpleMsg: &simpleMsg}

	// Encode the packet
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Establish a UDP connection
	udpAddr, err := net.ResolveUDPAddr("udp4", "localhost:0")
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
	if _, err = udpConn.WriteToUDP(buf, client.Addr); err != nil {
		fmt.Println(err)
		return
	}

}
