package units

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type fileData struct {
	BaseSymbol string `json:"base_symbol"`
	BaseName   string `json:"base_name"`
	Units      []unit `json:"units"`
}

type group struct {
	Name       string `json:"name"`
	BaseSymbol string `json:"base_symbol"`
}

type unit struct {
	Symbol string  `json:"symbol"`
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
}

// Units does conversions for physical units.
type Units struct {
	// List of valid conversions by category.
	// category: {from, to: true} .
	units map[string]map[string]unit

	// symbol -> category map.
	symbols map[string]group

	// Units list helptext.
	help []string
}

//go:embed units.json
var dataB []byte

var reParse = regexp.MustCompile(`(?i)([0-9\.]+)([a-z]{1,6})\-([a-z]{1,6})`)

// New returns a new instance of Units.
func New() (*Units, error) {
	u := &Units{
		units:   make(map[string]map[string]unit),
		symbols: make(map[string]group),
	}

	if err := u.load(dataB); err != nil {
		return nil, err
	}

	u.help = u.printUnitsList()
	return u, nil
}

// Query parses a unit conversion string and returns the results.
func (u *Units) Query(q string) ([]string, error) {
	if q == "unit." {
		return u.help, nil
	}

	res := reParse.FindStringSubmatch(q)
	if len(res) != 4 {
		return nil, errors.New("invalid unit query.")
	}

	// Parse the numeric value.
	val, err := strconv.ParseFloat(res[1], 32)
	if err != nil {
		return nil, errors.New("invalid number.")
	}

	var (
		fromSym = res[2]
		toSym   = res[3]
	)

	// Validate unit symbols.
	g, ok := u.symbols[fromSym]
	if !ok {
		// Try lowercase.
		f := strings.ToLower(fromSym)
		lg, ok := u.symbols[f]
		if !ok {
			return nil, fmt.Errorf("unknown unit: %v. 'dig unit' to see list of units.", fromSym)
		}

		g = lg
		fromSym = f
	}
	from := u.units[g.Name][fromSym]

	// Get the to unit irrespective of the group so that its name
	// can be used in case of an error.
	toG, ok := u.symbols[toSym]
	if !ok {
		// Try lowercase.
		f := strings.ToLower(toSym)
		lg, ok := u.symbols[f]
		if !ok {
			return nil, fmt.Errorf("unknown unit: %v. 'dig unit' to see list of units.", toSym)
		}

		toG = lg
		toSym = f
	}
	toReal := u.units[toG.Name][toSym]

	// The from->to conversion is only valid if the to symbol belongs to the same
	// group as the form symbol.
	to, ok := u.units[g.Name][toSym]
	if !ok {
		return nil, fmt.Errorf("cannot convert %s (%s) to %s (%s).",
			fromSym, from.Name, toSym, toReal.Name)
	}

	baseRate := u.units[g.Name][g.BaseSymbol].Value

	// Convert.
	conv := (baseRate / from.Value) / (baseRate / to.Value) * val

	r := fmt.Sprintf("%s 1 TXT \"%0.2f %s (%s) = %0.2f %s (%s)\"",
		q, val, from.Name, from.Symbol, conv, to.Name, to.Symbol)

	return []string{r}, nil
}

// Dump is not implemented in this package.
func (u *Units) Dump() ([]byte, error) {
	return nil, nil
}

func (u *Units) printUnitsList() []string {
	var (
		out    = make([]string, 0, len(u.symbols))
		groups = make([]string, 0, len(u.units))
	)

	// Collect the group names and sort.
	for g := range u.units {
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i] < groups[j]
	})

	for _, g := range groups {
		units := u.units[g]

		// Collect the units in the group in a list.
		list := make([]unit, 0, len(units))
		for _, un := range units {
			list = append(list, un)
		}

		// Sort the list.
		sort.Slice(list, func(i, j int) bool {
			return list[i].Symbol < list[j].Symbol
		})

		for _, un := range list {
			l := fmt.Sprintf("unit. 1 TXT \"%s\" \"%s (%s)\"", g, un.Symbol, un.Name)
			out = append(out, l)
		}
	}

	return out
}

func (u *Units) load(b []byte) error {
	data := map[string]fileData{}

	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	// Prepare the list of valid from-to conversions.
	for groupName, g := range data {
		u.units[groupName] = map[string]unit{}
		for _, un := range g.Units {
			u.symbols[un.Symbol] = group{
				Name:       groupName,
				BaseSymbol: g.BaseSymbol,
			}
			u.units[groupName][un.Symbol] = un
		}
	}

	return nil
}
