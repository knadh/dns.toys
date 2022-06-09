// package timezones returns times for various geographic locatons.
package timezones

import (
	"errors"
	"fmt"
	"strings"
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
	var (
		str     = strings.Split(q, "/")
		country = ""
	)

	// Is there a /2-letter-country-code?
	if len(str) == 2 && len(str[1]) == 2 {
		q = str[0]
		country = strings.ToUpper(str[1])
	}
	q = strings.ToLower(q)

	locs := t.geo.Query(q)
	if locs == nil {
		return nil, errors.New("unknown city.")
	}

	out := make([]string, 0, len(locs))
	for _, l := range locs {
		// Filter by country.
		if country != "" {
			if l.Country != country {
				continue
			}
		}

		zone, err := time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}

		r := fmt.Sprintf("%s 1 TXT \"%s (%s, %s)\" \"%s\"",
			q, l.Name, l.Timezone, l.Country, time.Now().In(zone).Format(time.RFC1123Z))

		out = append(out, r)
	}

	return out, nil
}

// Dump produces a gob dump of the cached data.
func (t *Timezones) Dump() ([]byte, error) {
	return nil, nil
}
