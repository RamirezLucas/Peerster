package parsing

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"os"
	"strings"
)

// ParseArgumentsGossiper - Parses the arguments for the gossiper
func ParseArgumentsGossiper(g *types.Gossiper) error {

	var uiPortDone, guiPortDone, gossipAddrDone, nameDone, peersDone bool

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
		case strings.HasPrefix(arg, "-GUIPort="):
			if guiPortDone {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "GUIPort defined twice"}
			}
			err := parsePort(arg[9:])
			if err != nil {
				return &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse GUIPort"}
			}
			g.ServerPort = arg[9:]
			guiPortDone = true
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
	if !guiPortDone {
		g.ServerPort = "8080"
	}
	if !gossipAddrDone {
		g.GossipAddr = "127.0.0.1:5000"
	}

	return nil
}
