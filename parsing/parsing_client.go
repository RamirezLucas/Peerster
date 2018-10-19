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
