package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parsePort(s string) error {
	if port, err := strconv.ParseInt(s, 10, 16); err != nil || port < 1024 || port > 65535 {
		return &CustomError{"parsePort", "failed to parse PORT number"}
	}
	return nil
}

func checkIPPortPair(s string) error {

	slices := strings.Split(s, ":")
	if len(slices) != 2 {
		return &CustomError{"checkIPPortPair", "failed to separate IP from PORT"}
	}

	// Parse the port number
	if err := parsePort(slices[1]); err != nil {
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

func parsePeers(s string) (GossipNetwork, error) {

	var network GossipNetwork

	slices := strings.Split(s, ",")
	for _, rawAddr := range slices {

		// Check that the IP has a correct format
		if err := checkIPPortPair(rawAddr); err != nil {
			return network, &CustomError{"parsePeers", "failed to parse peers IP/PORT pairs"}
		}

		var peer Peer
		if err := peer.CreatePeer(rawAddr); err != nil {
			return network, &CustomError{"parsePeers", "failed to create new peer"}
		}

		network.peers = append(network.peers, peer)
	}
	return network, nil
}

func (g *Gossiper) parseArgumentsGossiper() error {

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if g.clientAddr != "" {
				return &CustomError{"parseArgumentsGossiper", "UIPort defined twice"}
			}
			err := parsePort(arg[8:])
			if err != nil {
				return &CustomError{"parseArgumentsGossiper", "unable to parse UIPort"}
			}
			g.clientAddr = fmt.Sprintf("127.0.0.1:%s", arg[8:])

		case strings.HasPrefix(arg, "-gossipAddr="):
			if g.gossipAddr != "" {
				return &CustomError{"parseArgumentsGossiper", "gossipAddr defined twice"}
			}
			err := checkIPPortPair(arg[12:])
			if err != nil {
				return &CustomError{"parseArgumentsGossiper", "unable to parse gossipAddr"}
			}
			g.gossipAddr = arg[12:]

		case strings.HasPrefix(arg, "-name="):
			if g.name != "" {
				return &CustomError{"parseArgumentsGossiper", "name defined twice"}
			}
			if len(arg) == 6 {
				return &CustomError{"parseArgumentsGossiper", "name is empty"}
			}
			g.name = arg[6:]

		case strings.HasPrefix(arg, "-peers="):
			if len(g.network.peers) != 0 {
				return &CustomError{"parseArgumentsGossiper", "peers defined twice"}
			}
			peersPairs, err := parsePeers(arg[7:])
			if err != nil {
				return &CustomError{"parseArgumentsGossiper", "unable to parse peers"}
			}
			g.network = peersPairs

		case strings.HasPrefix(arg, "-simple"):
			g.simpleMode = true
		default:
			return &CustomError{"parseArgumentsGossiper", "unknown argument"}
		}
	}

	// The gossiper must have a name
	if g.name == "" {
		return &CustomError{"parseArgumentsGossiper", "the gossiper has no name"}
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

func (c *Client) parseArgumentsClient() error {
	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if c.addr != "" {
				return &CustomError{"parseArgumentsClient", "UIPort defined twice"}
			}
			err := parsePort(arg[8:])
			if err != nil {
				return &CustomError{"parseArgumentsClient", "unable to parse UIPort"}
			}
			c.addr = fmt.Sprintf("127.0.0.1:%s", arg[8:])

		case strings.HasPrefix(arg, "-msg"):
			if c.msg != "" {
				return &CustomError{"parseArgumentsClient", "msg defined twice"}
			}
			c.msg = arg[4:]
		}
	}

	// The client must have a message
	if c.msg == "" {
		return &CustomError{"parseArgumentsClient", "the client has no message to transmit"}
	}

	// Create default values for missing parameters
	if c.addr != "" {
		c.addr = "127.0.0.1:8080"
	}

	return nil
}
