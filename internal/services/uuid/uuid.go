package uuid

import (
	"fmt"
	"strconv"

	"github.com/gofrs/uuid"
)

type UUID struct {
	maxResults int
}

func New(maxResults int) *UUID {
	if maxResults < 1 {
		maxResults = 1
	}
	return &UUID{
		maxResults: maxResults,
	}
}

// Query returns a random UUID.
func (u *UUID) Query(q string) ([]string, error) {
	num := 1
	if q != ".uuid" {
		num, _ = strconv.Atoi(q)
		if num < 1 || num > u.maxResults {
			return nil, fmt.Errorf("provide 1-%d.uuid", u.maxResults)
		}
	}

	out := make([]string, 0, num)
	for n := 0; n < num; n++ {
		id, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, id.String()))
	}

	return out, nil
}

// Dump is not implemented in this package.
func (u *UUID) Dump() ([]byte, error) {
	return nil, nil
}
