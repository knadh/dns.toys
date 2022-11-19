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
	Lat float64
	Lng float64
}

// New returns a new instance of Aerial.
func New() *Aerial {
	return &Aerial{}
}

var validPointRegex = "(-?\\d+.\\d+)"
var delimiter = ","
var separator = "/"
var latlngpair = validPointRegex + delimiter + validPointRegex

var reParse = regexp.MustCompile("A" + latlngpair + separator + latlngpair)

// Query returns the aerial distance in KMs between lat lng pair
func (a *Aerial) Query(q string) ([]string, error) {
	regexGroups := reParse.FindStringSubmatch(q)

	if len(regexGroups) != 5 {
		return nil, errors.New("invalid lat lng format")
	}

	res := regexGroups[1:]
	cord := make([]float64, 0, len(res))
	for _, p := range res {
		// iterate overy every point to convert into float
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid point %s; Error: %w", p, err)
		}
		cord = append(cord, f)
	}

	l1 := Location{Lat: cord[0], Lng: cord[1]}
	l2 := Location{Lat: cord[2], Lng: cord[3]}

	d, e := CalculateAerialDistance(l1, l2)
	if e != nil {
		return nil, e
	}

	result := "aerial distance = " + strconv.FormatFloat(d, 'f', 2, 64) + " KMs"

	r := fmt.Sprintf(`%s 1 TXT "%s"`, q, result)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (n *Aerial) Dump() ([]byte, error) {
	return nil, nil
}

// calculates aerial distance in KMs
func CalculateAerialDistance(l1, l2 Location) (float64, error) {
	
	e1 := validateLocation(l1)
	e2 := validateLocation(l2)

	if e1 != nil || e2 != nil {
		errString := "";
		if (e1 != nil) {
			errString += e1.Error();
		}
		if (e2 != nil) {
			if (e1 != nil) {
				errString += "; "
			}
			errString += e2.Error()
		}

		return 0, errors.New(errString)
	}

	lat1 := l1.Lat
	lng1 := l1.Lng
	lat2 := l2.Lat
	lng2 := l2.Lng

	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	radtheta := float64(math.Pi * float64(lng1-lng2) / 180)

	d := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if d > 1 {
		d = 1
	}

	d = math.Acos(d)
	d = d * 180 / math.Pi
	d = d * 60 * 1.1515 * 1.609344

	return d, nil
}

func isValidPoint(point, maxVal float64) bool {
	absoluteVal := math.Abs(point)
	return absoluteVal <= maxVal
}

func validateLocation(l Location) error {
	errString := ""

	isLatValid := isValidPoint(l.Lat, 90)
	if !isLatValid {
		errString += strconv.FormatFloat(l.Lat, 'f', -1, 64) + " lat out of bounds"
	}

	isLngValid := isValidPoint(l.Lng, 180)
	if !isLngValid {
		if (!isLatValid) {
			errString += " "
		}
		errString += strconv.FormatFloat(l.Lng, 'f', -1, 64) + " lng out of bounds"
	}

	if (errString != "") {
		return errors.New(errString)
	}

	return nil
}