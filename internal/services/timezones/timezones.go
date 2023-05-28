// package timezones returns times for various geographic locatons.
package timezones

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
)

const (
	inFormat = "2006-01-02T15:04"
)

var (
	reqQuery = regexp.MustCompile(`(?i)(\d{4}\-\d{2}\-\d{2}T\d{2}:\d{2})\-([a-z/]+)\-([a-z/]+)`)
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
	// Convert from -> to time.
	if m := reqQuery.FindStringSubmatch(strings.TrimSpace(q)); len(m) == 4 {
		return t.convert(q, m)
	}

	// Get time from a timezone.
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

func (t *Timezones) convert(q string, m []string) ([]string, error) {
	// Timestamp, zone to convert from, zone to convert to.
	ts, fromGeo, toGeo := m[1], m[2], m[3]

	// Get one or more from->to time.Location zones.
	fromLocs := t.geo.Query(fromGeo)
	if len(fromLocs) == 0 {
		return nil, errors.New("unknown `from` city.")
	}

	toLocs := t.geo.Query(toGeo)
	if len(toLocs) == 0 {
		return nil, errors.New("unknown `from` city.")
	}

	var out []string
	for _, from := range fromLocs {
		tm, err := time.ParseInLocation(inFormat, ts, from.Loc)
		if err != nil {
			return nil, errors.New("invalid time format")
		}

		for _, to := range toLocs {
			r := fmt.Sprintf("%s 1 TXT \"%s (%s, %s) %s\" = \"%s (%s, %s) %s\"",
				q,
				from.Name, from.Timezone, from.Country, tm.In(from.Loc).Format(time.RFC1123Z),
				to.Name, to.Timezone, to.Country, tm.In(to.Loc).Format(time.RFC1123Z))

			out = append(out, r)
		}
	}

	return out, nil
}
