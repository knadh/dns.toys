package aqi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
)

const (
	apiURL = "https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%f&longitude=%f&hourly=pm10,pm2_5&timezone=auto&forecast_days=1"
	TTL    = 3600
)

type AQI struct {
	geo        *geo.Geo
	httpClient *http.Client
}

type response struct {
	Hourly struct {
		Time  []string  `json:"time"`
		PM10  []float64 `json:"pm10"`
		PM2_5 []float64 `json:"pm2_5"`
	} `json:"hourly"`
}

type Opt struct {
	ReqTimeout time.Duration
}

func New(o Opt, g *geo.Geo) *AQI {
	return &AQI{
		geo: g,
		httpClient: &http.Client{
			Timeout: o.ReqTimeout,
		},
	}
}

func formatTime(timeStr string) string {
	t, err := time.Parse("2006-01-02T15:04", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("15:04, Mon")
}

func (a *AQI) Query(q string) ([]string, error) {
	locs := a.geo.Query(q)
	if locs == nil {
		return nil, errors.New("unknown city")
	}

	loc := locs[0]
	url := fmt.Sprintf(apiURL, loc.Lat, loc.Lon)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching AQI data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var aqiResp response
	if err := json.Unmarshal(body, &aqiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Find current hour index
	currentTime := time.Now()
	currentIdx := -1
	for i, timeStr := range aqiResp.Hourly.Time {
		t, err := time.Parse("2006-01-02T15:04", timeStr)
		if err != nil {
			continue
		}
		if t.After(currentTime) {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 || currentIdx+5 > len(aqiResp.Hourly.Time) {
		return nil, fmt.Errorf("no forecast data available")
	}

	// Generate responses for next 5 hours
	var results []string
	for i := currentIdx; i < currentIdx+5; i++ {
		result := fmt.Sprintf("%s %d TXT \"%s (%s)\" \"PM10: %.1f \" \"PM2.5: %.1f \" \"%s\"",
			q,
			TTL,
			loc.Name,
			loc.Country,
			aqiResp.Hourly.PM10[i],
			aqiResp.Hourly.PM2_5[i],
			formatTime(aqiResp.Hourly.Time[i]))
		results = append(results, result)
	}

	return results, nil
}

func (a *AQI) Dump() ([]byte, error) {
	return nil, nil
}
