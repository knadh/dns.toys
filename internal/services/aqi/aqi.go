package aqi

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
	"golang.org/x/time/rate"
)

const (
	apiURL = "https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%f&longitude=%f&hourly=pm10,pm2_5&timezone=auto&forecast_days=2"

	apiRateLimit = 15

	TTL = 3600
)

type entry struct {
	Forecasts []forecast

	ExpiresAt time.Time
	Valid     bool
}

type forecast struct {
	Time        time.Time
	PM10, PM2_5 float32
}

type AQI struct {
	data map[string]entry

	// Queue for defering API fetch requests.
	fetchQueue chan geo.Location

	limiter *rate.Limiter
	mut     sync.RWMutex

	opt    Opt
	geo    *geo.Geo
	client *http.Client
}

type response struct {
	Hourly struct {
		Time  []string  `json:"time"`
		PM10  []float32 `json:"pm10"`
		PM2_5 []float32 `json:"pm2_5"`
	} `json:"hourly"`
}

type Opt struct {
	ForecastInterval time.Duration
	MaxEntries       int

	CacheTTL   time.Duration
	ReqTimeout time.Duration
	UserAgent  string
}

func New(o Opt, g *geo.Geo) *AQI {
	a := &AQI{
		data:       make(map[string]entry),
		fetchQueue: make(chan geo.Location, 1000),

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

	go a.runFetchQueue()

	return a

}

func (a *AQI) Query(q string) ([]string, error) {
	locs := a.geo.Query(q)
	if locs == nil {
		return nil, errors.New("unknown city")
	}

	out := make([]string, 0, len(locs)*3)
	for n, l := range locs {
		data, err := a.get(l)
		if err != nil {
			// Data never existed and has been queued. Show a friendly
			// message instead of an error.
			if err == errQueued {
				r := fmt.Sprintf("%s 1 TXT \"aqi data is being fetched. Try again in a few seconds.\"", q)
				return []string{r}, nil
			}

			return nil, err
		}

		zone, err := time.LoadLocation(l.Timezone)
		if err != nil {
			continue
		}

		for _, f := range data.Forecasts {
			r := fmt.Sprintf("%s %d TXT \"%s (%s)\" \"PM10: %.1f\" \"PM2.5: %.1f\" \"%s\"",
				q, TTL, l.Name, l.Country, f.PM10, f.PM2_5, f.Time.In(zone).Format("15:04, Mon"))
			out = append(out, r)
		}

		if n > 2 {
			break
		}
	}

	return out, nil
}

// Dump produces a gob dump of the cached data.
func (a *AQI) Dump() ([]byte, error) {
	buf := &bytes.Buffer{}

	a.mut.RLock()
	defer a.mut.RUnlock()
	if err := gob.NewEncoder(buf).Encode(a.data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Load loads a gob dump of cached data.
func (a *AQI) Load(b []byte) error {
	buf := bytes.NewBuffer(b)

	a.mut.RLock()
	defer a.mut.RUnlock()

	err := gob.NewDecoder(buf).Decode(&a.data)
	return err
}

func (a *AQI) runFetchQueue() {
	for {
		select {
		case l := <-a.fetchQueue:
			if !a.limiter.Allow() {
				log.Println("aqi API rate limit exceeded")
				continue
			}

			res, err := a.fetchAPI(l.Lat, l.Lon)

			// Even if it's an error, cache to avoid flooding the service.
			a.mut.Lock()
			a.data[l.ID] = res
			a.mut.Unlock()

			if err != nil {
				log.Printf("error fetching aqi API: %v", err)
				continue
			}
		}
	}
}

var errQueued = errors.New("data is queued.")

func (a *AQI) get(l geo.Location) (entry, error) {
	a.mut.RLock()
	data, ok := a.data[l.ID]
	a.mut.RUnlock()

	if !ok || data.ExpiresAt.Before(time.Now()) {
		// If data is cached but has expired, return the existing data
		// to respond instantly but queue re-fetch in the background to
		// update it for the next request.
		select {
		case a.fetchQueue <- l:
		default:
		}

		// Set the expiry date to the future to not send further
		// requests for the same location until the fetch queue is processed.
		data.ExpiresAt = time.Now().Add(time.Minute)
		a.mut.Lock()
		a.data[l.ID] = data
		a.mut.Unlock()
	}

	if !ok {
		return entry{}, errQueued
	}

	if !data.Valid {
		return entry{}, errors.New("aqi data is unavailable. Try again in a few seconds.")
	}

	return data, nil
}

func (a *AQI) fetchAPI(lat, lon float64) (entry, error) {
	// If the request fails, still cache the bad result with a TTL to avoid
	// flooding the upstream with subsequent requests.
	bad := entry{Valid: false, ExpiresAt: time.Now().Add(time.Minute * 10)}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(apiURL, lat, lon), nil)
	if err != nil {
		return bad, err
	}

	req.Header.Add("User-Agent", a.opt.UserAgent)
	req.Header.Add("Accept-Encoding", "gzip")

	r, err := a.client.Do(req)
	if err != nil {
		return bad, err
	}
	defer func() {
		// Drain and close the body to let the Transport reuse the connection
		io.Copy(io.Discard, r.Body)
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

		b, err := io.ReadAll(reader)
		if err != nil {
			return bad, err
		}

		body = b
	} else {
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return bad, err
		}

		body = b
	}

	// Even if the request failed, still cache the invalid request with a TTL
	// so as to not bombard the API.
	if r.StatusCode == http.StatusForbidden || r.StatusCode == http.StatusTooManyRequests {
		return bad, errors.New("error fetching aqi data.")
	}

	var data response
	if err := json.Unmarshal(body, &data); err != nil {
		return bad, err
	}

	out := entry{
		ExpiresAt: time.Now().Add(a.opt.CacheTTL),
		Valid:     true,
	}

	now := time.Now()
	for i, ts := range data.Hourly.Time {

		t, err := time.Parse("2006-01-02T15:04", ts)
		if err != nil {
			log.Printf("error parsing time %s: %v", ts, err)
			continue
		}

		// Skip stale entries.
		if t.Before(now) {
			continue
		}

		f := forecast{
			Time:  t,
			PM10:  data.Hourly.PM10[i],
			PM2_5: data.Hourly.PM2_5[i],
		}

		// Only pick up entries with with a certain gap.
		if len(out.Forecasts) > 0 {
			if out.Forecasts[len(out.Forecasts)-1].Time.Add(a.opt.ForecastInterval).After(t) {
				continue
			}
		}

		out.Forecasts = append(out.Forecasts, f)

		// Only store 1 day of forecast.
		if len(out.Forecasts) >= a.opt.MaxEntries {
			break
		}
	}

	return out, nil
}
