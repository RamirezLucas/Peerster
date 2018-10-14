package parsing

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// parsePort - Checks that a port number is correctly formed
func parsePort(s string) error {
	if port, err := strconv.ParseInt(s, 10, 32); err != nil || port < 1024 || port > 65535 {
		return &fail.CustomError{Fun: "parsePort", Desc: "failed to parse PORT number"}
	}
	return nil
}

// checkIPPortPair - Checks that a pair <ip:port> is correctly formed
func checkIPPortPair(s string) error {

	slices := strings.Split(s, ":")
	if len(slices) != 2 {
		return &fail.CustomError{Fun: "checkIPPortPair", Desc: "failed to separate IP from PORT"}
	}

	// Parse the port number
	if err := parsePort(slices[1]); err != nil {
		return &fail.CustomError{Fun: "checkIPPortPair", Desc: "failed to parse PORT number"}
	}

	// Parse the ip
	slicesIP := strings.Split(slices[0], ".")
	if len(slicesIP) != 4 {
		return &fail.CustomError{Fun: "checkIPPortPair", Desc: "IP doesn't have 4 components"}
	}

	for _, x := range slicesIP {
		if n, err := strconv.ParseInt(x, 10, 32); err != nil || n > 255 || n < 0 {
			fmt.Println(n)
			return &fail.CustomError{Fun: "checkIPPortPair", Desc: "IP component not in range [0, 256)"}
		}
	}
	return nil
}

// parsePeers - Parses a list of <ip:port>,
func parsePeers(peerIndex *types.PeerIndex, s string) error {

	slices := strings.Split(s, ",")
	for _, rawAddr := range slices {

		// Add the peer
		if udpAddr, err := net.ResolveUDPAddr("udp4", rawAddr); err == nil {
			peerIndex.AddPeerIfAbsent(udpAddr)
		} else {
			return &fail.CustomError{Fun: "parsePeers", Desc: "unable to parse peer"}
		}

	}
	return nil
}

// ParseArgumentsGossiper - Parses the arguments for the gossiper
func ParseArgumentsGossiper(g *types.Gossiper) error {

	var uiPortDone, gossipAddrDone, nameDone, peersDone bool

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if uiPortDone {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "UIPort defined twice"}
			}
			err := parsePort(arg[8:])
			if err != nil {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse UIPort"}
			}
			g.ClientAddr = fmt.Sprintf("127.0.0.1:%s", arg[8:])
			uiPortDone = true
		case strings.HasPrefix(arg, "-gossipAddr="):
			if gossipAddrDone {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "gossipAddr defined twice"}
			}
			err := checkIPPortPair(arg[12:])
			if err != nil {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse gossipAddr"}
			}
			g.GossipAddr = arg[12:]
			gossipAddrDone = true
		case strings.HasPrefix(arg, "-name="):
			if nameDone {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "name defined twice"}
			}
			if len(arg) == 6 {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "name is empty"}
			}
			g.Name = arg[6:]
			nameDone = true
		case strings.HasPrefix(arg, "-peers="):
			if peersDone {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "peers defined twice"}
			}

			if err := parsePeers(g.PeerIndex, arg[7:]); err != nil {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse peers"}
			}
			peersDone = true
		case strings.HasPrefix(arg, "-simple"):
			g.SimpleMode = true
		default:
			return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unknown argument"}
		}
	}

	// The gossiper must have a name
	if !nameDone {
		return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "the gossiper has no name"}
	}

	// Create default values for missing parameters
	if !uiPortDone {
		g.ClientAddr = "127.0.0.1:8080"
	}
	if !gossipAddrDone {
		g.GossipAddr = "127.0.0.1:5000"
	}

	return nil
}

// ParseArgumentsClient - Parses the arguments for the client
func ParseArgumentsClient(c *types.Client) error {

	var uiPortDone, msgDone bool

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if uiPortDone {
				return &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "UIPort defined twice"}
			}
			err := parsePort(arg[8:])
			if err != nil {
				return &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "unable to parse UIPort"}
			}

			// Resolve the address
			udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("127.0.0.1:%s", arg[8:]))
			if err != nil {
				return &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "cannot resolve UDP address"}
			}
			c.Addr = udpAddr
			uiPortDone = true
		case strings.HasPrefix(arg, "-msg="):
			if msgDone {
				return &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "msg defined twice"}
			}
			c.Msg = arg[5:]
			msgDone = true
		}
	}

	// The client must have a message
	if !msgDone {
		return &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "the client has no message to transmit"}
	}

	// Create default values for missing parameters
	if !uiPortDone {
		udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8080")
		if err != nil {
			return &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "cannot resolve UDP address"}
		}
		c.Addr = udpAddr
	}

	return nil
}
