package sky

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	apiURL = "https://api.n2yo.com/rest/v1/satellite/positions/%s/0/0/0/1/&apiKey=%s"

	// Max requests/sec allowed by the API.
	apiRateLimit = 900

	// TTL is set to 1 hour (60*60=3,600).
	TTL = 3600
)

type apiData struct {
	Info struct {
		SatName           string `json:"satname"`
		SatID             int    `json:"satid"`
		TransactionsCount int    `json:"transactionscount"`
	} `json:"info"`
	Positions []struct {
		SatLatitude  float64 `json:"satlatitude"`
		SatLongitude float64 `json:"satlongitude"`
		SatAltitude  float64 `json:"sataltitude"`
		Azimuth      float64 `json:"azimuth"`
		Elevation    float64 `json:"elevation"`
		RA           float64 `json:"ra"`
		Dec          float64 `json:"dec"`
		Timestamp    int64   `json:"timestamp"`
		Eclipsed     bool    `json:"eclipsed"`
	} `json:"positions"`
}

type entry struct {
	ExpiresAt time.Time `json:"expires_at"`
	Data      apiData   `json:"data"`
}

// Opt contains config options for Weather.
type Opt struct {
	CacheTTL   time.Duration
	ReqTimeout time.Duration
	APIKey     string
}

type Sky struct {
	data map[string]entry

	// Queue for defering API fetch requests.
	fetchQueue chan string

	limiter *rate.Limiter
	mut     sync.RWMutex

	opt    Opt
	client *http.Client
}

func New(o Opt) *Sky {
	w := &Sky{
		data:       make(map[string]entry),
		fetchQueue: make(chan string, 1000),

		// yr.no API request rate limit.
		limiter: rate.NewLimiter(apiRateLimit, 1),
		opt:     o,
		client: &http.Client{
			Timeout: o.ReqTimeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   apiRateLimit,
				ResponseHeaderTimeout: o.ReqTimeout,
			},
		},
	}

	return w
}

// Query queries the weather for a given location.
func (w *Sky) Query(q string) ([]string, error) {
	q = strings.ToLower(q)
	// For now, just support ISS.
	if q != "iss" {
		return nil, errors.New("only `ISS` is supported")
	}

	d, err := w.get("25544") // 25544 is the N2YO ID for ISS.
	if err != nil {
		return nil, err
	}

	r := fmt.Sprintf(`%s %d TXT "%s" "lat=%v" "lon=%v" "altitude=%vKM" "azimuth=%v" "elevation=%v" "ra=%v" "time=%s"`,
		q, d.Data.Info.SatID, d.Data.Info.SatName, d.Data.Positions[0].SatLatitude, d.Data.Positions[0].SatLongitude,
		d.Data.Positions[0].SatAltitude, d.Data.Positions[0].Azimuth,
		d.Data.Positions[0].Elevation, d.Data.Positions[0].RA, time.Unix(d.Data.Positions[0].Timestamp, 0).Format(time.RFC3339))
	l := fmt.Sprintf(`%s %d TXT "https://maps.google.com/?q=%v,%v"`, q, d.Data.Info.SatID, d.Data.Positions[0].SatLatitude, d.Data.Positions[0].SatLongitude)

	return []string{r, l}, nil
}

// Dump produces a gob dump of the cached data.
func (w *Sky) Dump() ([]byte, error) {
	return nil, nil
}

// Load loads a gob dump of cached data.
func (w *Sky) Load(b []byte) error {
	return nil
}

func (w *Sky) get(q string) (entry, error) {
	w.mut.RLock()
	data, ok := w.data[q]
	w.mut.RUnlock()

	if ok && data.ExpiresAt.After(time.Now()) {
		return data, nil
	}

	data, err := w.fetchAPI(q)
	if err != nil {
		return entry{}, err
	}

	w.mut.Lock()
	w.data[q] = data
	w.mut.Unlock()

	return data, nil
}

func (w *Sky) fetchAPI(q string) (entry, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(apiURL, q, w.opt.APIKey), nil)
	if err != nil {
		return entry{}, err
	}

	r, err := w.client.Do(req)
	if err != nil {
		return entry{}, err
	}
	defer func() {
		// Drain and close the body to let the Transport reuse the connection
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
	}()

	var body []byte
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return entry{}, err
	}

	body = b

	// Even if the request failed, still cache the invalid request with a TTL
	// so as to not bombard the API.
	if r.StatusCode == http.StatusForbidden || r.StatusCode == http.StatusTooManyRequests {
		return entry{}, errors.New("error fetching data.")
	}

	var data apiData
	if err := json.Unmarshal(body, &data); err != nil {
		return entry{}, err
	}

	out := entry{
		ExpiresAt: time.Now().Add(w.opt.CacheTTL),
		Data:      data,
	}

	return out, nil
}
