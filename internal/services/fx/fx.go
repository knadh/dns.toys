// fx does Foreign Exchange / currency conversions.
package fx

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const apiURL = "https://open.er-api.com/v6/latest/USD"

// TTL is set to 900 seconds (15 minutes).
const TTL = 900

var reParse = regexp.MustCompile("([0-9\\.]*)([A-Z]{3})\\-([A-Z]{3})")

// FX represents the currency coversion (Foreign Exchange) package.
type FX struct {
	opt  Opt
	data data
	mut  sync.RWMutex
}

type data struct {
	Base  string             `json:"base_code"`
	Date  string             `json:"time_last_update_utc"`
	Rates map[string]float64 `json:"rates"`
}

// Opt represents the config options for the FX converter.
type Opt struct {
	RefreshInterval time.Duration `json:"refresh_interval"`
}

// New returns an instace of the FX converter.
func New(o Opt) *FX {
	fx := &FX{
		opt: o,
	}

	// Periodically fetch and refresh the rates.
	go func() {
		for {
			log.Println("loading fx API")
			d, err := fx.load(apiURL)
			if err != nil {
				log.Printf("error loading fx rates API: %v", err)

				// HTTP fetch failed. Retry again in a minute.
				time.Sleep(time.Minute)
				continue
			}

			if _, ok := d.Rates[d.Base]; !ok {
				log.Printf("base currency %s not found in rates", d.Base)
				time.Sleep(time.Minute * 5)
				continue
			}
			log.Printf("%d fx currency pairs loaded", len(d.Rates))

			fx.mut.Lock()
			fx.data = d
			fx.mut.Unlock()

			time.Sleep(o.RefreshInterval)
		}
	}()

	return fx
}

// Query handles a currency rate conversion query.
// Format: 100USD-INR.FX
func (fx *FX) Query(q string) ([]string, error) {
	if len(fx.data.Rates) == 0 {
		return nil, errors.New("fx data unavailable. Please try later.")
	}

	q = strings.ToUpper(q)

	res := reParse.FindStringSubmatch(q)
	if len(res) != 4 {
		return nil, errors.New("invalid fx query.")
	}

	strVal := res[1]
	// If no value is provided, default to 1.
	if strVal == "" {
		strVal = "1"
	}

	// Parse the numeric value.
	val, err := strconv.ParseFloat(strVal, 32)
	if err != nil {
		return nil, errors.New("invalid number.")
	}

	var (
		from = res[2]
		to   = res[3]
	)

	// Validate the currency names.
	fx.mut.RLock()
	fromRate, ok := fx.data.Rates[from]
	fx.mut.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown from currency '%s'.", from)
	}

	toRate, ok := fx.data.Rates[to]
	if !ok {
		return nil, fmt.Errorf("unknown to currency '%s'.", to)
	}

	baseRate := fx.data.Rates[fx.data.Base]

	// Convert.
	conv := (baseRate / fromRate) / (baseRate / toRate) * val

	r := fmt.Sprintf("%s %d TXT \"%0.2f %s = %0.2f %s\" \"%s\"", q, TTL, val, from, conv, to, fx.data.Date)

	return []string{r}, nil
}

// Dump produces a gob dump of the cached data.
func (fx *FX) Dump() ([]byte, error) {
	buf := &bytes.Buffer{}

	fx.mut.RLock()
	defer fx.mut.RUnlock()
	if err := gob.NewEncoder(buf).Encode(fx.data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Load loads a gob dump of cached data.
func (fx *FX) Load(b []byte) error {
	buf := bytes.NewBuffer(b)

	fx.mut.RLock()
	defer fx.mut.RUnlock()

	err := gob.NewDecoder(buf).Decode(&fx.data)
	return err
}

func (fx *FX) load(url string) (data, error) {
	client := http.Client{
		Timeout: 6 * time.Second,
	}

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return data{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return data{}, fmt.Errorf("request failed: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data{}, err
	}

	var out data
	if err := json.Unmarshal(body, &out); err != nil {
		return data{}, err
	}

	return out, nil
}
