// package hash computes hashes of input strings.
package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"strings"
)

// TTL is set to 1 year (hashes are deterministic).
const TTL = 31536000

// Hash computes hashes of input strings.
type Hash struct{}

// New returns a new instance of Hash.
func New() *Hash {
	return &Hash{}
}

// Query handles a hash computation query.
// Format: <algorithm>/<text>.hash (e.g., "sha256/helloworld.hash")
func (h *Hash) Query(q string) ([]string, error) {
	parts := strings.SplitN(q, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("format: <algorithm>/<text>. algorithms: md5, sha1, sha256, sha512")
	}

	algo := strings.ToLower(parts[0])
	text := parts[1]

	if len(text) > 256 {
		return nil, errors.New("input text too long (max 256 chars).")
	}

	var result string
	switch algo {
	case "md5":
		result = fmt.Sprintf("%x", md5.Sum([]byte(text)))
	case "sha1":
		result = fmt.Sprintf("%x", sha1.Sum([]byte(text)))
	case "sha256":
		result = fmt.Sprintf("%x", sha256.Sum256([]byte(text)))
	case "sha512":
		result = fmt.Sprintf("%x", sha512.Sum512([]byte(text)))
	default:
		return nil, fmt.Errorf("unknown algorithm '%s'. supported: md5, sha1, sha256, sha512", algo)
	}

	r := fmt.Sprintf(`%s %d TXT "%s" "%s(%s)"`, q, TTL, result, algo, text)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (h *Hash) Dump() ([]byte, error) {
	return nil, nil
}
