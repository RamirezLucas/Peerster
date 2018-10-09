package types

import (
	"fmt"
	"net"
)

// UDPAddressToString - Returns a textual representation of net.UDPAddr
func UDPAddressToString(addr *net.UDPAddr) string {
	return fmt.Sprintf("%s", addr)
}

// CompareUDPAddress - Compares 2 UDP addresses
func CompareUDPAddress(a, b *net.UDPAddr) bool {
	return UDPAddressToString(a) == UDPAddressToString(b)
}
