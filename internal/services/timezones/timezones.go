// package timezones returns times for various geographic locatons.
package timezones

import (
	"errors"
	"fmt"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
)

// Timezones controller returns times for various geographic locations.
type Timezones struct {
	geo *geo.Geo
}

// Opt contains config options for the Time package.
type Opt struct{}

// New returns a new instance of Time.
func New(o Opt, g *geo.Geo) *Timezones {
	return &Timezones{
		geo: g,
	}
}

// Query parses a given query string and returns the answer.
// For the time package, the query is a location name.
func (t *Timezones) Query(q string) ([]string, error) {
	locs := t.geo.Query(q)
	if locs == nil {
		return nil, errors.New("unknown city.")
	}

	out := make([]string, 0, len(locs))
	for _, l := range locs {
		r := fmt.Sprintf("%s 1 TXT \"%s (%s, %s)\" \"%s\"",
			q, l.Name, l.Timezone, l.Country, time.Now().In(l.Loc).Format(time.RFC1123Z))

		out = append(out, r)
	}

	return out, nil
}

// Dump produces a gob dump of the cached data.
func (t *Timezones) Dump() ([]byte, error) {
	return nil, nil
}
