package parsing

import (
	"Peerster/entities"
	"Peerster/fail"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// ParseArgumentsClient - Parses the arguments for the client
func ParseArgumentsClient() (*entities.Client, error) {

	var client entities.Client
	var uiPortDone, msgDone, destDone, fileDone, reqDone, keyDone, budgetDone bool

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if uiPortDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "UIPort defined twice"}
			}
			if err := parsePort(arg[8:]); err != nil {
				fmt.Println(err)
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "unable to parse UIPort"}
			}

			// Resolve the address
			udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("127.0.0.1:%s", arg[8:]))
			if err != nil {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "cannot resolve UDP address"}
			}

			// Validate
			client.Addr = udpAddr
			uiPortDone = true

		case strings.HasPrefix(arg, "-msg="):
			if msgDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "msg defined twice"}
			}

			// Validate
			client.Msg = arg[5:]
			msgDone = true
		case strings.HasPrefix(arg, "-dest="):
			if destDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "msg defined twice"}
			}

			// Validate
			client.Dst = arg[6:]
			destDone = true
		case strings.HasPrefix(arg, "-file="):
			if fileDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "file defined twice"}
			}

			// Validate
			client.Filename = arg[6:]
			fileDone = true
		case strings.HasPrefix(arg, "-request="):
			if reqDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "request defined twice"}
			}

			// Validate
			if decoded, err := hex.DecodeString(arg[9:]); err == nil && len(decoded) == 32 {
				client.Request = decoded
			} else {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "hash isn't 32 bytes long"}
			}
			reqDone = true
		case strings.HasPrefix(arg, "-keywords="):
			if keyDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "keywords defined twice"}
			}

			// Validate
			client.Keywords = strings.Split(arg[10:], ",")
			keyDone = true
		case strings.HasPrefix(arg, "-budget="):
			if budgetDone {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "budget defined twice"}
			}

			// Validate
			if parsed, err := strconv.ParseInt(arg[8:], 10, 32); err == nil {
				client.Budget = uint64(parsed)
			} else {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "cannot parsed budget"}
			}
			budgetDone = true
		default:
			return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "unknown argument"}
		}
	}

	// The client must have a message
	if !msgDone && !fileDone && !keyDone {
		return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "the client has nothing to do"}
	}

	// Create default values for missing parameters
	if !uiPortDone {
		udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8080")
		if err != nil {
			return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "cannot resolve UDP address"}
		}
		client.Addr = udpAddr
	}
	if !reqDone {
		client.Request = nil
	}
	if !budgetDone {
		client.Budget = ^uint64(0)
	}

	return &client, nil
}
