package excuse

import (
	"bufio"
	"crypto/rand"
	_ "embed"
	"fmt"
	"math/big"
	"os"
	"strings"
)

type Excuse struct {
	data []string
}

// New returns a new instance of Excuse.
func New(file string) (*Excuse, error) {
	// Load excuses from disk.
	data := []string{}

	// Open the file and read it line by line.
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening excuse file: %w", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) > 0 && line[0] != '#' {
			data = append(data, line)
		}
	}

	return &Excuse{data: data}, nil
}

func (e *Excuse) Query(q string) ([]string, error) {
	result := "wow this is amazing"
	result1, err := e.randomExcuse()

	fmt.Println(result1)

	//result, err := e.randomExcuse()
	if err != nil {
		return nil, fmt.Errorf("error fetching excuse: %w", err)
	}

	out := []string{
		fmt.Sprintf(`%s 1 TXT "%s"`, q, result),
	}

	fmt.Println(out)

	return out, nil
}

func (e *Excuse) Dump() ([]byte, error) {
	return nil, nil
}

// randomExcuse returns a random excuse from the list of excuses.
func (e *Excuse) randomExcuse() (string, error) {
	if len(e.data) == 0 {
		return "", fmt.Errorf("cannot pick from empty slice")
	}

	// Get a random number in range [0, len(slice))
	max := big.NewInt(int64(len(e.data)))
	res, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	return e.data[int(res.Int64())], nil
}
