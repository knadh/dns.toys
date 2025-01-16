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

// returns Holiday instance with file path
// idk how else to pass the file path from config to load json
// might improve on this
func New(file string) (*Holiday, error) {
	return &Holiday{fileP: file}, nil
}

// loads data from disk and stores data in Holiday instance
func (h *Holiday) loadJson(countryCode string) (map[string][]string, error) {

	_, m, _ := time.Now().Date()
	var data map[string]map[string][]string

	holidayJson, err := os.ReadFile(fmt.Sprintf(h.fileP, countryCode))
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	json.Unmarshal(holidayJson, &data)

	h.data = data[strings.ToLower(m.String())]

	return data[strings.ToLower(m.String())], err
}

func (h *Holiday) Query(q string) ([]string, error) {
	// _, m, _ := time.Now().Date()
	splitQuery := strings.Split(q, ".")
	state := splitQuery[0]
	var countryCode string

	fmt.Println(splitQuery)

	if len(splitQuery) == 4 {
		countryCode = splitQuery[2]
	} else {
		countryCode = "india"
	}

	var results map[string][]string
	var err error

	// match current month with last stored month
	// fetch and parse json again only if month not matching
	// not valid after making json universal

	// if len(h.data) != 0 && h.data["month"][0] == strings.ToLower(m.String()) {
	// 	results = h.data
	// } else {
	// 	results, err = h.loadJson(countryCode)
	// }

	results, err = h.loadJson(countryCode)

	if err != nil {
		return nil, err
	}

	resultArr, exists := results[state]

	out := make([]string, 0, len(resultArr))

	// in case of mispell
	if !exists {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, "Maybe you mispelled the state/country?"))
		return out, nil
	}

	// in case no holiday that month in that state
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
