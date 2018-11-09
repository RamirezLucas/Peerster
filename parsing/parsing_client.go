package parsing

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"net"
	"os"
	"strings"
)

// ParseArgumentsClient - Parses the arguments for the client
func ParseArgumentsClient() (*types.Client, error) {

	var client types.Client
	var uiPortDone, msgDone, destDone, fileDone, reqDone bool

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
			client.Request = []byte(arg[9:])
			if len(client.Request) != 32 {
				return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "hash isn't 32 bytes long"}
			}
			reqDone = true
		default:
			return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "unknown argument"}
		}
	}

	// The client must have a message
	if !msgDone {
		return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "the client has no message to transmit"}
	}

	// Create default values for missing parameters
	if !uiPortDone {
		udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8080")
		if err != nil {
			return nil, &fail.CustomError{Fun: "ParseArgumentsClient", Desc: "cannot resolve UDP address"}
		}
		client.Addr = udpAddr
	}

	return &client, nil
}
