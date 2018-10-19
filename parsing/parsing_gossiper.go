package parsing

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseArgumentsGossiper - Parses the arguments for the gossiper
func ParseArgumentsGossiper() (*types.CLArgsGossiper, error) {

	var args types.CLArgsGossiper

	var uiPortDone, guiPortDone, gossipAddrDone, nameDone, peersDone, simpleDone, rTimerDone bool

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if uiPortDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "UIPort defined twice"}
			}
			if err := parsePort(arg[8:]); err != nil {
				fmt.Println(err)
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse UIPort"}
			}

			// Validate
			args.ClientAddr = fmt.Sprintf("127.0.0.1:%s", arg[8:])
			uiPortDone = true

		case strings.HasPrefix(arg, "-GUIPort="):
			if guiPortDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "GUIPort defined twice"}
			}
			if err := parsePort(arg[9:]); err != nil {
				fmt.Println(err)
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse GUIPort"}
			}

			// Validate
			args.ServerPort = arg[9:]
			guiPortDone = true

		case strings.HasPrefix(arg, "-gossipAddr="):
			if gossipAddrDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "gossipAddr defined twice"}
			}
			if err := parseIPPortPair(arg[12:]); err != nil {
				fmt.Println(err)
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse gossipAddr"}
			}

			// Validate
			args.GossipAddr = arg[12:]
			gossipAddrDone = true

		case strings.HasPrefix(arg, "-name="):
			if nameDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "name defined twice"}
			}
			if len(arg) == 6 {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "name is empty"}
			}

			// Validate
			args.Name = arg[6:]
			nameDone = true

		case strings.HasPrefix(arg, "-peers="):
			if peersDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "peers defined twice"}
			}
			if err := parsePeers(&args.Peers, arg[7:]); err != nil {
				fmt.Println(err)
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unable to parse peers"}
			}

			// Validate
			peersDone = true

		case strings.HasPrefix(arg, "-simple"):
			if simpleDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "simple flag defined twice"}
			}

			// Validate
			args.SimpleMode = true
			simpleDone = true
		case strings.HasPrefix(arg, "-rtimer="):
			if rTimerDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "rTimer defined twice"}
			}

			timer, err := strconv.ParseInt(arg[8:], 10, 32)
			if err != nil || timer < 0 {
				fmt.Println(err)
				return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "rTimer invalid"}
			}

			// Validate
			args.RTimer = uint(timer)
			rTimerDone = true

		default:
			return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "unknown argument"}
		}
	}

	// The gossiper must have a name
	if !nameDone {
		return nil, &fail.CustomError{Fun: "ParseArgumentsGossiper", Desc: "the gossiper has no name"}
	}

	// Create default values for missing parameters
	if !uiPortDone {
		args.ClientAddr = "127.0.0.1:8080"
	}
	if !guiPortDone {
		args.ServerPort = "8080"
	}
	if !gossipAddrDone {
		args.GossipAddr = "127.0.0.1:5000"
	}
	if !rTimerDone {
		args.RTimer = 0
	}

	return &args, nil
}
