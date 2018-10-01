package main

import (
	"fmt"
	"net"

	"github.com/dedis/protobuf"
)

const bufSize = 1024

func openUDPChannel(s string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return nil, &CustomError{"openUDPChannel", "cannot resolve UDP address"}
	}
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, &CustomError{"openUDPChannel", "cannot listen on UDP channel"}
	}
	return udpConn, nil
}

func (g *Gossiper) listenUDPChannel(addr string, callback func(*net.UDPConn, *Gossiper, *GossipPacket) error) {

	udpChannel, err := openUDPChannel(addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Program a call to close the channel when we are done
	defer udpChannel.Close()

	buf := make([]byte, bufSize)

	for {
		if _, _, err := udpChannel.ReadFromUDP(buf); err != nil {
			fmt.Println("Error: ", err)
		}
		// TODO: Check sender address ?

		var pkt *GossipPacket
		if err := protobuf.Decode(buf, pkt); err != nil {
			// Error: ignore the packet
		}

		if err := callback(udpChannel, g, pkt); err != nil {
			// Error: do something
		}

	}

}

func callbackClient(udpChannel *net.UDPConn, g *Gossiper, pkt *GossipPacket) error {

	// Print the message on standard output
	fmt.Println("CLIENT MESSAGE ", pkt.msg.contents)

	// Modify the packet
	pkt.msg.originalName = g.name
	pkt.msg.relayPeerAddr = g.gossipAddr

	// Create the packet
	buf, err := protobuf.Encode(*pkt)
	if err != nil {
		return &CustomError{"callbackClient", "failed to encode packet"}
	}

	// Send to everyone
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	for _, peer := range g.peers {
		// TODO: remove code copy
		udpAddr, err := net.ResolveUDPAddr("udp4", peer)
		if err != nil {
			return &CustomError{"callbackClient", "unable to resolve UDP address"}
		}
		if _, err = udpChannel.WriteToUDP(buf, udpAddr); err != nil {
			return &CustomError{"callbackClient", "unable to write on UDP channel"}
		}
	}

	return nil
}

func callbackPeer(udpChannel *net.UDPConn, g *Gossiper, pkt *GossipPacket) error {

	// Print the message on standard output
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s",
		pkt.msg.originalName, pkt.msg.relayPeerAddr, pkt.msg.contents)

	// Modify the packet
	sender := pkt.msg.relayPeerAddr
	pkt.msg.relayPeerAddr = g.gossipAddr

	// Create the packet
	buf, err := protobuf.Encode(*pkt)
	if err != nil {
		return &CustomError{"callbackPeer", "failed to encode packet"}
	}

	// Send to everyone (except the sender)
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	isPeerKnown := false
	for _, peer := range g.peers {
		if sender == peer {
			isPeerKnown = true
		} else {
			// TODO: remove code copy
			udpAddr, err := net.ResolveUDPAddr("udp4", peer)
			if err != nil {
				return &CustomError{"callbackPeer", "unable to resolve UDP address"}
			}
			if _, err = udpChannel.WriteToUDP(buf, udpAddr); err != nil {
				return &CustomError{"callbackPeer", "unable to write on UDP channel"}
			}
		}
	}
	if !isPeerKnown { // We need to add the sender to the peers list
		g.peers = append(g.peers, sender)
	}

	return nil
}

func main() {

	var gossiper Gossiper

	if err := gossiper.parseArgumentsGossiper(); err != nil {
		fmt.Println(err)
		return
	}

	go gossiper.listenUDPChannel(gossiper.clientAddr, callbackClient)
	go gossiper.listenUDPChannel(gossiper.gossipAddr, callbackClient)

}
