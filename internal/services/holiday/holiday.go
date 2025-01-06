package holiday

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Holiday struct {
	fileP string              // points to holiday.json
	data  map[string][]string //months state wise holiday
}

func New(file string) (*Holiday, error) {
	return &Holiday{fileP: file}, nil
}

func (h *Holiday) loadJson() (map[string][]string, error) {

	_, m, _ := time.Now().Date()
	var data map[string]map[string][]string

	holidayJson, err := os.ReadFile(h.fileP)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	json.Unmarshal(holidayJson, &data)

	h.data = data[strings.ToLower(m.String())]

	return data[strings.ToLower(m.String())], err
}

func (h *Holiday) Query(q string) ([]string, error) {
	_, m, _ := time.Now().Date()
	state := strings.Split(q, ".")[0]

	var results map[string][]string
	var err error

	if len(h.data) != 0 && h.data["month"][0] == strings.ToLower(m.String()) {
		fmt.Println("from snapshot waooo")
		results = h.data
	} else {
		results, err = h.loadJson()
	}

	if err != nil {
		return nil, err
	}

	resultArr, exists := results[state]

	out := make([]string, 0, len(resultArr))

	if !exists {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, "Maybe you mispelled the state?"))
		return out, nil
	}

	if len(results[state]) == 0 {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, "No Holidays this month :("))
		return out, nil
	}

	for _, r := range results[state] {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, r))
	}

	return out, nil
}

func (h *Holiday) Dump() ([]byte, error) {
	return nil, nil
}
