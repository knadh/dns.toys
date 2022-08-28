package dice

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type Dice struct{}

// New returns a new instance of Dice.
func New() *Dice {
	return &Dice{}
}

var queryFormat = regexp.MustCompile("([0-9]+)[dD]([0-9]+)(?:/([0-9]+))?")

// Query returns the result of the given dice roll
func (n *Dice) Query(q string) ([]string, error) {
	// Parse the query.
	reg := queryFormat.FindStringSubmatch(q)
	if len(reg) != 4 {
		return nil, errors.New("invalid dice query.")
	}

	// Parse the matched parts as ints:
	// The elements of reg are all integers, but converting them to int can still fail if they're too large.
	dice, err := strconv.Atoi(reg[1])
	if err != nil {
		return nil, errors.New("invalid dice query.")
	}

	sides, err := strconv.Atoi(reg[2])
	if err != nil {
		return nil, errors.New("invalid dice query.")
	}

	var modifier int
	if reg[3] != "" {
		modifier, err = strconv.Atoi(reg[3])
		if err != nil {
			return nil, errors.New("invalid dice query.")
		}
	}

	sb := strings.Builder{}
	// strings.Builder is guaranteed to return nil errors, see docs.
	sb.WriteString(q)
	sb.WriteString(" 1 TXT ")

	results, total, err := performDiceRoll(dice, sides, modifier)
	if err != nil {
		return nil, errors.New("Can't generate random numbers")
	}

	sb.WriteString(`"rolled = [`)
	for i, r := range results {
		sb.WriteString(strconv.Itoa(r))
		if i != len(results)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(`]"`)

	return []string{
		sb.String(),
		fmt.Sprintf("%s 1 TXT \"total = %d\"", q, total),
	}, nil
}

// Dump is not implemented in this package.
func (n *Dice) Dump() ([]byte, error) {
	return nil, nil
}

func performDiceRoll(dice, sides, modifier int) (results []int, total int, err error) {
	results = make([]int, 0, dice)
	total = modifier

	for i := 0; i < dice; i++ {
		// Get a random number in the range 0 (inclusive) to sides (exclusive).
		max := big.NewInt(int64(sides))
		res, err := rand.Int(rand.Reader, max)
		if err != nil {
			return nil, 0, err
		}

		// max was an int, res < max, so res is convertible to an int too.
		resInt := (int)(res.Int64())

		resInt += 1 // in range 1 (inclusive) to sides (inclusive).
		results = append(results, resInt)
		total += resInt
	}

	return results, total, nil
}
