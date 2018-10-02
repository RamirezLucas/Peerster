package main

import (
	"fmt"
	"math/rand"
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

func (g *Gossiper) listeningUDP(addrPort string, self bool, callbackBroadcast func(*Gossiper, *net.UDPConn, *SimpleMessage)) {

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

		/* If the gossiper is operating on simple broadcast reject rumors and
		status requests */
		if g.simpleMode && pkt.simpleMsg == nil {
			// Error: ignore the packet
			continue
		}

		// Select the right callback
		switch {
		case pkt.simpleMsg != nil:
			callbackBroadcast(g, udpChannel, pkt.simpleMsg)
		case pkt.rumor != nil:
			callbackRumor(g, udpChannel, pkt.rumor, sender, self)
		case pkt.status != nil:
			callbackStatus(g, udpChannel, pkt.status, sender)
		}

	}
}

func callbackRumor(g *Gossiper, udpChannel *net.UDPConn, rumor *RumorMessage, sender *net.UDPAddr, self bool) {

	// Print to the console
	g.mux.Lock()
	fmt.Printf("%s%v", rumor.ToString(fmt.Sprintf("%v", sender)), g.peers)

	// TODO: Check if message is sequential or from self

	// Pick random peer
	n := rand.Intn(len(g.peers.list))
	target := g.peers.list[n]
	g.mux.Unlock()

	// TODO: Save the message

	// Create the packet
	pkt := GossipPacket{rumor: rumor}
	buf, err := protobuf.Encode(pkt)
	if err != nil {
		return
	}

	// Relay to selected target
	if _, err = udpChannel.WriteToUDP(buf, target.udpAddr); err != nil {
		return
	}

	// Wait for timeout

}

func callbackStatus(g *Gossiper, udpChannel *net.UDPConn, status *StatusPacket, sender *net.UDPAddr) {

}

func main() {

	// Argument parsing
	var gossiper Gossiper
	if err := gossiper.parseArgumentsGossiper(); err != nil {
		fmt.Println(err)
		return
	}

	// Launch 2 threads for client and peer communication
	go gossiper.listeningUDP(gossiper.clientAddr, true, callbackClient)
	go gossiper.listeningUDP(gossiper.gossipAddr, false, callbackPeer)

	/*
		Question: display known peers taking into account potential new one?
		Question: message arriving not in sequence -> save or discard ?
	*/

}
