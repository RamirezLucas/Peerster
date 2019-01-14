package main

import (
	"Peerster/messages"
	"Peerster/parsing"
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

func main() {

	// Initialize the client
	client, err := parsing.ParseArgumentsClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create the packet
	var pkt messages.GossipPacket

	switch {
	// File search
	case client.Keywords != nil:
		search := messages.SearchRequest{
			Origin:   "",
			Budget:   client.Budget,
			Keywords: client.Keywords,
		}
		pkt = messages.GossipPacket{SearchRequest: &search}

	// Art publish
	case client.Filename != "" && client.ArtName != "" && client.ArtDesc != "":
		artTx := messages.ArtTx{
			HopLimit: 0,
			Artist:   &messages.ArtistInfo{},
			Artwork: &messages.ArtworkInfo{
				Filename:    client.Filename,
				Name:        client.ArtName,
				Description: client.ArtDesc,
			},
		}
		pkt = messages.GossipPacket{ArtTx: &artTx}
	// File request for someone else
	case client.Filename != "" && client.Request != nil:

		fileRequest := messages.DataRequest{
			HopLimit:    1,
			Destination: client.Dst,
			HashValue:   client.Request,
			Origin:      client.Filename,
		}
		pkt = messages.GossipPacket{DataRequest: &fileRequest}
	// File index
	case client.Filename != "":
		fileRequest := messages.DataRequest{
			HopLimit: 0,
			Origin:   client.Filename,
		}
		pkt = messages.GossipPacket{DataRequest: &fileRequest}
	// Private message
	case client.Dst != "" && client.Msg != "":
		privateMsg := messages.PrivateMessage{
			Text:        client.Msg,
			Destination: client.Dst,
		}
		pkt = messages.GossipPacket{Private: &privateMsg}
	// Simple rumor
	case client.Msg != "":
		simpleMsg := messages.SimpleMessage{Contents: client.Msg}
		pkt = messages.GossipPacket{SimpleMsg: &simpleMsg}
	default:
		fmt.Println("main(): Invalid arguments to main")
		return
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
