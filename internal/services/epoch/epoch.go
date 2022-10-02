// converts an epoch/unix timestamp to human readable form.
package epoch

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Epoch struct{}

// New returns a new instance of CIDR.
func New() *Epoch {
	return &Epoch{}
}

// parses the query which is a epoch and returns it in human readable
func (n *Epoch) Query(q string) ([]string, error) {
	ts, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return nil, errors.New("invalid epoch query")
	}

	if ts >= 1e16 || ts <= -1e16 {
		// Nanoseconds.
		ts = (ts / 1000000000)

	} else if ts >= 1e14 || ts <= -1e14 {
		// Microseconds
		ts = (ts / 1000000)

	} else if ts >= 1e11 || ts <= -3e10 {
		// Milliseconds
		ts = (ts / 1000)
	}

	var (
		utc   = time.Unix(ts, 0).UTC()
		local = time.Unix(ts, 0)

		out = fmt.Sprintf(`%s 1 TXT "%s" "%s"`, q, utc, local)
	)

	return []string{out}, nil
}

func (n *Epoch) Dump() ([]byte, error) {
	return nil, nil
}
