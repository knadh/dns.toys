// package timezones returns times for various zones and geographic locatons.
package timezones

import (
	"errors"
	"fmt"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
)

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
		zone, err := time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}

		r := fmt.Sprintf("%s TXT \"%s (%s, %s)\" \"%s\"",
			q, l.Name, l.Timezone, l.Country, time.Now().In(zone).Format(time.RFC1123Z))

		out = append(out, r)
	}

	return out, nil
}
