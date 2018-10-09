package main

import (
	"Peerster/utils"
	"fmt"
	"net"
	"os"
	"os/signal"

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
	if !isClientSide && !g.SimpleMode {
		go utils.AntiEntropy(g, channel)
	}

	// Create a buffer to store arriving data
	buf := make([]byte, utils.BufSize)

	for {

		var sender *net.UDPAddr
		var n int
		if n, sender, err = channel.ReadFromUDP(buf); err != nil {
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
		if !isPacketValid(&pkt, isClientSide, g.SimpleMode) {
			// Error: ignore the packet
			continue
		}

		// A client should never receive a StatusPacket
		if isClientSide {

			if g.SimpleMode { // Simple mode
				OnReceiveBroadcast(g, channel, pkt.SimpleMsg)
			} else {
				// Convert the message to a RumorMessage
				rumor := utils.RumorMessage{Text: pkt.SimpleMsg.Contents}
				go utils.OnReceiveRumor(g, channel, &rumor, sender, isClientSide)
			}

		} else {

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
					OnReceiveBroadcast(g, channel, pkt.SimpleMsg)
				case pkt.Rumor != nil:
					go utils.OnReceiveRumor(g, channel, pkt.Rumor, sender, isClientSide)
				case pkt.Status != nil:
					go utils.OnReceiveStatus(g, channel, pkt.Status, sender)
				default:
					// Should never happen
				}
			}
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

	// fmt.Printf("%s\n", utils.GossiperToString(&gossiper))

	// Launch 2 threads for client-side and network-side communication
	go UDPDispatcher(&gossiper, gossiper.ClientAddr, utils.OnBroadcastClient, true)
	go UDPDispatcher(&gossiper, gossiper.GossipAddr, utils.OnBroadcastNetwork, false)

	// Kill all goroutines before exiting
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

}
