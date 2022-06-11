package base

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Base struct{}

// New returns a new instance of Base.
func New() *Base {
	return &Base{}
}

var numQueryFormat = regexp.MustCompile("([0-9A-F\\.]+)([A-Z]{3})\\-([A-Z]{3})")

var baseStrToNum = map[string]int{
	"HEX": 16,
	"DEC": 10,
	"OCT": 8,
	"BIN": 2,
}

// Query converts a number from one base to another base format
func (n *Base) Query(q string) ([]string, error) {

	q = strings.ToUpper(q)

	reg := numQueryFormat.FindStringSubmatch(q)
	if len(reg) != 4 {
		return nil, errors.New("invalid base query.")
	}

	fromBase, ok := baseStrToNum[reg[2]]
	if !ok {
		return nil, errors.New("invalid number system; must be one of hex, dec, oct, bin.")
	}

	toBase, ok := baseStrToNum[reg[3]]
	if !ok {
		return nil, errors.New("invalid number system; must be one of hex, dec, oct, bin.")
	}

	num, err := strconv.ParseInt(reg[1], fromBase, 64)
	if err != nil {
		return nil, errors.New("invalid number.")
	}

	res := strings.ToUpper(strconv.FormatInt(num, toBase))

	r := fmt.Sprintf("%s 1 TXT \"%s %s = %s %s\"", q, reg[1], reg[2], res, reg[3])
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (n *Base) Dump() ([]byte, error) {
	return nil, nil
}
