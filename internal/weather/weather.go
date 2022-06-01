package weather

import (
	"compress/gzip"
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
	"github.com/miekg/dns"
)

const apiURL = "https://api.met.no/weatherapi/locationforecast/2.0/compactx?lat=%0.5f&lon=%0.5f"

// Weather fetches weather forecasts for a given geo location.
type Weather struct {
	data map[string]entry
	mut  sync.RWMutex

	opt    Opt
	geo    *geo.Geo
	client *http.Client
}

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
	MaxEntries int

	CacheTTL   time.Duration
	ReqTimeout time.Duration
	UserAgent  string
}

func New(o Opt, g *geo.Geo) *Weather {
	w := &Weather{
		data: make(map[string]entry),
		opt:  o,
		geo:  g,
		client: &http.Client{
			Timeout: o.ReqTimeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   15,
				ResponseHeaderTimeout: o.ReqTimeout,
			},
		},
	}

	return w
}

// Query queries the weather for a given location.
func (w *Weather) Query(q string) ([]dns.RR, error) {
	locs := w.geo.Query(q)
	if locs == nil {
		return nil, errors.New("unknown city.")
	}

	out := make([]dns.RR, 0, len(locs)*3)
	for n, l := range locs {
		data, err := w.get(l)
		if err != nil {
			return nil, err
		}

		zone, err := time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}

		for _, f := range data.Forecasts {
			r, err := dns.NewRR(fmt.Sprintf("%s 1 TXT \"%s (%s)\" \"%0.2fC (%0.2fF)\" \"%0.2f%% hu.\" \"%s\" \"%s\"",
				q, l.Name, l.Country, f.TempC, f.TempF, f.Humidity, f.Forecast1H, f.Time.In(zone).Format("15:04, Mon")))
			if err != nil {
				return nil, err
			}

			out = append(out, r)
		}

		if n > 2 {
			break
		}
	}

	return out, nil
}

func (w *Weather) get(l geo.Location) (entry, error) {
	w.mut.RLock()
	data, ok := w.data[l.ID]
	w.mut.RUnlock()

	// If data isn't cached, fetch it.
	if !ok || data.ExpiresAt.Before(time.Now()) {
		res, err := w.fetchAPI(l.Lat, l.Lon)

		// Even if it's an error, cache to avoid flooding the service.
		w.mut.Lock()
		w.data[l.ID] = res
		w.mut.Unlock()

		if err != nil {
			log.Printf("error fetching weather API: %v", err)
			return res, errors.New("error fetching weather data.")
		}

		data = res
	}

	if !data.Valid {
		return entry{}, errors.New("weather data is currently unavailable.")
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
	intval := time.Hour * 4
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
			if out.Forecasts[len(out.Forecasts)-1].Time.Add(intval).After(p.Time) {
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
