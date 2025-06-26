package vitamin

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type vitamin struct {
	CommonName string `json:"common_name"`
	ScientificName string `json:"scientific_name"`
	Sources []string `json:"sources"`
}

type VitaminStore struct {
	data map[string]vitamin
}

func New(filePath string) (*VitaminStore, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading the JSON file: %s: %w", filePath, err)
	}

	data := make(map[string]vitamin)
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling the %s data: %w", filePath, err)
	}

	return &VitaminStore{data: data}, nil
}

func (v *VitaminStore) Query(query string) ([]string, error) {
	var output []string

	if vitaminData, ok := v.data[strings.ToUpper(query)]; ok {
		output = []string{
			fmt.Sprintf(`%s.vitamin. 1 IN TXT "Common name: %s"`, query, vitaminData.CommonName),
			fmt.Sprintf(`%s.vitamin. 1 IN TXT "Scientific name: %s"`, query, vitaminData.ScientificName),
			fmt.Sprintf(`%s.vitamin. 1 IN TXT "Sources: %s"`, query, strings.Join(vitaminData.Sources, ", ")),
		}
	} else {
		output = []string{fmt.Sprintf(`%s.vitamin. 1 IN TXT "Vitamin %s not found"`, query, query)}
	}

	return output, nil
}

// Dump is not implemented in this package.
func (v *VitaminStore) Dump() ([]byte, error) {
	return nil, nil
}

