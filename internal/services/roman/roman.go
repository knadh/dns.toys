// package roman converts between Arabic and Roman numerals.
package roman

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// TTL is set to 1 year (conversions are deterministic).
const TTL = 31536000

var (
	// Roman numeral values, ordered from largest to smallest.
	romanValues = []struct {
		value  int
		symbol string
	}{
		{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
		{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
	}

	// Map for Roman to Arabic conversion.
	romanMap = map[byte]int{
		'I': 1, 'V': 5, 'X': 10, 'L': 50,
		'C': 100, 'D': 500, 'M': 1000,
	}
)

// Roman converts between Arabic and Roman numerals.
type Roman struct{}

// New returns a new instance of Roman.
func New() *Roman {
	return &Roman{}
}

// Query handles a roman numeral conversion query.
func (r *Roman) Query(q string) ([]string, error) {
	q = strings.TrimSuffix(q, ".")

	// Try to parse as an integer first (Arabic → Roman).
	if num, err := strconv.Atoi(q); err == nil {
		if num <= 0 || num > 3999 {
			return nil, errors.New("number must be between 1 and 3999.")
		}
		roman := toRoman(num)
		out := fmt.Sprintf(`%s %d TXT "%d = %s"`, q, TTL, num, roman)
		return []string{out}, nil
	}

	// Try to parse as Roman numerals (Roman → Arabic).
	upper := strings.ToUpper(q)
	num, err := fromRoman(upper)
	if err != nil {
		return nil, err
	}

	out := fmt.Sprintf(`%s %d TXT "%s = %d"`, q, TTL, upper, num)
	return []string{out}, nil
}

// Dump is not implemented in this package.
func (r *Roman) Dump() ([]byte, error) {
	return nil, nil
}

// toRoman converts an integer to a Roman numeral string.
func toRoman(num int) string {
	var result strings.Builder
	for _, rv := range romanValues {
		for num >= rv.value {
			result.WriteString(rv.symbol)
			num -= rv.value
		}
	}
	return result.String()
}

// fromRoman converts a Roman numeral string to an integer.
func fromRoman(s string) (int, error) {
	if len(s) == 0 {
		return 0, errors.New("empty roman numeral.")
	}

	total := 0
	for i := 0; i < len(s); i++ {
		val, ok := romanMap[s[i]]
		if !ok {
			return 0, fmt.Errorf("invalid roman numeral character: %c", s[i])
		}

		// If the next value is larger, subtract current value (e.g., IV = 4).
		if i+1 < len(s) {
			next, ok := romanMap[s[i+1]]
			if !ok {
				return 0, fmt.Errorf("invalid roman numeral character: %c", s[i+1])
			}
			if next > val {
				total += next - val
				i++ // Skip next character.
				continue
			}
		}

		total += val
	}

	// Validate by converting back.
	if toRoman(total) != s {
		return 0, errors.New("invalid roman numeral sequence.")
	}

	if total > 3999 {
		return 0, errors.New("roman numeral too large (max 3999).")
	}

	return total, nil
}
