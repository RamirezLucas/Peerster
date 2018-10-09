package main

import (
	"Peerster/utils"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/dedis/protobuf"
)

func openUDPChannel(s string) (*net.UDPConn, error) {

	// Resolve the address
	udpAddr, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return nil, &utils.CustomError{Fun: "openUDPChannel", Desc: "cannot resolve UDP address"}
	}

	// Open an UDP connection
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, &utils.CustomError{Fun: "openUDPChannel", Desc: "cannot listen on UDP channel"}
	}
	return udpConn, nil
}

func isPacketValid(pkt *utils.GossipPacket, isClientSide bool, isSimpleMode bool) bool {

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

func antiEntropy(g *utils.Gossiper) {

	// Create a timeout timer
	timer := time.NewTicker(4 * time.Second)
	for {
		select {
		case <-timer.C:
			// Pick a random target and send a StatusPacket
			g.Network.Mux.Lock()
			target := g.Network.GetRandomPeer(nil)
			vectorClock := g.Network.VectorClock
			g.Network.Mux.Unlock()
			if target != nil {
				utils.OnSendStatus(&vectorClock, g.GossipChannel, target)
			}
		}
	}
}

func udpDispatcherGossip(g *utils.Gossiper) {

	// Create a buffer to store arriving data
	buf := make([]byte, utils.BufSize)

	for {

		var sender *net.UDPAddr
		var n int
		var err error

		if n, sender, err = g.GossipChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt utils.GossipPacket
		if err := protobuf.Decode(buf[:n], &pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(&pkt, false, g.SimpleMode) {
			// Error: ignore the packet
			continue
		}

		isPacketHandled := false

		// Timeout management
		if pkt.Status != nil {

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
		}

		if !isPacketHandled {
			// Select the right callback
			switch {
			case pkt.SimpleMsg != nil:
				utils.OnBroadcastNetwork(g, pkt.SimpleMsg)
			case pkt.Rumor != nil:
				go utils.OnReceiveRumor(g, pkt.Rumor, sender, false)
			case pkt.Status != nil:
				go utils.OnReceiveStatus(g, pkt.Status, sender)
			default:
				// Should never happen
			}
		}

	}
}

func udpDispatcherClient(g *utils.Gossiper) {

	// Create a buffer to store arriving data
	buf := make([]byte, utils.BufSize)

	for {

		var sender *net.UDPAddr
		var n int
		var err error
		if n, sender, err = g.ClientChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt utils.GossipPacket
		if err := protobuf.Decode(buf[:n], &pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(&pkt, true, g.SimpleMode) {
			// Error: ignore the packet
			continue
		}

		if g.SimpleMode { // Simple mode
			utils.OnBroadcastClient(g, pkt.SimpleMsg)
		} else {
			// Convert the message to a RumorMessage
			rumor := utils.RumorMessage{Text: pkt.SimpleMsg.Contents}
			go utils.OnReceiveRumor(g, &rumor, sender, true)
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

	// Create 2 communication channels
	var err error
	if gossiper.ClientChannel, err = openUDPChannel(gossiper.ClientAddr); err != nil {
		return
	}
	if gossiper.GossipChannel, err = openUDPChannel(gossiper.GossipAddr); err != nil {
		return
	}

	// Program a call to close the channels
	defer gossiper.ClientChannel.Close()
	defer gossiper.GossipChannel.Close()

	/* Launch 2 threads for client-side and network-side communication and one thread
	for the Anti-Entropy protocol */
	go udpDispatcherClient(&gossiper)
	go udpDispatcherGossip(&gossiper)
	go antiEntropy(&gossiper)

	// Kill all goroutines before exiting
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

}
