package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/dedis/protobuf"
)

/* ======== TYPE DECLARATIONS ======== */
type ipPortPair struct {
	ip   [4]byte
	port int
}

type customError struct {
	fun  string
	desc string
}

// Gossiper -- Represent a gossiper
type Gossiper struct {
	clientAddr string     // IP/Port on which the client talks
	gossipAddr string     // IP/Port on which to listen to other gossips
	name       string     // Name of that gossiper
	peers      []string   // List of known peers
	simpleMode bool       // Indicate whether the gossiper operated in simple mode (broadcast)
	mux        sync.Mutex // Mutex to manipulate the structure from different threads
}

// SimpleMessage -- Represents a message
type SimpleMessage struct {
	originalName  string
	relayPeerAddr string
	contents      string
}

// GossipPacket -- Represents a gossip packet
type GossipPacket struct {
	msg *SimpleMessage
}

/* ======== INTERFACE FUNCTIONS ======== */
func (pair ipPortPair) String() string {
	return fmt.Sprintf("%v.%v.%v.%v:%v", pair.ip[0], pair.ip[1],
		pair.ip[2], pair.ip[3], pair.port)
}

func (g *Gossiper) String() string {
	acc := fmt.Sprintf("clientAddr: %v\nname: %v\nsimpleMode: %v\ngossipAddr: %v\npeers:\n", g.clientAddr, g.name, g.simpleMode, g.gossipAddr)
	for _, x := range g.peers {
		acc = acc + fmt.Sprintf("\t%v\n", x)
	}
	return acc
}

func (e *customError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.fun, e.desc)
}

/* ======== PARSING ======== */
func parsePort(s string) error {
	if port, err := strconv.ParseInt(s, 10, 16); err != nil || port < 1024 || port > 65535 {
		return &customError{"parsePort", "failed to parse PORT number"}
	}
	return nil
}

func checkIPPortPair(s string) error {

	slices := strings.Split(s, ":")
	if len(slices) != 2 {
		return &customError{"checkIPPortPair", "failed to separate IP from PORT"}
	}

	// Parse the port number
	if err := parsePort(slices[1]); err != nil {
		return &customError{"checkIPPortPair", "failed to parse PORT number"}
	}

	// Parse the ip
	slicesIP := strings.Split(slices[0], ".")
	if len(slicesIP) != 4 {
		return &customError{"checkIPPortPair", "IP doesn't have 4 components"}
	}

	for _, x := range slicesIP {
		if n, err := strconv.ParseInt(x, 10, 8); err != nil || n > 255 || n < 0 {
			return &customError{"checkIPPortPair", "IP component not in range [0, 256)"}
		}
	}
	return nil
}

func parsePeers(s string) ([]string, error) {

	var pairs []string

	slices := strings.Split(s, ",")
	for _, pair := range slices {
		if err := checkIPPortPair(pair); err != nil {
			return pairs, &customError{"parsePeers", "failed to parse peers IP/PORT pairs"}
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func (g *Gossiper) parseArguments() error {

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if g.clientAddr != "" {
				return &customError{"parseArguments", "UIPort defined twice"}
			}
			err := parsePort(arg[8:])
			if err != nil {
				return &customError{"parseArguments", "unable to parse UIPort"}
			}
			g.clientAddr = fmt.Sprintf("127.0.0.1:%s", arg[8:])

		case strings.HasPrefix(arg, "-gossipAddr="):
			if g.gossipAddr != "" {
				return &customError{"parseArguments", "gossipAddr defined twice"}
			}
			err := checkIPPortPair(arg[12:])
			if err != nil {
				return &customError{"parseArguments", "unable to parse gossipAddr"}
			}
			g.gossipAddr = arg[12:]

		case strings.HasPrefix(arg, "-name="):
			if g.name != "" {
				return &customError{"parseArguments", "name defined twice"}
			}
			if len(arg) == 6 {
				return &customError{"parseArguments", "name is empty"}
			}
			g.name = arg[6:]

		case strings.HasPrefix(arg, "-peers="):
			if len(g.peers) != 0 {
				return &customError{"parseArguments", "peers defined twice"}
			}
			peersPairs, err := parsePeers(arg[7:])
			if err != nil {
				return &customError{"parseArguments", "unable to parse peers"}
			}
			g.peers = peersPairs

		case strings.HasPrefix(arg, "-simple"):
			g.simpleMode = true
			// TODO: care about double definition ?
		default:
			return &customError{"parseArguments", "unknown argument"}
		}
	}

	// The gossiper must have a name
	if g.name == "" {
		return &customError{"parseArguments", "the gossiper has no name"}
	}

	// Create default values for missing parameters
	if g.clientAddr != "" {
		g.clientAddr = "127.0.0.1:8080"
	}
	if g.gossipAddr != "" {
		g.gossipAddr = "127.0.0.1:5000"
	}

	return nil
}

/* ======== NETWORK ======== */
func openUDPChannel(s string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return nil, &customError{"openUDPChannel", "cannot resolve UDP address"}
	}
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, &customError{"openUDPChannel", "cannot listen on UDP channel"}
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

	buf := make([]byte, 1024)

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
		return &customError{"callbackClient", "failed to encode packet"}
	}

	// Send to everyone
	g.mux.Lock() // Lock the gossiper because we are accessing peers
	defer g.mux.Unlock()

	for _, peer := range g.peers {
		// TODO: remove code copy
		udpAddr, err := net.ResolveUDPAddr("udp4", peer)
		if err != nil {
			return &customError{"callbackClient", "unable to resolve UDP address"}
		}
		if _, err = udpChannel.WriteToUDP(buf, udpAddr); err != nil {
			return &customError{"callbackClient", "unable to write on UDP channel"}
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
		return &customError{"callbackPeer", "failed to encode packet"}
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
				return &customError{"callbackPeer", "unable to resolve UDP address"}
			}
			if _, err = udpChannel.WriteToUDP(buf, udpAddr); err != nil {
				return &customError{"callbackPeer", "unable to write on UDP channel"}
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

	if err := gossiper.parseArguments(); err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println(&gossiper)

	go gossiper.listenUDPChannel(gossiper.clientAddr, callbackClient)
	go gossiper.listenUDPChannel(gossiper.gossipAddr, callbackClient)

}
