// package timezones returns times for various zones and geographic locatons.
package timezones

import (
	"errors"
	"fmt"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
	"github.com/miekg/dns"
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
func (t *Timezones) Query(q string) ([]dns.RR, error) {
	locs := t.geo.Query(q)
	if locs == nil {
		return nil, errors.New("unknown city.")
	}

	out := make([]dns.RR, 0, len(locs))
	for _, l := range locs {
		zone, err := time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}

		now := time.Now().In(zone)
		rr, err := dns.NewRR(fmt.Sprintf("%s TXT \"%s (%s, %s)\" \"%s\"",
			q, l.Name, l.Timezone, l.Country, now.Format(time.RFC1123Z)))
		if err != nil {
			return nil, err
		}
		out = append(out, rr)
	}

	return out, nil
}
