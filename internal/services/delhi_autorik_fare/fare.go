package fare

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	baseFare  = 31   // Initial fare for the first km
	perKmFare = 15   // Fare per km after the first km
	TTL       = 900  // Time-to-live for DNS TXT records
)

var reParseFare = regexp.MustCompile(`(?i)([0-9\.]+)km-fare`)

// Fare calculates the fare for a given distance in kilometers.
type Fare struct {
	help []string
}

// New returns a new instance of Fare.
func New() (*Fare, error) {
	f := &Fare{}
	f.help = []string{
		fmt.Sprintf("Calculate fare: usage example - 42km-fare (for 42 kilometers). Base fare is ₹%d, ₹%d per km after the first.", baseFare, perKmFare),
	}
	return f, nil
}

// Query parses a fare calculation query and returns the results.
func (f *Fare) Query(q string) ([]string, error) {
	if q == "fare." {
		return f.help, nil
	}

	res := reParseFare.FindStringSubmatch(q)
	if len(res) != 2 {
		return nil, errors.New("invalid fare query. Use the format: 'Xkm-fare' where X is the number of kilometers.")
	}

	// Parse the distance in km
	kms, err := strconv.ParseFloat(res[1], 32)
	if err != nil {
		return nil, errors.New("invalid distance.")
	}

	// Calculate the fare.
	fare := baseFare + (kms-1)*perKmFare
	if kms <= 1 {
		fare = baseFare
	}

	r := fmt.Sprintf("%s %d TXT \"%.2f km = ₹%.2f\"", q, TTL, kms, fare)

	return []string{r}, nil
}
