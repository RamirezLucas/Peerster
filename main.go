package main

import (
	"Peerster/backend"
	"Peerster/entities"
	"Peerster/fail"
	"Peerster/messages"
	"Peerster/network"
	"Peerster/parsing"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dedis/protobuf"
)

// BufSize - Size of the UDP buffer
const BufSize = 16384

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

func isPacketValid(pkt *messages.GossipPacket, isClientSide bool, isSimpleMode bool) bool {

	// Exactly one field of the packet must be non-nil
	counter := 0
	if pkt.SimpleMsg != nil {
		counter++
	}
	if pkt.Rumor != nil {
		counter++
	}
	if pkt.Status != nil {
		counter++
	}
	if pkt.Private != nil {
		counter++
	}
	if pkt.DataRequest != nil {
		counter++
	}
	if pkt.DataReply != nil {
		counter++
	}
	if pkt.SearchRequest != nil {
		counter++
	}
	if pkt.SearchReply != nil {
		counter++
	}
	if pkt.TxPublish != nil {
		counter++
	}
	if pkt.BlockPublish != nil {
		counter++
	}
	if counter != 1 {
		return false
	}

	// The client only sends SimpleMessage and PrivateMessage and DataRequest and SearchRequest
	if isClientSide && (pkt.SimpleMsg == nil && pkt.Private == nil &&
		pkt.DataRequest == nil && pkt.SearchRequest == nil) {
		return false
	}
	// In simple mode only accept simple messages
	if isSimpleMode && pkt.SimpleMsg == nil {
		return false
	}

	return true
}

func antiEntropy(g *entities.Gossiper) {

	// Create a timeout timer
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
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

func rumorEntropy(g *entities.Gossiper, chanID chan uint32) {

	go network.OnSendRouteRumor(g, <-chanID)

	// Create a timeout timer
	timer := time.NewTicker(time.Duration(g.Args.RTimer) * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			go network.OnSendRouteRumor(g, <-chanID)
		}
	}

}

func udpDispatcherGossip(g *entities.Gossiper, chanID chan uint32) {

	for {

		// Create a buffer to store arriving data
		buf := make([]byte, BufSize)

		var sender *net.UDPAddr
		var n int
		var err error

		if n, sender, err = g.GossipChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt messages.GossipPacket
		if err := protobuf.Decode(buf[:n], &pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(&pkt, false, g.Args.SimpleMode) {
			// Error: ignore the packet
			continue
		}

		// Attempt to add the sending peer to the list of neighbors
		g.PeerIndex.AddPeerIfAbsent(sender)

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
			go network.OnReceiveRumor(g, pkt.Rumor, sender, <-chanID)
		case pkt.Status != nil:
			isPacketHandled := g.Timeouts.SearchAndForward(sender, pkt.Status)
			if !isPacketHandled {
				go network.OnReceiveStatus(g, pkt.Status, sender, <-chanID)
			}
		case pkt.Private != nil:
			go network.OnReceivePrivate(g, pkt.Private, sender)
		case pkt.DataRequest != nil:
			go network.OnReceiveDataRequest(g, pkt.DataRequest, sender)
		case pkt.DataReply != nil:
			go network.OnReceiveDataReply(g, pkt.DataReply, sender)
		case pkt.SearchRequest != nil:
			go network.OnReceiveSearchRequest(g, pkt.SearchRequest, sender)
		case pkt.SearchReply != nil:
			go network.OnReceiveSearchReply(g, pkt.SearchReply, sender)
		case pkt.TxPublish != nil:
			go network.OnReceiveTransaction(g, pkt.TxPublish, sender)
		case pkt.BlockPublish != nil:
			go network.OnReceiveBlock(g, pkt.BlockPublish, sender)
		default:
			// Should never happen
		}

	}
}

func udpDispatcherClient(g *entities.Gossiper, chanID chan uint32) {

	for {

		// Create a buffer to store arriving data
		buf := make([]byte, BufSize)

		var n int
		var err error
		if n, _, err = g.ClientChannel.ReadFromUDP(buf); err != nil {
			// Error: ignore the packet
			continue
		}

		// Decode the packet
		var pkt messages.GossipPacket
		if err := protobuf.Decode(buf[:n], &pkt); err != nil {
			// Error: ignore the packet
			continue
		}

		// Check the packet's validity
		if !isPacketValid(&pkt, true, g.Args.SimpleMode) {
			// Error: ignore the packet
			continue
		}

		switch {
		case pkt.SimpleMsg != nil:

			if g.Args.SimpleMode { // Simple mode
				go network.OnBroadcastClient(g, pkt.SimpleMsg)
			} else {
				// Promote SimpleMessage to RumorMessage
				rumor := messages.RumorMessage{Text: pkt.SimpleMsg.Contents}
				go network.OnReceiveClientRumor(g, &rumor, <-chanID)
			}

		case pkt.Private != nil:
			go network.OnReceiveClientPrivate(g, pkt.Private)
		case pkt.DataRequest != nil:

			if pkt.DataRequest.HopLimit == 0 {
				// File index
				if file := g.FileIndex.AddLocalFile(pkt.DataRequest.Origin); file != nil {
					fail.LeveledPrint(1, "udpDispatcherClient", "Adding transaction to blockchain")
					// Broadcast the transaction and publish to the blockchain
					network.OnReceiveTransaction(g, &messages.TxPublish{
						File:     *file,
						HopLimit: network.TransactionHopLimit,
					}, nil)
				}
			} else {
				// Remote file request
				if pkt.DataRequest.Destination == "" {
					go network.OnRemoteMetafileRequestMultisource(g, pkt.DataRequest.HashValue, pkt.DataRequest.Origin)
				} else {
					go network.OnRemoteMetafileRequestMonosource(g, pkt.DataRequest.HashValue,
						pkt.DataRequest.Origin, pkt.DataRequest.Destination)
				}

			}

		case pkt.SearchRequest != nil:
			go network.OnInitiateFileSearch(g, pkt.SearchRequest.Budget, pkt.SearchRequest.Keywords)
		default:
			// Should never happen
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
	gossiper := entities.NewGossiper(args)

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

	if !gossiper.Args.SimpleMode {
		// Anti Entropy
		//go antiEntropy(gossiper)

		// RouteRumor
		if gossiper.Args.RTimer != 0 {
			go rumorEntropy(gossiper, chanID)
		}
	}

	// Kill all goroutines before exiting
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

}
