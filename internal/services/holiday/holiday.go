package holiday

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// map model for ref
// {
// 	 countryCode: {
// 			month: ["someMonth"]
// 			national: []
// 			...
// 	 	}
// }

type Holiday struct {
	fileP string                         // points to holiday.json
	data  map[string]map[string][]string //months state wise holiday
}

// returns Holiday instance with file path
// idk how else to pass the file path from config to load json
// might improve on this
func New(file string) (*Holiday, error) {
	return &Holiday{fileP: file, data: map[string]map[string][]string{}}, nil
}

// loads data from disk and stores data in Holiday instance
func (h *Holiday) loadJson(countryCode string) (map[string][]string, error) {

	// match current month with last stored month
	// fetch and parse json again only if month not matching
	_, m, _ := time.Now().Date()
	currMonth := strings.ToLower(m.String())

	if len(h.data[countryCode]) != 0 && h.data[countryCode]["month"][0] == currMonth {
		fmt.Println("cache hit!")
		return h.data[countryCode], nil
	}

	fmt.Println("cache missed!")

	var data map[string]map[string][]string

	holidayJson, err := os.ReadFile(fmt.Sprintf(h.fileP, countryCode))
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	json.Unmarshal(holidayJson, &data)

	h.data[countryCode] = data[currMonth]

	return data[currMonth], err
}

func (h *Holiday) Query(q string) ([]string, error) {
	splitQuery := strings.Split(q, ".")
	var state, countryCode string

	if len(splitQuery) > 1 {
		state = splitQuery[0]
		countryCode = splitQuery[1]
	} else {
		countryCode = splitQuery[0]
	}

	results, err := h.loadJson(countryCode)

	if r := "Country Support To Be Added Soon!"; err != nil {
		log.Printf("error preparing response: %v", err)
		return []string{fmt.Sprintf(`%s 1 TXT "%s"`, q, r)}, nil
	}

	var resultsArr []string
	exists := true

	out := make([]string, 0, len(resultsArr))

	resultsArr = results["national"]
	var stateRes []string
	if state != "" {
		stateRes, exists = results[state]
		resultsArr = append(resultsArr, stateRes...)
	}

	sort.Strings(resultsArr)

	// in case of mispell
	if !exists {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, "Maybe you mispelled the state/country?"))
		return out, nil
	}

	// in case no holiday that month in that state
	if len(resultsArr) == 0 {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, "No Holidays this month :("))
		return out, nil
	}

	for _, r := range resultsArr {
		out = append(out, fmt.Sprintf(`%s 1 TXT "%s"`, q, r))
	}

	return out, nil
}

func (h *Holiday) Dump() ([]byte, error) {
	return nil, nil
}
