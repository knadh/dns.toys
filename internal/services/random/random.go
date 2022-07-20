package random

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
)

type Random struct{}

// New returns a new instance of Random.
func New() *Random {
	return &Random{}
}

var queryFormat = regexp.MustCompile("([0-9]+)-([0-9]+)")

// Query returns a random number in the given range
func (n *Random) Query(q string) ([]string, error) {
	// Parse the query:
	reg := queryFormat.FindStringSubmatch(q)

	if len(reg) != 3 {
		return nil, errors.New("invalid random query.")
	}

	// Parse the matched parts as ints:
	// The elements of reg are all integers, but converting them to int can still fail if they're too large.

	min, err := strconv.Atoi(reg[1])
	if err != nil {
		return nil, errors.New("invalid random query.")
	}

	max, err := strconv.Atoi(reg[2])
	if err != nil {
		return nil, errors.New("invalid random query.")
	}

	res := random(min, max)
	s := fmt.Sprintf("%s 1 TXT \"Result: %d\"", q, res)
	return []string{s}, nil
}

// Dump is not implemented in this package.
func (n *Random) Dump() ([]byte, error) {
	return nil, nil
}

func random(min, max int) int {
	return rand.Int()%(max-min) + 1 + min // inclusive
}
