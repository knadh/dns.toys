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

	timestamp, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return nil, errors.New("invalid epoch query")
	}

	if timestamp >= 1e16 || timestamp <= -1e16 {
		timestamp = (timestamp / 1000000000) // To handle Nanoseconds

	} else if timestamp >= 1e14 || timestamp <= -1e14 {
		timestamp = (timestamp / 1000000) // To handle Microseconds

	} else if timestamp >= 1e11 || timestamp <= -3e10 {
		timestamp = (timestamp / 1000) // To handle Milliseconds
	}
	resinutc := time.Unix(timestamp, 0).UTC()
	resinlocal := time.Unix(timestamp, 0)
	out := fmt.Sprintf(`%s 1 TXT "%s" "%s"`, q, resinutc, resinlocal)
	return []string{out}, nil
}

func (n *Epoch) Dump() ([]byte, error) {
	return nil, nil
}
