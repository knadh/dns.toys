package ip2asn

import (
	"fmt"

	"github.com/jamesog/iptoasn"
)

// TTL is set to 900 seconds (15 minutes).
const TTL = 900

type Ip2Asn struct{}

// Returns a new Ip2Asn
func New() *Ip2Asn {
	return &Ip2Asn{}
}

// Query to get ASN info from IP
func (asn *Ip2Asn) Query(q string) ([]string, error) {
	res, err := iptoasn.LookupIP(q)
	if err != nil {
		return nil, err
	}

	s := fmt.Sprintf("%v %s", res.ASNum, res.ASName)
	return []string{fmt.Sprintf(`%s %d TXT "%s"`, q, TTL, s)}, nil
}

// Dump produces a gob dump of the cached data.
func (asn *Ip2Asn) Dump() ([]byte, error) {
	return nil, nil
}
