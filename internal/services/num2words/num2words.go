// Package num2words implements numbers to words converter.
// Forked from https://github.com/divan/num2words (Copyright (c) 2013 Ivan Daniluk, MIT License)
package num2words

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

const groupsNumber int = 4

var (
	_smallNumbers = []string{
		"zero", "one", "two", "three", "four",
		"five", "six", "seven", "eight", "nine",
		"ten", "eleven", "twelve", "thirteen", "fourteen",
		"fifteen", "sixteen", "seventeen", "eighteen", "nineteen",
	}
	_tens = []string{
		"", "", "twenty", "thirty", "forty", "fifty",
		"sixty", "seventy", "eighty", "ninety",
	}
	_scaleNumbers = []string{
		"", "thousand", "million", "billion",
	}
)

type Num2Words struct{}

// New returns a new instance of Num2Words.
func New() *Num2Words {
	return &Num2Words{}
}

// Query converts a number to words.
func (n *Num2Words) Query(q string) ([]string, error) {
	num, err := strconv.Atoi(q)
	if err != nil {
		return nil, errors.New("invalid number.")
	}

	w := convert(num, false)
	r := fmt.Sprintf("%s 1 TXT \"%d = %s\"", q, num, w)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (n *Num2Words) Dump() ([]byte, error) {
	return nil, nil
}

func convert(number int, useAnd bool) string {
	// Zero rule
	if number == 0 {
		return _smallNumbers[0]
	}

	// Divide into three-digits group
	var groups [groupsNumber]int
	positive := math.Abs(float64(number))

	// Form three-digit groups
	for i := 0; i < groupsNumber; i++ {
		groups[i] = int(math.Mod(positive, 1000))
		positive /= 1000
	}

	var textGroup [groupsNumber]string
	for i := 0; i < groupsNumber; i++ {
		textGroup[i] = digitGroup2Text(groups[i], useAnd)
	}
	combined := textGroup[0]
	and := useAnd && (groups[0] > 0 && groups[0] < 100)

	for i := 1; i < groupsNumber; i++ {
		if groups[i] != 0 {
			prefix := textGroup[i] + " " + _scaleNumbers[i]

			if len(combined) != 0 {
				prefix += separator(and)
			}

			and = false

			combined = prefix + combined
		}
	}

	if number < 0 {
		combined = "minus " + combined
	}

	return combined
}

func intMod(x, y int) int {
	return int(math.Mod(float64(x), float64(y)))
}

func digitGroup2Text(group int, useAnd bool) (ret string) {
	hundreds := group / 100
	tensUnits := intMod(int(group), 100)

	if hundreds != 0 {
		ret += _smallNumbers[hundreds] + " hundred"

		if tensUnits != 0 {
			ret += separator(useAnd)
		}
	}

	tens := tensUnits / 10
	units := intMod(tensUnits, 10)

	if tens >= 2 {
		ret += _tens[tens]

		if units != 0 {
			ret += "-" + _smallNumbers[units]
		}
	} else if tensUnits != 0 {
		ret += _smallNumbers[tensUnits]
	}

	return
}

// separator returns proper separator string between
// number groups.
func separator(useAnd bool) string {
	if useAnd {
		return " and "
	}
	return " "
}
