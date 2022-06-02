package main

import (
	"errors"
	"fmt"
	"log"
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
func handle(cb func(m *dns.Msg, w dns.ResponseWriter) ([]string, error)) func(w dns.ResponseWriter, r *dns.Msg) {
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
				out := make([]dns.RR, 0, len(res))
				for _, l := range res {
					r, err := dns.NewRR(l)
					if err != nil {
						log.Printf("error preparing response: %v", err)
						respErr(err, m)
						return
					}

					out = append(out, r)
				}

				m.Answer = out
			}
		}

		// Write the response.
		w.WriteMsg(m)
	}
}

func (h *handlers) handleTime(m *dns.Msg, w dns.ResponseWriter) ([]string, error) {
	var out []string
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

func (h *handlers) handleFX(m *dns.Msg, w dns.ResponseWriter) ([]string, error) {
	var out []string
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

func (h *handlers) handleMyIP(m *dns.Msg, w dns.ResponseWriter) ([]string, error) {
	var out []string
	for _, q := range m.Question {
		if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
			continue
		}

		a := strings.Split(w.RemoteAddr().String(), ":")
		if len(a) != 2 {
			return nil, errors.New("unable to detect IP.")
		}

		return []string{fmt.Sprintf("myip. TXT %s", a[0])}, nil
	}

	return out, nil
}

func (h *handlers) handleWeather(m *dns.Msg, w dns.ResponseWriter) ([]string, error) {
	var out []string
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

func (h *handlers) handleHelp(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false
	m.Answer = h.help
	w.WriteMsg(m)
}

func (h *handlers) handleDefault(m *dns.Msg, w dns.ResponseWriter) ([]string, error) {
	return nil, fmt.Errorf(`unknown query. Try dig help @%s`, h.domain)
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
