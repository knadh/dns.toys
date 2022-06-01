// fx does Foreign Exchange / currency conversions.
package fx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const apiURL = "https://api.apilayer.com/exchangerates_data/latest?base=USD"

var reParse = regexp.MustCompile("([0-9\\.]+)([A-Z]{3})\\-([A-Z]{3})\\.FX")

// FX represents the currency coversion (Foreign Exchange) package.
type FX struct {
	opt  Opt
	data data
	mut  sync.RWMutex
}

type data struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

// Opt represents the config options for the FX converter.
type Opt struct {
	APIkey          string        `json:"api_key"`
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
			}

			if _, ok := d.Rates[d.Base]; !ok {
				log.Printf("base currency %s not found in rates", d.Base)
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
// 100USD-INR.FX
func (fx *FX) Query(q string) ([]dns.RR, error) {
	res := reParse.FindStringSubmatch(strings.ToUpper(q))
	if len(res) != 4 {
		return nil, errors.New("invalid fx query.")
	}

	// Parse the numeric value.
	val, err := strconv.ParseFloat(res[1], 32)
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
	if !ok {
		return nil, fmt.Errorf("unknown from currency '%s'.", from)
	}

	toRate, ok := fx.data.Rates[to]
	if !ok {
		return nil, fmt.Errorf("unknown to currency '%s'.", to)
	}

	baseRate := fx.data.Rates[fx.data.Base]
	fx.mut.RUnlock()

	// Convert.
	conv := (baseRate / fromRate) / (baseRate / toRate) * val

	out, err := dns.NewRR(fmt.Sprintf("%s TXT \"%0.2f %s = %0.2f %s\" \"%s\"", q, val, from, conv, to, fx.data.Date))
	if err != nil {
		return nil, err
	}

	return []dns.RR{out}, nil
}

func (fx *FX) load(url string) (data, error) {
	client := http.Client{
		Timeout: 6 * time.Second,
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apiKey", fx.opt.APIkey)

	resp, err := client.Do(req)
	if err != nil {
		return data{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return data{}, fmt.Errorf("request failed: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data{}, err
	}

	var out data
	if err := json.Unmarshal(body, &out); err != nil {
		return data{}, err
	}

	return out, nil
}
