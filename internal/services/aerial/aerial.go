package aerial

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
)

type Aerial struct{}

// New returns a new instance of Aerial.
func New() *Aerial {
	return &Aerial{}
}

var validPointRegex = "(-?\\d+.\\d+)"
var delimiter = "-"

var reParse = regexp.MustCompile(validPointRegex + delimiter + validPointRegex + delimiter + validPointRegex + delimiter + validPointRegex)

// TODO: remove debug comments and decide limiter ","

// Query returns the aerial distance in KMs between lat lng pair
func (a *Aerial) Query(q string) ([]string, error) {
	fmt.Println(q)
	regexGroups := reParse.FindStringSubmatch(q)
	fmt.Println("regex groups"); // remove comment 
	for _, rg := range regexGroups {
		fmt.Println(rg); // remove comment
	}

	if len(regexGroups) != 5 {
		return nil, errors.New("invalid lat lng format")
	}

	res := regexGroups[1:]
	cord := make([]float64, 0, len(res))
	fmt.Println("after parsing regex groups"); // remove comment
	for _, p := range res {
		// iterate overy every point to convert into float
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid point %s; Error: %w", p, err)
		}
		fmt.Println(f); // remove comment
		cord = append(cord, f)
	}

	d, err := calculateAerialDistance(cord[0], cord[1], cord[2], cord[3]);
	if err != nil {
		return nil, fmt.Errorf("error in aerial distance calculation: %w", err)
	}

	result := "Aerial Distance is " + d + " KMs"

	r := fmt.Sprintf(`%s 1 TXT "%s"`, q, result)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (n *Aerial) Dump() ([]byte, error) {
	return nil, nil
}

// calculates aerial distance in KMs
func calculateAerialDistance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) (string, error) {
	fmt.Println("in fn", lat1, lng1, lat2, lng2) // remove comment

	validateLat(lat1)
	validateLng(lng1)
	validateLat(lat2)
	validateLng(lng2)
	
	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)
	
	radtheta := float64(math.Pi * float64(lng1 - lng2) / 180)
	
	d := math.Sin(radlat1) * math.Sin(radlat2) + math.Cos(radlat1) * math.Cos(radlat2) * math.Cos(radtheta);
	if d > 1 {
		d = 1
	}
	
	d = math.Acos(d)
	d = d * 180 / math.Pi
	d = d * 60 * 1.1515 * 1.609344

	s := strconv.FormatFloat(d, 'f', 2, 64)
	
	return s, nil
}

func validatePoint(point, maxVal float64) (error) {
	absoluteVal := math.Abs(point);
	if (absoluteVal > maxVal) {
		return errors.New("point out of bounds")
	}
	return nil
}

func validateLat(lat float64) (error) {
	return validatePoint(lat, 90);
}

func validateLng(lng float64) (error) {
	return validatePoint(lng, 180);
}