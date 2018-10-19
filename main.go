package main

import (
	"Peerster/backend"
	"Peerster/fail"
	"Peerster/network"
	"Peerster/parsing"
	"Peerster/types"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dedis/protobuf"
)

func threadIDGenerator(chanID chan uint32) {
	threadID := uint32(1)
	for {
		chanID <- threadID
		threadID++
	}
}

func openUDPChannel(s string) (*net.UDPConn, error) {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return nil, &fail.CustomError{Fun: "openUDPChannel", Desc: "cannot resolve UDP address"}
	}

	// Open an UDP connection
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, &fail.CustomError{Fun: "openUDPChannel", Desc: "cannot listen on UDP channel"}
	}
	return udpConn, nil
}

func isPacketValid(pkt *types.GossipPacket, isClientSide bool, isSimpleMode bool) bool {

	// Exactly one of the field of the GossipPacket must be non-nil
	if (pkt.SimpleMsg != nil && pkt.Rumor != nil) || (pkt.SimpleMsg != nil && pkt.Status != nil) ||
		(pkt.Rumor != nil && pkt.Status != nil) || (pkt.SimpleMsg == nil && pkt.Rumor == nil && pkt.Status == nil) {
		return false
	}

	// The client only sends SimpleMsg
	if isClientSide && pkt.SimpleMsg == nil {
		return false
	}

	// In simple mode only accept simple messages
	if isSimpleMode && pkt.SimpleMsg == nil {
		return false
	}

	return true
}

func antiEntropy(g *types.Gossiper) {

	// Create a timeout timer
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			// Pick a random target and send a StatusPacket
			target := g.PeerIndex.GetRandomPeer(nil)
			vectorClock := g.NameIndex.GetVectorClock()
			if target != nil {
				network.OnSendStatus(vectorClock, g.GossipChannel, target)
			}
		}
	}
}

func routeRumor(g *types.Gossiper, chanID chan uint32) {

	// Send a rumor on startup
	sendRouteRumor(g, chanID)

	// Create a timeout timer
	timer := time.NewTicker(time.Duration(g.Args.RTimer) * time.Second)
	for {
		select {
		case <-timer.C:
			sendRouteRumor(g, chanID)
		}
	}

}

func sendRouteRumor(g *types.Gossiper, chanID chan uint32) {
	// Pick a random target and send a RouteRumor message
	target := g.PeerIndex.GetRandomPeer(nil)
	nextID := g.NameIndex.GetLastMessageID(g.Args.Name)

	// Create a RouteRumor message
	routeRumor := &types.RumorMessage{Origin: g.Args.Name, ID: nextID, Text: ""}

	if target != nil {
		network.OnSendRumor(g, routeRumor, target, <-chanID)
	}
}

func udpDispatcherGossip(g *types.Gossiper, chanID chan uint32) {

	// Create a buffer to store arriving data
	buf := make([]byte, types.BufSize)

	for {

		var sender *net.UDPAddr
		var n int
		var err error

		if n, sender, err = g.GossipChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt types.GossipPacket
		if err := protobuf.Decode(buf[:n], &pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(&pkt, false, g.Args.SimpleMode) {
			// Error: ignore the packet
			continue
		}

		// Select the right callback
		switch {
		case pkt.SimpleMsg != nil:

			// Make sure the client isn't talking on the network port
			if pkt.SimpleMsg.RelayPeerAddr == "" {
				// Error: ignore the packet
				continue
			}

			network.OnBroadcastNetwork(g, pkt.SimpleMsg)
		case pkt.Rumor != nil:
			go network.OnReceiveRumor(g, pkt.Rumor, sender, false, <-chanID)
		case pkt.Status != nil:

			isPacketHandled := g.Timeouts.SearchAndForward(sender, pkt.Status)
			if !isPacketHandled {
				go network.OnReceiveStatus(g, pkt.Status, sender, <-chanID)
			}
		default:
			// Should never happen
		}

	}
}

func udpDispatcherClient(g *types.Gossiper, chanID chan uint32) {

	// Create a buffer to store arriving data
	buf := make([]byte, types.BufSize)

	for {

		var sender *net.UDPAddr
		var n int
		var err error
		if n, sender, err = g.ClientChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt types.GossipPacket
		if err := protobuf.Decode(buf[:n], &pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(&pkt, true, g.Args.SimpleMode) {
			// Error: ignore the packet
			continue
		}

		if g.Args.SimpleMode { // Simple mode
			network.OnBroadcastClient(g, pkt.SimpleMsg)
		} else {
			// Convert the message to a RumorMessage
			rumor := types.RumorMessage{Text: pkt.SimpleMsg.Contents}
			go network.OnReceiveRumor(g, &rumor, sender, true, <-chanID)
		}

	}

}

func main() {

	// Argument parsing
	args, err := parsing.ParseArgumentsGossiper()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create the gossiper
	gossiper := types.NewGossiper(args)

	// Add myself to the named peer list
	gossiper.NameIndex.AddName(gossiper.Args.Name)

	// Create 2 communication channels
	if gossiper.ClientChannel, err = openUDPChannel(gossiper.Args.ClientAddr); err != nil {
		return
	}
	if gossiper.GossipChannel, err = openUDPChannel(gossiper.Args.GossipAddr); err != nil {
		return
	}

	// Program a call to close the channels
	defer gossiper.ClientChannel.Close()
	defer gossiper.GossipChannel.Close()

	// Launch a thread giving thread IDs
	chanID := make(chan uint32)
	go threadIDGenerator(chanID)

	// Launch a thread for the gossiper dispatcher
	go udpDispatcherGossip(gossiper, chanID)

	// Launch a thread for the client dispatcher
	go udpDispatcherClient(gossiper, chanID)

	// Launch the webserver if the ports do not clash
	if strings.Split(gossiper.Args.ClientAddr, ":")[1] != gossiper.Args.ServerPort {
		go backend.Webserver(gossiper, chanID)
	}

	// Anti Entropy
	if !gossiper.Args.SimpleMode {
		go antiEntropy(gossiper)
	}

	// RouteRumor
	if gossiper.Args.RTimer != 0 {
		go routeRumor(gossiper, chanID)
	}

	// Kill all goroutines before exiting
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

}
