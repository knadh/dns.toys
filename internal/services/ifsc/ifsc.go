package ifsc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	ifscCodeLen = 11
)

type Branch struct {
	Bank string `json:"BANK"`
	IFSC string `json:"IFSC"`
	Branch string `json:"BRANCH"`
	City string `json:"CITY"`
	State string `json:"STATE"`
}

type BranchDetails struct {
	Bank string
	Branch string
	City string
	State string
}

type IFSC struct {
	data map[string]BranchDetails
}

func New(dir string) (*IFSC, error) {
	// Load IFSCs from disk
	log.Printf("loading IFSC data from %s", dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error accessing IFSC directory: %w", err)
	}

	var ifsc IFSC
	ifsc.data = make(map[string]BranchDetails)
	
	for _, file := range files {
		var jsonFilePath = filepath.Join(dir, file.Name())
		jsonFile, err := os.Open(jsonFilePath)
		if err != nil {
			return nil, fmt.Errorf("error opening IFSC JSON file: %w", err)
		}
		defer jsonFile.Close()
	
		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading IFSC JSON file: %w", err)
		}
	
		var branches map[string]Branch	
		json.Unmarshal(byteValue, &branches)

		for _, v := range branches {
			ifsc.data[v.IFSC] = BranchDetails{Bank: v.Bank, Branch: v.Branch, City: v.City, State: v.State}
		}
	}
	log.Printf("loaded IFSC data")

	return &ifsc, nil
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
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "BANK: %s"`, ifscCode, value.Bank), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "BRANCH: %s"`, ifscCode, value.Branch), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "CITY: %s"`, ifscCode, value.City), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "STATE: %s"`, ifscCode, value.State), 
		}
	} else {
		output = []string{fmt.Sprintf(`%s.ifsc. 1 IN TXT "IFSC code %s not found"`, ifscCode, ifscCode)}
	}

	return output, nil
}

func (i *IFSC) Dump() ([]byte, error) {
	return nil, nil
}
