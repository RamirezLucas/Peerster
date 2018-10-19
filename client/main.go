package main

import (
	"Peerster/parsing"
	"Peerster/types"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func main() {

	// Initialize the client
	client, err := parsing.ParseArgumentsClient(&client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create the packet
	var pkt types.GossipPacket
	if client.Dst == "" { // Normal message
		simpleMsg := types.SimpleMessage{Contents: client.Msg}
		pkt = types.GossipPacket{SimpleMsg: &simpleMsg}
	} else { // Private message
		privateMsg := types.PrivateMessage{Text: client.Msg,
			Destination: client.Dst}
		pkt = types.GossipPacket{Private: &privateMsg}
	}

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
