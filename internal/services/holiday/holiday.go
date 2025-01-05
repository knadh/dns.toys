package holiday

import (
	"fmt"
	"strings"
)

type Holiday struct {
	data []int
}

func New() (*Holiday, error) {
	//TODO
	return nil, nil
}

func (h *Holiday) fetchHolidays() ([]string, error) {
	//TODO
	return nil, nil
}

func (h *Holiday) Query(q string) ([]string, error) {
	results := make(map[string][]string)
	state := strings.Split(q, ".")[0]

	results["maharashtra"] = []string{"1st January", "26th January", "30th January"}
	results["delhi"] = []string{"2nd January", "10th January", "15th January"}

	resultStr, exists := results[state]

	if !exists {
		return nil, fmt.Errorf("misspelled state or not available")
	}

	out := make([]string, 0, len(results))
	for _, r := range resultStr {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, r))
	}

	return out, nil
}

func (h *Holiday) Dump() ([]byte, error) {
	return nil, nil
}
