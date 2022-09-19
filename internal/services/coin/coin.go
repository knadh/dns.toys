package coin

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type CoinTossResult string

var (
	Heads CoinTossResult = "heads"
	Tails CoinTossResult = "tails"
)

type Coin struct{}

// New returns a new instance of Coin.
func New() *Coin {
	return &Coin{}
}

var maxTosses = 42

// Query returns the result of the given coin toss
func (n *Coin) Query(q string) ([]string, error) {
	var tosses int
	var err error

	if q == "coin." {
		// dig coin @dns.toys
		tosses = 1
	} else {
		tosses, err = strconv.Atoi(q)
		if err != nil {
			return nil, errors.New("invalid coin toss query")
		}
	}

	if tosses > maxTosses {
		return nil, errors.New("toss overflow")
	}

	sb := strings.Builder{}
	// strings.Builder is guaranteed to return nil errors, see docs.
	sb.WriteString(q)
	sb.WriteString(" 1 TXT ")

	results, err := performCoinToss(tosses)
	if err != nil {
		return nil, fmt.Errorf("error occurred whiler performing the coin toss: %w", err)
	}

	sb.WriteString(`"tossed = [`)
	for i, r := range results {
		sb.WriteString(string(r))
		if i != len(results)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(`]"`)

	return []string{
		sb.String(),
	}, nil
}

// Dump is not implemented in this package.
func (n *Coin) Dump() ([]byte, error) {
	return nil, nil
}

func performCoinToss(tosses int) (results []CoinTossResult, err error) {
	results = make([]CoinTossResult, 0, tosses)
	for i := 0; i < tosses; i++ {
		res := Heads
		if rand.Int()%2 == 0 {
			res = Tails
		}
		results = append(results, res)
	}

	return results, nil
}
