package nanoid

import (
	"fmt"
	"strconv"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type NanoID struct {
	maxResults int
	maxLength  int
}

func New(maxResults, maxLength int) *NanoID {
	if maxResults < 1 {
		maxResults = 1
	}
	if maxLength < 1 {
		maxLength = 1
	}
	return &NanoID{
		maxResults: maxResults,
		maxLength:  maxLength,
	}
}

// Query returns a random NanoID.
func (n *NanoID) Query(q string) ([]string, error) {
	parts := strings.Split(q, ".")
	num := 1
	length := 21 // default length for NanoID

	if len(parts) > 1 {
		num, _ = strconv.Atoi(parts[0])
		length, _ = strconv.Atoi(parts[1])
	}

	if num < 1 || num > n.maxResults {
		return nil, fmt.Errorf("provide 1-%d.nanoid", n.maxResults)
	}
	if length < 1 || length > n.maxLength {
		return nil, fmt.Errorf("provide length 1-%d.nanoid", n.maxLength)
	}

	out := make([]string, 0, num)
	for i := 0; i < num; i++ {
		id, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", length)
		if err != nil {
			return nil, err
		}
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, id))
	}

	return out, nil
}

// Dump is not implemented in this package.
func (n *NanoID) Dump() ([]byte, error) {
	return nil, nil
}
