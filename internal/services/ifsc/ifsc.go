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
	Bank 		string 	`json:"BANK"`
	IFSC 		string 	`json:"IFSC"`
	MICR 		string 	`json:"MICR"`
	Branch 		string 	`json:"BRANCH"`
	Address 	string 	`json:"ADDRESS"`
	State 		string 	`json:"STATE"`
	Contact 	string 	`json:"CONTACT"`
	UPI 		bool 	`json:"UPI"`
	RTGS 		bool 	`json:"RTGS"`
	City 		string 	`json:"CITY"`
	Centre 		string 	`json:"CENTRE"`
	District 	string 	`json:"DISTRICT"`
	NEFT 		bool 	`json:"NEFT"`
	IMPS 		bool 	`json:"IMPS"`
	SWIFT 		string 	`json:"SWIFT"`
	ISO3166 	string 	`json:"ISO3166"`
}

type BranchDetails struct {
	Bank 		string 
	MICR 		string 
	Branch 		string 
	Address 	string 
	State 		string 
	Contact 	string 
	UPI 		bool 
	RTGS 		bool 
	City 		string 
	Centre 		string 
	District 	string 
	NEFT 		bool 
	IMPS 		bool 
	SWIFT 		string 
	ISO3166 	string 
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
			ifsc.data[v.IFSC] = BranchDetails{
									Bank: v.Bank, 
									MICR: v.MICR, 
									Branch: v.Branch, 
									Address: v.Address, 
									State: v.State, 
									Contact: v.Contact, 
									UPI: v.UPI, 
									RTGS: v.RTGS, 
									City: v.City, 
									Centre: v.Centre, 
									District: v.District, 
									NEFT: v.NEFT, 
									IMPS: v.IMPS, 
									SWIFT: v.SWIFT, 
									ISO3166: v.ISO3166, 
								}
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
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "MICR: %s"`, ifscCode, value.MICR), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "BRANCH: %s"`, ifscCode, value.Branch), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "ADDRESS: %s"`, ifscCode, value.Address), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "STATE: %s"`, ifscCode, value.State), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "CONTACT: %s"`, ifscCode, value.Contact), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "UPI: %t"`, ifscCode, value.UPI), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "RTGS: %t"`, ifscCode, value.RTGS), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "CITY: %s"`, ifscCode, value.City), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "CENTRE: %s"`, ifscCode, value.Centre), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "DISTRICT: %s"`, ifscCode, value.District), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "NEFT: %t"`, ifscCode, value.NEFT), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "IMPS: %t"`, ifscCode, value.IMPS), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "SWIFT: %s"`, ifscCode, value.SWIFT), 
			fmt.Sprintf(`%s.ifsc. 1 IN TXT "ISO3166: %s"`, ifscCode, value.ISO3166), 
		}
	} else {
		output = []string{fmt.Sprintf(`%s.ifsc. 1 IN TXT "IFSC code %s not found"`, ifscCode, ifscCode)}
	}

	return output, nil
}

func (i *IFSC) Dump() ([]byte, error) {
	return nil, nil
}
