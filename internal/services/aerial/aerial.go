package aerial

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
)

type Aerial struct{}

type Location struct {
	Lat  float64
	Long float64
}

const (
	delimiter = ","
	separator = "/"
	// TTL is set to 900 seconds (15 minutes).
	TTL = 900
)

var (
	validPointRegex = "(-?\\d+.\\d+)"
	latLongPair     = validPointRegex + delimiter + validPointRegex
	reParse         = regexp.MustCompile("A" + latLongPair + separator + latLongPair)
)

// New returns a new instance of Aerial.
func New() *Aerial {
	return &Aerial{}
}

// Query returns the aerial distance in KMs between lat long pair.
func (a *Aerial) Query(q string) ([]string, error) {
	parts := reParse.FindStringSubmatch(q)

	if len(parts) != 5 {
		return nil, errors.New("invalid lat long format")
	}

	var (
		res   = parts[1:]
		coord = make([]float64, 0, len(res))
	)
	for _, p := range res {
		// Iterate overy every point to convert into float.
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid point %s: %w", p, err)
		}
		coord = append(coord, f)
	}

	l1 := Location{Lat: coord[0], Long: coord[1]}
	l2 := Location{Lat: coord[2], Long: coord[3]}

	d, e := Calculate(l1, l2)
	if e != nil {
		return nil, e
	}
	result := "aerial distance = " + strconv.FormatFloat(d, 'f', 2, 64) + " KM"
	r := fmt.Sprintf(`%s %d TXT "%s"`, q, TTL, result)

	return []string{r}, nil
}

// Dump is not implemented in this package.
func (n *Aerial) Dump() ([]byte, error) {
	return nil, nil
}

// calculates aerial distance in KMs
func Calculate(l1, l2 Location) (float64, error) {
	if err := validateLoc(l1); err != nil {
		return 0, err
	}
	if err := validateLoc(l2); err != nil {
		return 0, err
	}

	var (
		radlat1  = float64(math.Pi * l1.Lat / 180)
		radlat2  = float64(math.Pi * l2.Lat / 180)
		radtheta = float64(math.Pi * float64(l1.Long-l2.Long) / 180)
	)

	d := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if d > 1 {
		d = 1
	}

	d = math.Acos(d)
	d = d * 180 / math.Pi
	d = d * 60 * 1.1515 * 1.609344

	return d, nil
}

func validateLoc(l Location) error {
	err := ""

	isLatValid := isValidPoint(l.Lat, 90)
	if !isLatValid {
		err += strconv.FormatFloat(l.Lat, 'f', -1, 64) + ": lat out of bounds"
	}

	isLongValid := isValidPoint(l.Long, 180)
	if !isLongValid {
		if !isLatValid {
			err += " "
		}
		err += strconv.FormatFloat(l.Long, 'f', -1, 64) + ": long out of bounds"
	}

	if err != "" {
		return errors.New(err)
	}

	return nil
}

func isValidPoint(point, maxVal float64) bool {
	return math.Abs(point) <= maxVal
}
