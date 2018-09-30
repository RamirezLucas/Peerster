package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

/* ======== TYPE DECLARATIONS ======== */
type ipPortPair struct {
	ip   [4]byte
	port int
}

type clArgs struct {
	uiPort     int
	gossipAddr ipPortPair
	name       string
	peers      []ipPortPair
	simpleMode bool
}

type customError struct {
	fun  string
	desc string
}

/* ======== INTERFACE FUNCTIONS ======== */
func (pair ipPortPair) String() string {
	return fmt.Sprintf("%v.%v.%v.%v:%v", pair.ip[0], pair.ip[1],
		pair.ip[2], pair.ip[3], pair.port)
}

func (cli clArgs) String() string {
	acc := fmt.Sprintf("UIPort: %v\nname: %v\nsimpleMode: %v\ngossipAddr: %v\npeers:\n", cli.uiPort, cli.name, cli.simpleMode, cli.gossipAddr)
	for _, x := range cli.peers {
		acc = acc + fmt.Sprintf("\t%v\n", x)
	}
	return acc
}

func (e *customError) Error() string {
	return fmt.Sprintf("Error in %s(): %s", e.fun, e.desc)
}

/* ======== MAIN CODE ======== */
func parsePort(s string) (int, error) {
	port, errPort := strconv.ParseInt(s, 10, 16)
	if errPort != nil || port < 1024 || port > 65535 {
		return int(port), &customError{"parsePort", "failed to parse PORT number"}
	}
	return int(port), nil
}

func parseIPPortPair(s string) (ipPortPair, error) {

	ipPort := ipPortPair{}

	slices := strings.Split(s, ":")
	if len(slices) != 2 {
		return ipPort, &customError{"parseIPPortPair", "failed to separate IP from PORT"}
	}

	// Parse the port number
	port, errPort := parsePort(slices[1])
	if errPort != nil {
		return ipPort, &customError{"parseIPPortPair", "failed to parse PORT number"}
	}
	ipPort.port = port

	// Parse the ip
	slicesIP := strings.Split(slices[0], ".")
	if len(slicesIP) != 4 {
		return ipPort, &customError{"parseIPPortPair", "IP doesn't have 4 components"}
	}

	for i, x := range slicesIP {
		n, errIP := strconv.ParseInt(x, 10, 8)
		if errIP != nil || n > 255 || n < 0 {
			return ipPort, &customError{"parseIPPortPair", "IP component not in range [0, 256)"}
		}
		ipPort.ip[i] = byte(n)
	}

	return ipPort, nil
}

func parsePeers(s string) ([]ipPortPair, error) {

	var pairs []ipPortPair

	slices := strings.Split(s, ",")
	for _, pair := range slices {
		pair, err := parseIPPortPair(pair)
		if err != nil {
			return pairs, &customError{"parsePeers", "failed to parse peers IP/PORT pairs"}
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func parseArguments() (clArgs, error) {

	cli := clArgs{}

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "-UIPort="):
			if cli.uiPort != 0 {
				return cli, &customError{"parseArguments", "UIPort defined twice"}
			}
			port, err := parsePort(arg[8:])
			if err != nil {
				return cli, &customError{"parseArguments", "unable to parse UIPort"}
			}
			cli.uiPort = port
		case strings.HasPrefix(arg, "-gossipAddr="):
			if cli.gossipAddr.port != 0 {
				return cli, &customError{"parseArguments", "gossipAddr defined twice"}
			}
			gossipPair, err := parseIPPortPair(arg[12:])
			if err != nil {
				return cli, &customError{"parseArguments", "unable to parse gossipAddr"}
			}
			cli.gossipAddr = gossipPair
		case strings.HasPrefix(arg, "-name="):
			if cli.name != "" {
				return cli, &customError{"parseArguments", "name defined twice"}
			}
			if len(arg) == 6 {
				return cli, &customError{"parseArguments", "name is empty"}
			}
			cli.name = arg[6:]
		case strings.HasPrefix(arg, "-peers="):
			if len(cli.peers) != 0 {
				return cli, &customError{"parseArguments", "peers defined twice"}
			}
			peersPairs, err := parsePeers(arg[7:])
			if err != nil {
				return cli, &customError{"parseArguments", "unable to parse peers"}
			}
			cli.peers = peersPairs
		case strings.HasPrefix(arg, "-simple"):
			cli.simpleMode = true
			// TODO: care about double definition ?
		default:
			return cli, &customError{"parseArguments", "unknown argument"}
		}
	}

	// The gossiper must have a name
	if cli.name == "" {
		return cli, &customError{"parseArguments", "the gossiper has no name"}
	}

	// Create default values for missing parameters
	if cli.uiPort == 0 {
		cli.uiPort = 8080
	}
	if cli.gossipAddr.port == 0 {
		cli.gossipAddr.ip = [4]byte{127, 0, 0, 1}
		cli.gossipAddr.port = 5000
	}

	return cli, nil
}

func main() {

	for _, arg := range os.Args[1:] {
		fmt.Println(arg)
	}
	return
	cli, err := parseArguments()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cli)

}
