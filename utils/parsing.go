package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParsePort(s string) error {
	if port, err := strconv.ParseInt(s, 10, 16); err != nil || port < 1024 || port > 65535 {
		return &CustomError{"parsePort", "failed to parse PORT number"}
	}
	return nil
}

func CheckIPPortPair(s string) error {

	slices := strings.Split(s, ":")
	if len(slices) != 2 {
		return &CustomError{"checkIPPortPair", "failed to separate IP from PORT"}
	}

	// Parse the port number
	if err := ParsePort(slices[1]); err != nil {
		return &CustomError{"checkIPPortPair", "failed to parse PORT number"}
	}

	// Parse the ip
	slicesIP := strings.Split(slices[0], ".")
	if len(slicesIP) != 4 {
		return &CustomError{"checkIPPortPair", "IP doesn't have 4 components"}
	}

	for _, x := range slicesIP {
		if n, err := strconv.ParseInt(x, 10, 8); err != nil || n > 255 || n < 0 {
			return &CustomError{"checkIPPortPair", "IP component not in range [0, 256)"}
		}
	}
	return nil
}

func (network *GossipNetwork) ParsePeers(s string) error {

	slices := strings.Split(s, ",")
	for _, rawAddr := range slices {

		// Check that the IP has a correct format
		if err := CheckIPPortPair(rawAddr); err != nil {
			return &CustomError{"parsePeers", "failed to parse peers IP/PORT pairs"}
		}

		var peer Peer
		if err := peer.CreatePeer(rawAddr); err != nil {
			return &CustomError{"parsePeers", "failed to create new peer"}
		}

		network.Peers = append(network.Peers, peer)
	}
	return nil
}

func (g *Gossiper) ParseArgumentsGossiper() error {

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if g.ClientAddr != "" {
				return &CustomError{"parseArgumentsGossiper", "UIPort defined twice"}
			}
			err := ParsePort(arg[8:])
			if err != nil {
				return &CustomError{"parseArgumentsGossiper", "unable to parse UIPort"}
			}
			g.ClientAddr = fmt.Sprintf("127.0.0.1:%s", arg[8:])

		case strings.HasPrefix(arg, "-gossipAddr="):
			if g.GossipAddr != "" {
				return &CustomError{"parseArgumentsGossiper", "gossipAddr defined twice"}
			}
			err := CheckIPPortPair(arg[12:])
			if err != nil {
				return &CustomError{"parseArgumentsGossiper", "unable to parse gossipAddr"}
			}
			g.GossipAddr = arg[12:]

		case strings.HasPrefix(arg, "-name="):
			if g.Name != "" {
				return &CustomError{"parseArgumentsGossiper", "name defined twice"}
			}
			if len(arg) == 6 {
				return &CustomError{"parseArgumentsGossiper", "name is empty"}
			}
			g.Name = arg[6:]

		case strings.HasPrefix(arg, "-peers="):
			if len(g.Network.Peers) != 0 {
				return &CustomError{"parseArgumentsGossiper", "peers defined twice"}
			}

			if err := g.Network.ParsePeers(arg[7:]); err != nil {
				return &CustomError{"parseArgumentsGossiper", "unable to parse peers"}
			}

		case strings.HasPrefix(arg, "-simple"):
			g.SimpleMode = true
		default:
			return &CustomError{"parseArgumentsGossiper", "unknown argument"}
		}
	}

	// The gossiper must have a name
	if g.Name == "" {
		return &CustomError{"parseArgumentsGossiper", "the gossiper has no name"}
	}

	// Create default values for missing parameters
	if g.ClientAddr != "" {
		g.ClientAddr = "127.0.0.1:8080"
	}
	if g.GossipAddr != "" {
		g.GossipAddr = "127.0.0.1:5000"
	}

	return nil
}

func (c *Client) ParseArgumentsClient() error {
	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if c.Addr != "" {
				return &CustomError{"parseArgumentsClient", "UIPort defined twice"}
			}
			err := ParsePort(arg[8:])
			if err != nil {
				return &CustomError{"parseArgumentsClient", "unable to parse UIPort"}
			}
			c.Addr = fmt.Sprintf("127.0.0.1:%s", arg[8:])

		case strings.HasPrefix(arg, "-msg"):
			if c.Msg != "" {
				return &CustomError{"parseArgumentsClient", "msg defined twice"}
			}
			c.Msg = arg[4:]
		}
	}

	// The client must have a message
	if c.Msg == "" {
		return &CustomError{"parseArgumentsClient", "the client has no message to transmit"}
	}

	// Create default values for missing parameters
	if c.Addr != "" {
		c.Addr = "127.0.0.1:8080"
	}

	return nil
}
