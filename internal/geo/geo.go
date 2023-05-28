// Package geo parses and returns a geonames.org geolocation data.
package geo

import (
	"encoding/csv"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Geo is the geolocation controller.
type Geo struct {
	locations []Location

	// { $keyword: { $timezone: $country_code }}
	tzMap map[string][]Location

	count int
}

// Location represents a geographic location.
type Location struct {
	ID         string
	Name       string
	Lat        float64
	Lon        float64
	Timezone   string
	Country    string
	Population int

	Loc *time.Location
}

var (
	reClean = regexp.MustCompile("[^a-z/]+")
)

// New initiates a new geo location map.
func New(filePath string) (*Geo, error) {
	g := &Geo{
		tzMap: make(map[string][]Location),
	}

	locs, err := g.readFile(filePath)
	if err != nil {
		return nil, err
	}

	g.load(locs)
	return g, nil
}

// Query queries a loaded geo location by the given keyword.
func (g *Geo) Query(q string) []Location {
	// If there's a country code, separate it.
	var (
		str     = strings.Split(q, "/")
		country = ""
	)
	if len(str) == 2 && len(str[1]) == 2 {
		q = str[0]
		country = strings.ToUpper(str[1])
	}

	q = reClean.ReplaceAllString(strings.ToLower(q), "")
	zones, ok := g.tzMap[q]
	if !ok {
		return nil
	}

	// Filter by country.
	if country != "" {
		out := make([]Location, 0, len(zones))
		for _, z := range zones {
			if z.Country != country {
				continue
			}

			out = append(out, z)
		}

		return out
	}

	return zones
}

// Count returns the number of unique locations loaded.
func (g *Geo) Count() int {
	return g.count
}

func (g *Geo) load(locs []Location) {
	for _, l := range locs {
		// Add the city name.
		name := reClean.ReplaceAllString(strings.ToLower(l.Name), "")

		if _, ok := g.tzMap[name]; !ok {
			g.tzMap[name] = []Location{}
		}

		g.tzMap[name] = append(g.tzMap[name], l)

		g.count++
	}

	// Cities in timezone names that don't exist in the map, add to the map.
	for _, l := range locs {
		city := reClean.ReplaceAllString(strings.Split(l.Timezone, "/")[1], "")
		_, ok := g.tzMap[city]
		if !ok {
			g.tzMap[city] = []Location{l}
		}
	}

	// Sort cities in the collated map by population under the assumption
	// that bigger cities are likely to be searched more.
	for _, locs := range g.tzMap {
		sort.Slice(locs, func(i, j int) bool {
			return locs[i].Population > locs[j].Population
		})
	}
}

// readFile loads a geonames.org geolocation file and returns the list
// of parses Locations.
func (g *Geo) readFile(filePath string) ([]Location, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rd := csv.NewReader(f)
	rd.Comma = '\t'

	out := []Location{}
	for {
		r, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if len(r) != 19 {
			continue
		}

		// Create the location record.
		var (
			lat, _ = strconv.ParseFloat(r[4], 32)
			lon, _ = strconv.ParseFloat(r[5], 32)
			pop, _ = strconv.Atoi(r[14])
		)

		// Remove values in brackets.
		r[2] = strings.TrimSpace(strings.Split(r[2], "(")[0])

		loc, err := time.LoadLocation(r[17])
		if err != nil {
			continue
		}

		out = append(out, Location{
			ID:         r[0],
			Name:       r[2],
			Lat:        lat,
			Lon:        lon,
			Country:    r[8],
			Timezone:   r[17],
			Population: pop,
			Loc:        loc,
		})
	}

	return out, nil
}
