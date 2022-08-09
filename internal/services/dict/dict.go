package dict

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hold7door/wnram"
)

const (
	Verb      = wnram.Verb
	Adjective = wnram.Adjective
	Adverb    = wnram.Adverb
	Noun      = wnram.Noun
)

type clusterGroupItem struct {
	sem     int
	meaning string
	example string
}

// Dict implements a WordNet based English dictionary services.
type Dict struct {
	opt Opt

	wn *wnram.Handle
}

// Opt represents the external options/config required by this service.
type Opt struct {
	WordNetPath string
	MaxResults  int
}

var posMap = map[wnram.PartOfSpeech]string{
	Verb:      "verb",
	Adjective: "adjective",
	Adverb:    "adverb",
	Noun:      "noun",
}

// New returns a new instance of the WordNet Dictionary service.
func New(o Opt) *Dict {
	log.Printf("loading wordnet data from %s", o.WordNetPath)

	wn, err := wnram.New(o.WordNetPath)
	if err != nil {
		log.Fatalf("error loading worndet: %v", err)
	}
	log.Printf("loaded wordnet dictionary data")

	return &Dict{
		wn:  wn,
		opt: o,
	}
}

// Query queries the WordNet dictionary for an English word.
func (d *Dict) Query(q string) ([]string, error) {
	out, err := d.get(strings.ToLower(q))
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, errors.New("no definitions found.")
	}

	return out, nil
}

func (d *Dict) get(q string) ([]string, error) {
	// Retain the original one with `-` instead of space.
	original := q

	// Replace dashes with spaces for actual lookup.
	q = strings.ReplaceAll(q, "-", " ")

	// Get all semantic clusters of this word.
	clusters, err := d.wn.Lookup(wnram.Criteria{Matching: q, POS: []wnram.PartOfSpeech{Verb, Adverb, Noun, Adjective}})
	if err != nil {
		return nil, err
	}

	// group clusters by part by part of speech
	group := map[wnram.PartOfSpeech][]clusterGroupItem{
		Verb: []clusterGroupItem{},
	}

	for _, f := range clusters {
		totSem := f.NumRelations()

		// Gloss is the string containing the brief definition (“gloss”) and,
		// in most cases, one or more short sentences illustrating the use.
		means, ex := extractExamples(f.Gloss())
		pos := f.POS()

		group[pos] = append(group[pos], clusterGroupItem{
			sem:     totSem,
			meaning: means,
			example: ex,
		})
	}

	var out []string

	// For each group, get the top two entries.
	// The top two entries correspond to most commonly used definitions.
	for _, p := range []wnram.PartOfSpeech{Verb, Adjective, Noun, Adverb} {
		if len(group[p]) == 0 {
			continue
		}

		// Sort each cluster by number of semantic relationships.
		sort.Slice(group[p], func(i, j int) bool {
			return group[p][i].sem > group[p][j].sem
		})

		// Restrict max results.
		if d.opt.MaxResults < len(group[p]) {
			group[p] = group[p][0:d.opt.MaxResults]
		}

		for _, entry := range group[p] {
			if len(entry.example) > 0 {
				out = append(out, fmt.Sprintf("%s 1 TXT \"%s\" \"%s\" \"eg: '%s'\"", original, posMap[p], entry.meaning, entry.example))
			} else {
				out = append(out, fmt.Sprintf("%s 1 TXT \"%s\" \"%s\"", original, posMap[p], entry.meaning))
			}
		}
	}
	return out, nil
}

// Dump is not implemented in this package.
func (d *Dict) Dump() ([]byte, error) {
	return nil, nil
}

// extractExamples extracts examples from a definition string.
// Example input: cause to perform; "run a subject"; "run a process"
func extractExamples(m string) (string, string) {
	var (
		res = strings.Split(m, ";")
		def = res[0]
		eg  = ""
	)

	if len(res) > 1 {
		// Remove unnecessary characters from example string.
		eg = strings.TrimSpace(res[1])
		eg = strings.Trim(eg, "\"")
	}

	return def, eg
}
