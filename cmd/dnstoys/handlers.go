package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/knadh/dns.toys/internal/fx"
	"github.com/knadh/dns.toys/internal/timezones"
	"github.com/knadh/dns.toys/internal/weather"
	"github.com/miekg/dns"
)

type handlers struct {
	tz      *timezones.Timezones
	weather *weather.Weather
	fx      *fx.FX

	domain string
	help   []dns.RR
}

var reClean = regexp.MustCompile("[^a-z/]")

// handle wraps all query query handlers with general query handling.
func handle(cb func(m *dns.Msg, w dns.ResponseWriter) ([]dns.RR, error)) func(w dns.ResponseWriter, r *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := &dns.Msg{}
		m.SetReply(r)
		m.Compress = false

		if r.Opcode == dns.OpcodeQuery {
			if len(m.Question) > 5 {
				respErr(errors.New("too many queries."), m)
				return
			}

			// Execute the handler.
			res, err := cb(m, w)
			if err != nil {
				respErr(err, m)
			} else {
				m.Answer = res
			}
		}

		// Write the response.
		w.WriteMsg(m)
	}
}

func (h *handlers) handleTime(m *dns.Msg, w dns.ResponseWriter) ([]dns.RR, error) {
	var out []dns.RR
	for _, q := range m.Question {
		if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
			continue
		}

		query := cleanQuery(q.Name, ".time.")
		ans, err := h.tz.Query(query)
		if err != nil {
			return nil, err
		}

		out = append(out, ans...)
	}

	return out, nil
}

func (h *handlers) handleFX(m *dns.Msg, w dns.ResponseWriter) ([]dns.RR, error) {
	var out []dns.RR
	for _, q := range m.Question {
		if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
			continue
		}

		ans, err := h.fx.Query(strings.TrimSuffix(q.Name, ".fx."))
		if err != nil {
			return nil, err
		}

		out = append(out, ans...)
	}

	return out, nil
}

func (h *handlers) handleMyIP(m *dns.Msg, w dns.ResponseWriter) ([]dns.RR, error) {
	var out []dns.RR
	for _, q := range m.Question {
		if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
			continue
		}

		a := strings.Split(w.RemoteAddr().String(), ":")
		if len(a) != 2 {
			return nil, errors.New("unable to detect IP.")
		}

		r, err := dns.NewRR(fmt.Sprintf("myip. TXT %s", a[0]))
		if err != nil {
			return nil, err
		}

		return []dns.RR{r}, nil
	}

	return out, nil
}

func (h *handlers) handleWeather(m *dns.Msg, w dns.ResponseWriter) ([]dns.RR, error) {
	var out []dns.RR
	for _, q := range m.Question {
		if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
			continue
		}

		query := cleanQuery(q.Name, ".weather.")
		ans, err := h.weather.Query(query)
		if err != nil {
			return nil, err
		}

		out = append(out, ans...)
	}

	return out, nil
}

func (h *handlers) handleHelp(m *dns.Msg, w dns.ResponseWriter) ([]dns.RR, error) {
	return h.help, nil
}

func (h *handlers) handleDefault(m *dns.Msg, w dns.ResponseWriter) ([]dns.RR, error) {
	return nil, fmt.Errorf(`unknown query. Try dig help @%s`, h.domain)
}

func makeHelp() []dns.RR {
	var (
		domain = ko.String("server.domain")
		help   = [][]string{}
		out    = []dns.RR{}
	)

	if ko.Bool("timezones.enabled") {
		help = append(help, []string{"get time for a city or country code (london.time, in.time ...)", "dig london.time @%s"})
	}
	if ko.Bool("fx.enabled") {
		help = append(help, []string{"convert currency rates (25USD-EUR.fx, 99.5JPY-INR.fx)", "dig 25USD-EUR.fx @%s"})
	}

	for _, h := range help {
		r, err := dns.NewRR(fmt.Sprintf("help. TXT \"%s\" \"%s\"", h[0], fmt.Sprintf(h[1], domain)))
		if err != nil {
			lo.Fatalf("error preparing help responses.")
		}

		out = append(out, r)
	}

	return out
}

// respErr applies a response error to the incoming query.
func respErr(err error, m *dns.Msg) {
	r, err := dns.NewRR(fmt.Sprintf(". 1 IN TXT \"error: %s\"", err.Error()))
	if err != nil {
		lo.Println(err)
		return
	}

	m.Rcode = dns.RcodeServerFailure
	m.Extra = []dns.RR{r}
}

func cleanQuery(q, trimSuffix string) string {
	return reClean.ReplaceAllString(strings.ToLower(strings.TrimSuffix(q, trimSuffix)), "")
}
