package coin

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
)

const (
	heads     = "heads"
	tails     = "tails"
	maxTosses = 42
)

type Coin struct{}

// New returns a new instance of Coin.
func New() *Coin {
	return &Coin{}
}

// Query returns the result of the given coin toss
func (n *Coin) Query(q string) ([]string, error) {
	tosses := 1
	if q != "coin." {
		t, err := strconv.Atoi(q)
		if err != nil {
			return nil, errors.New("invalid coin toss query")
		}
		tosses = t
	}

	if tosses > maxTosses {
		return nil, fmt.Errorf("max allowed tosses is %d", maxTosses)
	}

	results, err := performCoinToss(tosses)
	if err != nil {
		return nil, fmt.Errorf("error occurred whiler performing the coin toss: %w", err)
	}

	out := make([]string, 0, len(results))
	for _, r := range results {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, r))
	}

	return out, nil
}

// Dump is not implemented in this package.
func (n *Coin) Dump() ([]byte, error) {
	return nil, nil
}

func performCoinToss(tosses int) ([]string, error) {
	out := make([]string, 0, tosses)

	for i := 0; i < tosses; i++ {
		res := heads
		if rand.Int()%2 == 0 {
			res = tails
		}
		out = append(out, res)
	}

	return out, nil
}
