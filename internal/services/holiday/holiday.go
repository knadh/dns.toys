package holiday

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Holiday struct {
	data []string
}

func New(file string) (*Holiday, error) {
	//TODO

	LoadJson(file)

	return nil, nil
}

// fetch holidays of current month from disk
func (h *Holiday) fetchHolidays(s string) ([]string, error) {
	//TODO
	return nil, nil
}

type YearlyHolidays struct {
	Data map[string][]string
}

func LoadJson(file string) (interface{}, error) {

	_, m, _ := time.Now().Date()
	holidayJson, err := os.ReadFile(file)

	var data map[string]map[string][]string

	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	json.Unmarshal(holidayJson, &data)

	fmt.Println(data[strings.ToLower(m.String())]["maharashtra"], "line 42: hoiday.go", m)

	return data, nil

}

// func (h *Holiday) checkForUpdate() {
// 	currMonth := time.Now()
// }

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
