package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/miekg/dns"
)

// Service represents a Service that responds to a particular kind
// of DNS query.
type Service interface {
	Query(string) ([]string, error)
	Dump() ([]byte, error)
}

type handlers struct {
	services map[string]Service
	domain   string
	help     []dns.RR
}

var reClean = regexp.MustCompile("[^a-zA-Z0-9/\\-\\.]")

// register registers a Service for a given query suffix on the DNS server.
// A Service responds to a DNS query via Query().
func (h *handlers) register(suffix string, s Service, mux *dns.ServeMux) func(w dns.ResponseWriter, r *dns.Msg) {
	f := func(w dns.ResponseWriter, r *dns.Msg) {
		m := &dns.Msg{}
		m.SetReply(r)
		m.Compress = false

		if r.Opcode != dns.OpcodeQuery {
			w.WriteMsg(m)
			return
		}

		if len(m.Question) > 5 {
			respErr(errors.New("too many queries."), w, m)
			return
		}

		// Execute the service on all the questions.
		out := []dns.RR{}
		for _, q := range m.Question {
			if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
				continue
			}

			// Call the service with the incoming query.
			// Strip the service suffix from the query eg: mumbai.time.
			ans, err := s.Query(cleanQuery(q.Name, "."+suffix+"."))
			if err != nil {
				respErr(err, w, m)
				return
			}

			// Convert string responses to dns.RR{}.
			o, err := makeResp(ans)
			if err != nil {
				log.Printf("error preparing response: %v", err)
				respErr(errors.New("error preparing response."), w, m)
				return
			}

			out = append(out, o...)
		}

		// Write the response.
		m.Answer = out
		w.WriteMsg(m)
	}

	h.services[suffix] = s
	mux.HandleFunc(suffix+".", f)
	return f
}

// handleMyIP returns the client's IP address as a DNS response.
// Although it is a service, it's not registered like a Service as it
// uses w.RemoteAddr() instead of m.Question unlike a typical service.
func (h *handlers) handleMyIP(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false

	for _, q := range m.Question {
		if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
			continue
		}

		a := strings.Split(w.RemoteAddr().String(), ":")
		if len(a) != 2 {
			respErr(errors.New("unable to detect IP."), w, m)
			w.WriteMsg(m)
			return
		}

		rr, err := dns.NewRR(fmt.Sprintf("myip. 1 TXT \"%s\"", a[0]))
		if err != nil {
			lo.Printf("error preparing myip response: %v", err)
			return
		}

		m.Answer = append(m.Answer, rr)
	}

	w.WriteMsg(m)
}

func (h *handlers) handleHelp(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false
	m.Answer = h.help
	w.WriteMsg(m)
}

func (h *handlers) handleDefault(w dns.ResponseWriter, m *dns.Msg) {
	respErr(fmt.Errorf(`unknown query. try: dig help @%s`, h.domain), w, m)
	w.WriteMsg(m)
}

// respErr writes an error message to a DNS response.
func respErr(err error, w dns.ResponseWriter, m *dns.Msg) {
	r, err := dns.NewRR(fmt.Sprintf(". 1 IN TXT \"error: %s\"", err.Error()))
	if err != nil {
		lo.Println(err)
		return
	}

	m.Rcode = dns.RcodeServerFailure
	m.Extra = []dns.RR{r}

	w.WriteMsg(m)
}

// cleanQuery removes all non-alpha chars, and trims the service suffix
// from the given query string.
func cleanQuery(q, trimSuffix string) string {
	return reClean.ReplaceAllString(strings.TrimSuffix(q, trimSuffix), "")
}

// makeResp converts a []string of DNS responses to []dns.RR.
func makeResp(ans []string) ([]dns.RR, error) {
	out := make([]dns.RR, 0, len(ans))
	for _, a := range ans {
		r, err := dns.NewRR(a)
		if err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, nil
}
