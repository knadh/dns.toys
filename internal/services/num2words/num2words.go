// Package num2words implements numbers to words converter.
package num2words

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ones = []string{
		"zero", "one", "two", "three", "four",
		"five", "six", "seven", "eight", "nine",
		"ten", "eleven", "twelve", "thirteen", "fourteen",
		"fifteen", "sixteen", "seventeen", "eighteen", "nineteen",
	}
	tens = []string{
		"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety",
	}
	big = []string{
		"", "thousand", "million", "billion", "trillion", "quadrillion", "quintillion",
	}
)

type Num2Words struct{}

// New returns a new instance of Num2Words.
func New() *Num2Words {
	return &Num2Words{}
}

// Query converts a number to words.
func (n *Num2Words) Query(q string) ([]string, error) {
	num, err := strconv.ParseFloat(q, 64)
	if err != nil {
		return nil, errors.New("invalid number.")
	}

	w := num2words(int(num))
	decimalIndex := strings.IndexRune(q, '.')
	if decimalIndex >= 0 {
		decimalValue := q[decimalIndex+1:decimalIndex+2] + strings.TrimRight(q[decimalIndex+2:], "0")
		w += " Point" + decimal2words(decimalValue)
	}

	// TTL is set to 900 seconds (15 minutes).
	r := fmt.Sprintf("%s 900 TXT \"%g = %s\"", q, num, w)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (n *Num2Words) Dump() ([]byte, error) {
	return nil, nil
}

func num2words(number int) string {
	out := ""

	if number == 0 {
		return ones[0]
	} else if number < 0 {
		out += "minus"
		number = number * -1
	}

	// Divide the number into groups of 3.
	// eg: 9876543210 = [210 543 876 9] (should be read in reverse)
	groups := []int{}
	for {
		if number < 1 {
			break
		}

		groups = append(groups, number%1000)
		number /= 1000
	}

	ln := len(groups) - 1
	for i := ln; i >= 0; i-- {
		n := groups[i]
		if n == 0 {
			continue
		}

		num := n
		if v := num / 100; v != 0 {
			out += " " + ones[v] + " hundred"
			num = num % 100
		}

		if v := num / 10; num >= 20 && v != 0 {
			out += " " + tens[v]
			num = num % 10
		}

		if num > 0 {
			out += " " + ones[num]
		}

		if i > 0 && i <= len(big) {
			out += " " + big[i] + ","
		}

	}

	return out[1:]
}

func decimal2words(decimal string) string {
	out := ""
	for _, c := range decimal {
		out += " " + ones[c-'0']
	}
	return out
}
