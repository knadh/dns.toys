package ifsc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	ifscCodeLen = 11
)

type branch struct {
	Bank     string `json:"BANK"`
	IFSC     string `json:"IFSC"`
	MICR     string `json:"MICR"`
	Branch   string `json:"BRANCH"`
	Address  string `json:"ADDRESS"`
	State    string `json:"STATE"`
	City     string `json:"CITY"`
	Centre   string `json:"CENTRE"`
	District string `json:"DISTRICT"`
}

type IFSC struct {
	data map[string]branch
}

func New(dir string) (*IFSC, error) {
	log.Printf("loading IFSC data from %s", dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading IFSC directory: %w", err)
	}

	// Read individual per-bank IFSC files.
	data := make(map[string]branch)
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error opening IFSC JSON file: %s: %w", path, err)
		}

		var branches map[string]branch
		if err := json.Unmarshal(b, &branches); err != nil {
			return nil, fmt.Errorf("error unmarshalling file: %s: %v", path, err)
		}

		for _, b := range branches {
			data[b.IFSC] = b
		}
	}

	return &IFSC{data: data}, nil
}

func (i *IFSC) Query(q string) ([]string, error) {
	ifscCode := strings.TrimSuffix(q, ".")
	ifscCode = strings.TrimSuffix(q, ".ifsc")
	ifscCode = strings.ToUpper(ifscCode)

	if len(ifscCode) != ifscCodeLen {
		return nil, fmt.Errorf("invalid IFSC code length: %d", len(ifscCode))
	}

	var output []string
	value, ok := i.data[ifscCode]
	if ok {
		output = []string{
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "Bank: %s"`, ifscCode, value.Bank),
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "Micr: %s"`, ifscCode, value.MICR),
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "Branch: %s"`, ifscCode, value.Branch),
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "Address: %s"`, ifscCode, value.Address),
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "City: %s"`, ifscCode, value.City),
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "Centre: %s"`, ifscCode, value.Centre),
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "District: %s"`, ifscCode, value.District),
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "State: %s"`, ifscCode, value.State),
		}
	} else {
		output = []string{fmt.Sprintf(`%s.ifsc. 1 IN TXT "IFSC code %s not found"`, ifscCode, ifscCode)}
	}

	return output, nil
}

func (i *IFSC) Dump() ([]byte, error) {
	return nil, nil
}
