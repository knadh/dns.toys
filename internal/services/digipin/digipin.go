package digipin

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DIGIPIN_GRID maps coordinates to alphanumeric characters
var DIGIPIN_GRID = [4][4]string{
	{"F", "C", "9", "8"},
	{"J", "3", "2", "7"},
	{"K", "4", "5", "6"},
	{"L", "M", "P", "T"},
}

// Bounds for India Post's grid
const (
	minLat = 2.5
	maxLat = 38.5
	minLon = 63.5
	maxLon = 99.5
	TTL    = 900 // 15 minutes
)

type Digipin struct{}

var (
	reDigipinEncode = regexp.MustCompile(`^(-?\d+\.?\d*),(-?\d+\.?\d*)$`)
	reDigipinDecode = regexp.MustCompile(`^([FC98J327K456LMPT-]+)$`)
)

// New returns a new instance of Digipin service.
func New() *Digipin {
	return &Digipin{}
}

// Query handles digipin encoding/decoding queries.
func (d *Digipin) Query(q string) ([]string, error) {
	q = strings.ToUpper(q)

	// Try to decode digipin to lat,lng
	if matches := reDigipinDecode.FindStringSubmatch(q); matches != nil {
		lat, lng, err := GetLatLngFromDigiPin(matches[1])
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("lat,lng = %.6f,%.6f", lat, lng)
		r := fmt.Sprintf(`%s %d TXT "%s"`, q, TTL, result)
		return []string{r}, nil
	}

	// Try to encode lat,lng to digipin
	if matches := reDigipinEncode.FindStringSubmatch(q); matches != nil {
		lat, lng, err := ParseCoordinates(fmt.Sprintf("%s,%s", matches[1], matches[2]))
		if err != nil {
			return nil, err
		}

		digipin, err := GetDigiPin(lat, lng)
		if err != nil {
			return nil, err
		}

		result := fmt.Sprintf("digipin = %s", digipin)
		r := fmt.Sprintf(`%s %d TXT "%s"`, q, TTL, result)
		return []string{r}, nil
	}

	return nil, errors.New("invalid digipin format")
}

// Dump is not implemented for this service.
func (d *Digipin) Dump() ([]byte, error) {
	return nil, nil
}

// GetDigiPin encodes lat/lon into a 10-character DIGIPIN
func GetDigiPin(lat, lon float64) (string, error) {
	if lat < minLat || lat > maxLat {
		return "", errors.New("latitude out of range")
	}
	if lon < minLon || lon > maxLon {
		return "", errors.New("longitude out of range")
	}

	var digiPin strings.Builder
	minLatLocal, maxLatLocal := minLat, maxLat
	minLonLocal, maxLonLocal := minLon, maxLon

	for level := 1; level <= 10; level++ {
		latDiv := (maxLatLocal - minLatLocal) / 4
		lonDiv := (maxLonLocal - minLonLocal) / 4

		row := 3 - int((lat-minLatLocal)/latDiv)
		col := int((lon - minLonLocal) / lonDiv)

		// Clamp indices
		if row < 0 {
			row = 0
		} else if row > 3 {
			row = 3
		}
		if col < 0 {
			col = 0
		} else if col > 3 {
			col = 3
		}

		digiPin.WriteString(DIGIPIN_GRID[row][col])

		if level == 3 || level == 6 {
			digiPin.WriteString("-")
		}

		// Update bounds
		maxLatLocal = minLatLocal + latDiv*float64(4-row)
		minLatLocal = minLatLocal + latDiv*float64(3-row)
		minLonLocal = minLonLocal + lonDiv*float64(col)
		maxLonLocal = minLonLocal + lonDiv
	}

	return digiPin.String(), nil
}

// GetLatLngFromDigiPin decodes a DIGIPIN string into lat/lon center
func GetLatLngFromDigiPin(digiPin string) (float64, float64, error) {
	pin := strings.ReplaceAll(digiPin, "-", "")
	if len(pin) != 10 {
		return 0, 0, errors.New("invalid DIGIPIN length")
	}

	minLatLocal, maxLatLocal := minLat, maxLat
	minLonLocal, maxLonLocal := minLon, maxLon

	for i := 0; i < 10; i++ {
		char := string(pin[i])
		found := false
		var row, col int

		// Find char in DIGIPIN_GRID
		for r := 0; r < 4; r++ {
			for c := 0; c < 4; c++ {
				if DIGIPIN_GRID[r][c] == char {
					row = r
					col = c
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			return 0, 0, fmt.Errorf("invalid character in DIGIPIN: %s", char)
		}

		latDiv := (maxLatLocal - minLatLocal) / 4
		lonDiv := (maxLonLocal - minLonLocal) / 4

		lat1 := maxLatLocal - latDiv*float64(row+1)
		lat2 := maxLatLocal - latDiv*float64(row)
		lon1 := minLonLocal + lonDiv*float64(col)
		lon2 := minLonLocal + lonDiv*float64(col+1)

		minLatLocal = lat1
		maxLatLocal = lat2
		minLonLocal = lon1
		maxLonLocal = lon2
	}

	latitude := (minLatLocal + maxLatLocal) / 2
	longitude := (minLonLocal + maxLonLocal) / 2

	return latitude, longitude, nil
}

// ParseCoordinates parses coordinate string in format "lat,lng"
func ParseCoordinates(coordStr string) (float64, float64, error) {
	parts := strings.Split(coordStr, ",")
	if len(parts) != 2 {
		return 0, 0, errors.New("coordinates must be in format lat,lng")
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude: %v", err)
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude: %v", err)
	}

	return lat, lng, nil
}
