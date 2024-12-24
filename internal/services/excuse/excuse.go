package excuse

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"math/big"
)

// excuse represents a single excuse.
type excuse struct {
	ID     int    `yaml:"id"`
	TextEn string `yaml:"text_en"`
}

type Excuse struct {
	excuses []excuse
}

//go:embed excuses.yml
var dataB []byte

// New returns a new instance of Excuse.
func New() (*Excuse, error) {
	e := &Excuse{
		excuses: make([]excuse, 0),
	}

	if err := e.load(); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Excuse) Query(q string) ([]string, error) {
	result, err := e.randomExcuse()
	if err != nil {
		return nil, fmt.Errorf("error fetching excuse: %w", err)
	}

	out := []string{
		fmt.Sprintf(`%s 1 TXT "%s"`, q, result),
	}

	return out, nil
}

func (e *Excuse) Dump() ([]byte, error) {
	return nil, nil
}

func (e *Excuse) load() error {
	if err := yaml.Unmarshal(dataB, &e.excuses); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// randomExcuse returns a random excuse from the list of excuses.
func (e *Excuse) randomExcuse() (string, error) {
	if len(e.excuses) == 0 {
		return "", fmt.Errorf("cannot pick from empty slice")
	}

	// Get a random number in range [0, len(slice))
	max := big.NewInt(int64(len(e.excuses)))
	res, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	return e.excuses[int(res.Int64())].TextEn, nil
}
