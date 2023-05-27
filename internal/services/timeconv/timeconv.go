package timeconv

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
)

const TimeFormat = "2006-01-02T15:04:05.9999"
const outputLayout = "15:04"
const queryRegex = `^(\d{1,2}:\d{2})([a-zA-Z //]+)-([a-zA-Z //]+)$`

// Timezones controller returns times for various geographic locations.
type Timeconv struct{
	geo *geo.Geo
}

// Opt contains config options for the Time package.
type Opt struct{}

// New returns a new instance of Timeconv
func New(o Opt, g *geo.Geo) *Timeconv {
	return &Timeconv{
		geo: g,
	}
}

//convertTime converts the time from one timezone to another
func (t *Timeconv) convertTime(input string) (string, error) {
	var queryFormat = regexp.MustCompile(queryRegex)
	matches := queryFormat.FindStringSubmatch(strings.TrimSpace(input))
	if len(matches) != 4 {
		return "", errors.New("invalid input format")
	}

	timeToConvert, fromTimezone, toTimezone := matches[1], matches[2], matches[3]

	var (
		str1     = strings.Split(fromTimezone, "/")
		str2     = strings.Split(toTimezone, "/")
		country1 = ""
		country2 = ""
		fromLoc, toLoc *time.Location
	)

	// Is there a /2-letter-country-code?
	if len(str1) == 2 && len(str1[1]) == 2 {
		fromTimezone = strings.ToLower(str1[0])
		country1 = strings.ToUpper(str1[1])
	}
	if len(str2) == 2 && len(str2[1]) == 2 {
		toTimezone = strings.ToLower(str2[0])
		country2 = strings.ToUpper(str2[1])
	}

	locs1 := t.geo.Query(fromTimezone)
	if locs1 == nil {
		return "", errors.New("unknown city "+fromTimezone)
	}

	locs2 := t.geo.Query(toTimezone)
	if locs2 == nil {
		return "", errors.New("unknown city "+toTimezone)
	}
	
	for _, l := range locs1 {		
		if country1 != "" {
			if l.Country != country1 {
				continue
			}
		}
		var err error
		fromLoc, err = time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}
	}
	
	for _, l := range locs2 {
		if country2 != "" {
			if l.Country != country2 {
				continue
			}
		}
		var err error
		toLoc, err = time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}
	}
	
	timeToConvTarget:="2023-05-27T"+timeToConvert+":00.0000"
	convertedTime, err := time.ParseInLocation(TimeFormat, timeToConvTarget, fromLoc)
	if err != nil {
		return "", errors.New("invalid time format")
	}
	result := convertedTime.In(toLoc).Format(outputLayout)
	return result, nil
}

// Query returns the converted time in the required format
func (t *Timeconv) Query(q string) ([]string, error) {
	v, err := t.convertTime(q)
	if err != nil {
		return nil, err
	}
	s := fmt.Sprintf("%s 1 TXT \"%s\"", q, v)
	return []string{s}, nil
}

// Dump is not implemented in this package.
func (n *Timeconv) Dump() ([]byte, error) {
	return nil, nil
}