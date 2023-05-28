package weather

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
	"golang.org/x/time/rate"
)

const (
	apiURL = "https://api.met.no/weatherapi/locationforecast/2.0/compact?lat=%0.5f&lon=%0.5f"

	// Max requests/sec allowed by the API.
	apiRateLimit = 15
)

type entry struct {
	Forecasts []forecast
	Location  string
	Timezone  string
	Lat, Lon  float32
	ExpiresAt time.Time
	Valid     bool
}

type forecast struct {
	Time         time.Time
	TempC, TempF float32
	Humidity     float32

	// English weather descriptions.
	Forecast1H string
}

type apiData struct {
	Properties struct {
		Meta struct {
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"meta"`
		Timeseries []struct {
			Time time.Time `json:"time"`
			Data struct {
				Instant struct {
					Details struct {
						AirTemperature   float32 `json:"air_temperature"`
						RelativeHumidity float32 `json:"relative_humidity"`
						WindSpeed        float32 `json:"wind_speed"`
					} `json:"details"`
				} `json:"instant"`
				Next12Hours struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
				} `json:"next_12_hours"`
				Next1Hours struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
				} `json:"next_1_hours"`
				Next6Hours struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
				} `json:"next_6_hours"`
			} `json:"data,omitempty"`
		} `json:"timeseries"`
	} `json:"properties"`
}

// Opt contains config options for Weather.
type Opt struct {
	ForecastInterval time.Duration
	MaxEntries       int

	CacheTTL   time.Duration
	ReqTimeout time.Duration
	UserAgent  string
}

// Weather fetches weather forecasts for a given geo location.
type Weather struct {
	data map[string]entry

	// Queue for defering API fetch requests.
	fetchQueue chan geo.Location

	limiter *rate.Limiter
	mut     sync.RWMutex

	opt    Opt
	geo    *geo.Geo
	client *http.Client
}

var errQueued = errors.New("data is queued.")

func New(o Opt, g *geo.Geo) *Weather {
	w := &Weather{
		data:       make(map[string]entry),
		fetchQueue: make(chan geo.Location, 1000),

		// yr.no API request rate limit.
		limiter: rate.NewLimiter(apiRateLimit, 1),
		opt:     o,
		geo:     g,
		client: &http.Client{
			Timeout: o.ReqTimeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   apiRateLimit,
				ResponseHeaderTimeout: o.ReqTimeout,
			},
		},
	}

	go w.runFetchQueue()

	return w
}

// Query queries the weather for a given location.
func (w *Weather) Query(q string) ([]string, error) {
	locs := w.geo.Query(q)
	if locs == nil {
		return nil, errors.New("unknown city.")
	}

	out := make([]string, 0, len(locs)*3)
	for n, l := range locs {
		data, err := w.get(l)
		if err != nil {
			// Data never existed and has been queued. Show a friendly
			// message instead of an error.
			if err == errQueued {
				r := fmt.Sprintf("%s 1 TXT \"weather data is being fetched. Try again in a few seconds.\"", q)
				return []string{r}, nil
			}

			return nil, err
		}

		zone, err := time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}

		for _, f := range data.Forecasts {
			r := fmt.Sprintf("%s 1 TXT \"%s (%s)\" \"%0.2fC (%0.2fF)\" \"%0.2f%% hu.\" \"%s\" \"%s\"",
				q, l.Name, l.Country, f.TempC, f.TempF, f.Humidity, f.Forecast1H, f.Time.In(zone).Format("15:04, Mon"))
			out = append(out, r)
		}

		if n > 2 {
			break
		}
	}

	return out, nil
}

// Dump produces a gob dump of the cached data.
func (w *Weather) Dump() ([]byte, error) {
	buf := &bytes.Buffer{}

	w.mut.RLock()
	defer w.mut.RUnlock()
	if err := gob.NewEncoder(buf).Encode(w.data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Load loads a gob dump of cached data.
func (w *Weather) Load(b []byte) error {
	buf := bytes.NewBuffer(b)

	w.mut.RLock()
	defer w.mut.RUnlock()

	err := gob.NewDecoder(buf).Decode(&w.data)
	return err
}

func (w *Weather) runFetchQueue() {
	for {
		select {
		case l := <-w.fetchQueue:
			if !w.limiter.Allow() {
				log.Println("weather API rate limit exceeded")
				continue
			}

			res, err := w.fetchAPI(l.Lat, l.Lon)

			// Even if it's an error, cache to avoid flooding the service.
			w.mut.Lock()
			w.data[l.ID] = res
			w.mut.Unlock()

			if err != nil {
				log.Printf("error fetching weather API: %v", err)
				continue
			}
		}
	}
}

func (w *Weather) get(l geo.Location) (entry, error) {
	w.mut.RLock()
	data, ok := w.data[l.ID]
	w.mut.RUnlock()

	if !ok || data.ExpiresAt.Before(time.Now()) {
		// If data is cached but has expired, return the existing data
		// to respond instantly but queue re-fetch in the background to
		// update it for the next request.
		select {
		case w.fetchQueue <- l:
		default:
		}

		// Set the expiry date to the future to not send further
		// requests for the same location until the fetch queue is processed.
		data.ExpiresAt = time.Now().Add(time.Minute)
		w.mut.Lock()
		w.data[l.ID] = data
		w.mut.Unlock()
	}

	if !ok {
		return entry{}, errQueued
	}

	if !data.Valid {
		return entry{}, errors.New("weather data is unavailable. Try again in a few seconds.")
	}

	return data, nil
}

func (w *Weather) fetchAPI(lat, lon float64) (entry, error) {
	// If the request fails, still cache the bad result with a TTL to avoid
	// flooding the upstream with subsequent requests.
	bad := entry{Valid: false, ExpiresAt: time.Now().Add(time.Minute * 10)}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(apiURL, lat, lon), nil)
	if err != nil {
		return bad, err
	}

	req.Header.Add("User-Agent", w.opt.UserAgent)
	req.Header.Add("Accept-Encoding", "gzip")

	r, err := w.client.Do(req)
	if err != nil {
		return bad, err
	}
	defer func() {
		// Drain and close the body to let the Transport reuse the connection
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	var body []byte
	if !r.Uncompressed {
		defer r.Body.Close()
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			return bad, err
		}
		defer reader.Close()

		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return bad, err
		}

		body = b
	} else {
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return bad, err
		}

		body = b
	}

	// Even if the request failed, still cache the invalid request with a TTL
	// so as to not bombard the API.
	if r.StatusCode == http.StatusForbidden || r.StatusCode == http.StatusTooManyRequests {
		return bad, errors.New("error fetching weather data.")
	}

	var data apiData
	if err := json.Unmarshal(body, &data); err != nil {
		return bad, err
	}

	// ExpiresAt header.
	// exp, err := http.ParseTime(r.Header.Get("expires"))
	// if err != nil {
	// 	exp = time.Now().Add(time.Hour * 1)
	// }

	out := entry{
		ExpiresAt: time.Now().Add(w.opt.CacheTTL),
		Valid:     true,
	}

	now := time.Now()
	for _, p := range data.Properties.Timeseries {
		// Skip stale entries.
		if p.Time.Before(now) {
			continue
		}

		f := forecast{
			Time:       p.Time,
			TempC:      p.Data.Instant.Details.AirTemperature,
			TempF:      (p.Data.Instant.Details.AirTemperature * 1.8) + 32.0,
			Forecast1H: p.Data.Next1Hours.Summary.SymbolCode,
			Humidity:   p.Data.Instant.Details.RelativeHumidity,
		}

		// Only pick up entries with with a certain gap.
		if len(out.Forecasts) > 0 {
			if out.Forecasts[len(out.Forecasts)-1].Time.Add(w.opt.ForecastInterval).After(p.Time) {
				continue
			}
		}

		out.Forecasts = append(out.Forecasts, f)

		// Only store 3 days of forecast.
		if len(out.Forecasts) >= w.opt.MaxEntries {
			break
		}
	}

	return out, nil
}
