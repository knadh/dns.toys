// converts an epoch/unix timestamp to human readable form.
package epoch

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Epoch struct {
	localTime bool
}

// New returns a new instance of CIDR.
func New(localTime bool) *Epoch {
	return &Epoch{localTime: localTime}
}

// parses the query which is a epoch and returns it in human readable
func (e *Epoch) Query(q string) ([]string, error) {
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
	)

	// TTL is set to 900 seconds (15 minutes).
	out := fmt.Sprintf(`%s 900 TXT "%s"`, q, utc)
	if e.localTime {
		out += ` "` + local.String() + `"`
	}

	return []string{out}, nil
}

func (e *Epoch) Dump() ([]byte, error) {
	return nil, nil
}
