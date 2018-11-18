package peers

import (
	"fmt"
	"net"
)

// StringToUDPAddress - Returns a UDP address based on a string <ip:port>
func StringToUDPAddress(rawAddr string) *net.UDPAddr {
	addr, _ := net.ResolveUDPAddr("udp4", rawAddr)
	return addr
}

// UDPAddressToString - Returns a textual representation of net.UDPAddr
func UDPAddressToString(addr *net.UDPAddr) string {
	return fmt.Sprintf("%s", addr)
}

// CompareUDPAddress - Compares 2 UDP addresses
func CompareUDPAddress(a, b *net.UDPAddr) bool {
	return UDPAddressToString(a) == UDPAddressToString(b)
}
