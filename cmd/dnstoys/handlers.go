package main

import (
	"errors"
	"fmt"
	"log"
	"net"
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

var reClean = regexp.MustCompile("[^a-zA-Z0-9/\\-\\.:,]")

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

// handleEchoIP returns the client's IP address as a DNS response.
// Although it is a service, it's not registered like a Service as it
// uses w.RemoteAddr() instead of m.Question unlike a typical service.
func (h *handlers) handleEchoIP(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false

	for _, q := range m.Question {
		if q.Qtype != dns.TypeTXT && q.Qtype != dns.TypeA {
			continue
		}

		// Parse the Host:Port.
		h, _, err := net.SplitHostPort(w.RemoteAddr().String())
		if err != nil {
			respErr(errors.New("unable to detect IP."), w, m)
			return
		}

		// Get the IP representaion.
		ip := net.ParseIP(h)
		if ip == nil {
			respErr(errors.New("unable to detect IP."), w, m)
			return
		}

		switch {
		// Handle ipv4.
		case ip.To4() != nil:
			rr, err := dns.NewRR(fmt.Sprintf("ip. 1 TXT \"%s\"", ip.To4().String()))
			if err != nil {
				lo.Printf("error preparing ip response: %v", err)
				return
			}
			m.Answer = append(m.Answer, rr)
		// Handle ipv6.
		case ip.To16() != nil:
			rr, err := dns.NewRR(fmt.Sprintf("ip. 1 TXT \"%s\"", ip.To16().String()))
			if err != nil {
				lo.Printf("error preparing ip response: %v", err)
				return
			}
			m.Answer = append(m.Answer, rr)
		}
	}

	w.WriteMsg(m)
}

// handlePi returns values of pi relevant for the record type.
// TXT  record: "3.141592653589793238462643383279502884197169"
// A    record: 3.141.59.27
// AAAA record: 3141:5926:5358:9793:2384:6264:3383:2795
func (h *handlers) handlePi(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false

	for _, q := range m.Question {
		var rrstr string
		if q.Qtype == dns.TypeTXT {
			rrstr = "pi. 1 TXT 3.141592653589793238462643383279502884197169"
		} else if q.Qtype == dns.TypeA {
			rrstr = "pi. IN A 3.141.59.27"
		} else if q.Qtype == dns.TypeAAAA {
			rrstr = "pi. IN AAAA 3141:5926:5358:9793:2384:6264:3383:2795"
		} else {
			continue
		}
		rr, err := dns.NewRR(rrstr)
		if err != nil {
			lo.Printf("error preparing pi response: %v", err)
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
