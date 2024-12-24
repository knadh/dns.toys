// package cidr returns usable IP range for a given IP Address prefix.
package cidr

import (
	"errors"
	"fmt"
	"math/big"
	"net"
)

type CIDR struct{}

// New returns a new instance of CIDR.
func New() *CIDR {
	return &CIDR{}
}

// TTL is set to 900 seconds (15 minutes).
const TTL = 900

// Query parses a given query string and returns the answer.
// For the cidr package, the query is an IP Address Prefix (CIDR notation).
func (c *CIDR) Query(q string) ([]string, error) {
	ipAddr, network, err := net.ParseCIDR(q)
	if err != nil {
		return nil, errors.New("invalid cidr notation.")
	}
	prefixLen, bits := network.Mask.Size()

	switch {
	// Handle ipv4.
	case ipAddr.To4() != nil:
		// Binary "AND" operation between the IP address and the subnet mask.
		for i := range network.IP.To4() {
			network.IP[i] &= network.Mask[i]
		}
		// Ignore the first IP as it's the base IP which is unusable.
		// If /31 assume a point-to-point // link and return the lower address.
		if prefixLen < 31 {
			network.IP[3]++
		}
		first := network.IP.To4().String()

		// Binary "OR" operation on the IP with the bitwise binary inverse of the subnet mask to the first IP address.
		for i := range network.IP.To4() {
			network.IP[i] |= ^network.Mask[i]
		}
		// Ignore the last IP as it's the broadcast IP which is unusable.
		// If /31 then assume a point-to-point link and return upper address.
		if prefixLen < 31 {
			network.IP[3]--
		}
		last := network.IP.To4().String()

		// Get the size of subnet.
		size := 1 << (uint64(bits) - uint64(prefixLen))

		r := fmt.Sprintf("%s %d TXT \"%s\" \"%s\" \"%d\"", q, TTL, first, last, size)
		return []string{r}, nil

	// Handle ipv6.
	case ipAddr.To16() != nil:
		for i := range network.IP.To16() {
			network.IP[i] &= network.Mask[i]
		}
		first := network.IP.To16().String()

		for i := range network.IP.To16() {
			network.IP[i] |= ^network.Mask[i]
		}
		last := network.IP.To16().String()

		// uint32 won't suffice for IPv6 prefixes lesser than /65.
		size := big.NewInt(1)
		size = size.Lsh(size, uint(bits-prefixLen))

		r := fmt.Sprintf("%s %d TXT \"%s\" \"%s\" \"%d\"", q, TTL, first, last, size)
		return []string{r}, nil

	default:
		return nil, errors.New("unable to parse ip.")
	}
}

// Dump produces a gob dump of the cached data.
func (c *CIDR) Dump() ([]byte, error) {
	return nil, nil
}
