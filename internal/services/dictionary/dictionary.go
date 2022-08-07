package dictionary

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hold7door/wnram"
)

type Dictionary struct {
	wn *wnram.Handle
}

type Opt struct {
	WN *wnram.Handle
}

type clusterGroupItem struct {
	sem int
	meaning string
	example string
}


const (
	Verb = wnram.Verb
	Adjective = wnram.Adjective
	Adverb = wnram.Adverb
	Noun = wnram.Noun
)

var POSToString = map[wnram.PartOfSpeech]string{
	Verb: "verb",
	Adjective: "adjective",
	Adverb: "adverb",
	Noun: "noun",
}

func New(o Opt) *Dictionary {
	d := &Dictionary{
		wn: o.WN,
	}

	return d
}

func (d *Dictionary) Query(q string) ([]string, error){
	q = strings.ToLower(q)

	out, err := d.get(q)

	if err != nil {
		r := fmt.Sprintf("%s 1 TXT \"dictionary unavailable, try again later.\"", q)
		return []string{r}, nil
	}

	if len(out) == 0 {
		out = append(out, fmt.Sprintf("%s 1 TXT \"word definition was not found in our dictionary, please try other sources.\"", q))
	}

	return out, nil

   
}

// returns number of semantic relationships from a cluster
func getSemantic(l *wnram.Lookup) (int){
	return l.NumRelations()
}

func getMeaningExample(m string) (string, string){
	res := strings.Split(m, ";")
	means := res[0]
	ex := ""
	if len(res) > 1 {
		// remove unnecessary characters from example string
		ex = strings.Trim(res[1], " ")
		ex = strings.Trim(ex, "\"")
	}
	return means, ex
}

func (d *Dictionary) get(w string) ([]string, error) {
	// get all semantic clusters of this word 
	if clusters, err := d.wn.Lookup(wnram.Criteria{Matching: w, POS: []wnram.PartOfSpeech{Verb, Adverb, Noun, Adjective}}); err != nil {
		return nil, err
	} else {

		// group clusters by part by part of speech
		group := map[wnram.PartOfSpeech][]clusterGroupItem{
			Verb: []clusterGroupItem{},
		}

		for _, f := range clusters {
			totSem := getSemantic(&f)
			// gloss is a string containing
			// a brief definition (“gloss”) and, in most cases, one or more short sentences illustrating the use
			means, ex := getMeaningExample(f.Gloss())
			pos := f.POS()

			group[pos] = append(group[pos], clusterGroupItem{
				sem: totSem,
				meaning: means,
				example: ex,
			})
		}

		var out []string

		// for each group get top two entries
		// top two entries correspond to most commonly used meaning
		for _, p := range []wnram.PartOfSpeech{Verb, Adjective, Noun, Adverb} {
			if len(group[p]) == 0 {
				continue
			}

			// sort each cluster by number of semantic relationships
			sort.Slice(group[p], func(i, j int) bool {
				return group[p][i].sem >  group[p][j].sem
			})

			// get top two entries
			maxSize := 2
			if len(group[p]) < maxSize{
				maxSize = len(group[p]) 
			}
			group[p] = group[p][0 : maxSize]

			// create TXT record in out
			for _, entry := range group[p]{
				if len(entry.example) > 0 {
					out = append(out, fmt.Sprintf("%s 1 TXT \"%s:\" \"%s\" \"%s\"", w, POSToString[p], entry.meaning, entry.example))
				}else{
					out = append(out, fmt.Sprintf("%s 1 TXT \"%s:\" \"%s\"", w, POSToString[p], entry.meaning))
				}
			}
		}
		return out, nil
	}
}

// Dump produces a gob dump of the cached data.
func (d *Dictionary) Dump() ([]byte, error) {
	return nil, nil
}