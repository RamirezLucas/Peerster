package parsing

import (
	"Peerster/fail"
	"Peerster/types"
	"fmt"
	"net"
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

		if rawAddr == "" {
			return nil
		}

		// Add the peer
		if udpAddr, err := net.ResolveUDPAddr("udp4", rawAddr); err == nil {
			peerIndex.AddPeerIfAbsent(udpAddr)
		} else {
			return &fail.CustomError{Fun: "parsePeers", Desc: "unable to parse peer"}
		}

	}
	return nil
}
