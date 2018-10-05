package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"Peerster/utils"
	"github.com/dedis/protobuf"
)

func openUDPChannel(s string) (*net.UDPConn, error) {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return nil, &utils.CustomError{"openUDPChannel", "cannot resolve UDP address"}
	}

	// Open an UDP connection
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, &utils.CustomError{"openUDPChannel", "cannot listen on UDP channel"}
	}
	return udpConn, nil
}

func isPacketValid(pkt *utils.GossipPacket, isSimpleMode bool, isClientSide bool) bool {

	// Exactly one of the field of the GossipPacket must be non-nil
	if (pkt.SimpleMsg != nil && pkt.Rumor != nil) || (pkt.SimpleMsg != nil && pkt.Status != nil) ||
		(pkt.Rumor != nil && pkt.Status != nil) || (pkt.SimpleMsg == nil && pkt.Rumor == nil && pkt.Status == nil) {
		return false
	}

	// If we are in simple mode then it must be a SimpleMessage
	if isSimpleMode && pkt.SimpleMsg == nil {
		return false
	}

	/* If we are not in simple mode then it must be either a RumorMessage
	or a StatusPacket */
	if !isSimpleMode && pkt.SimpleMsg != nil {
		return false
	}

	/* If the packet comes from the client and we are not in simple mode
	then it must be a RumorMessage (the client can't send a StatusPacket)*/
	if isClientSide && !isSimpleMode && pkt.Rumor == nil {
		return false
	}

	return true
}

// UDPDispatcher --
func UDPDispatcher(g *utils.Gossiper, addrPort string, OnReceiveBroadcast func(*utils.Gossiper, *net.UDPConn, *utils.SimpleMessage), isClientSide bool) {

	var err error

	// Open an UDP connection
	var channel *net.UDPConn
	if channel, err = openUDPChannel(addrPort); err != nil {
		return
	}

	// Program a call to close the channel when we are done
	defer channel.Close()

	// Launch the anti-entropy thread
	go utils.AntiEntropy(g, channel)

	// Create a buffer to store arriving data
	buf := make([]byte, utils.BufSize)

	for {

		var sender *net.UDPAddr
		if _, sender, err = channel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt *utils.GossipPacket
		if err := protobuf.Decode(buf, pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(pkt, g.SimpleMode, isClientSide) {
			// Error: ignore the packet
			continue
		}

		// A client should never receive a StatusPacket
		if !isClientSide && pkt.Status != nil {
			isPacketHandled := false
			g.Timeouts.Mux.Lock()
			for _, t := range g.Timeouts.Responses { // Attempt to find a handler for this response
				if !t.Done && utils.CompareUDPAddress(t.Addr, sender) { // Match !
					t.Com <- *pkt.Status
					t.Done = true
					isPacketHandled = true
					break
				}
			}

			g.Timeouts.Mux.Unlock()
			if isPacketHandled {
				continue
			}
		}

		// Select the right callback
		switch {
		case pkt.SimpleMsg != nil:
			OnReceiveBroadcast(g, channel, pkt.SimpleMsg)
		case pkt.Rumor != nil:
			go utils.OnReceiveRumor(g, channel, pkt.Rumor, sender, isClientSide)
		case pkt.Status != nil:
			go utils.OnReceiveStatus(g, channel, pkt.Status, sender)
		}

	}
}

func main() {

	// Argument parsing
	var gossiper utils.Gossiper
	if err := gossiper.ParseArgumentsGossiper(); err != nil {
		fmt.Println(err)
		return
	}

	// Add myself to the named peer list
	gossiper.Network.AddNamedPeer(gossiper.Name)

	// Launch 2 threads for client-side and network-side communication
	go UDPDispatcher(&gossiper, gossiper.ClientAddr, utils.OnBroadcastClient, true)
	go UDPDispatcher(&gossiper, gossiper.GossipAddr, utils.OnBroadcastNetwork, false)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	/*
		Question: display known peers taking into account potential new one?
		Question: message arriving not in sequence -> save or discard ?
		Question: print Peers on invalid message ?
	*/

}
