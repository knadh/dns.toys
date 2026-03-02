// package pass generates cryptographically secure random passwords.
package pass

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	// Character sets for password generation.
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits    = "0123456789"
	symbols   = "!@#$%^&*()-_=+[]{}|;:,.<>?"

	charset = lowercase + uppercase + digits + symbols

	// TTL is set to 1 second (passwords should not be cached).
	TTL = 1

	// Limits.
	minLength = 8
	maxLength = 64
)

// Pass generates random passwords.
type Pass struct{}

// New returns a new instance of Pass.
func New() *Pass {
	return &Pass{}
}

// Query generates a random password of the requested length.
// Format: <length>.pass (e.g., "16.pass" generates a 16-char password)
func (p *Pass) Query(q string) ([]string, error) {
	q = strings.TrimSuffix(q, ".")

	length, err := strconv.Atoi(q)
	if err != nil || q == "" {
		// Default to 16 if no length specified.
		length = 16
	}

	if length < minLength {
		return nil, fmt.Errorf("minimum password length is %d.", minLength)
	}
	if length > maxLength {
		return nil, fmt.Errorf("maximum password length is %d.", maxLength)
	}

	pw, err := generate(length)
	if err != nil {
		return nil, errors.New("error generating password.")
	}

	r := fmt.Sprintf(`%s %d TXT "%s"`, q, TTL, pw)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (p *Pass) Dump() ([]byte, error) {
	return nil, nil
}

// generate produces a cryptographically secure random password.
func generate(length int) (string, error) {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = charset[idx.Int64()]
	}

	// Ensure at least one char from each category for strong passwords.
	if length >= 4 {
		categories := []string{lowercase, uppercase, digits, symbols}
		for i, cat := range categories {
			catLen := big.NewInt(int64(len(cat)))
			idx, err := rand.Int(rand.Reader, catLen)
			if err != nil {
				return "", err
			}
			// Place required chars at positions spread across the password.
			pos := i * (length / 4)
			result[pos] = cat[idx.Int64()]
		}
	}

	return string(result), nil
}
