package stock

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type Stock struct{}

// New returns a new instance of Stock.
func New() *Stock {
	return &Stock{}
}

// Query returns the stock price for a given symbol
func (c *Stock) Query(q string) ([]string, error) {
	res, err := http.Get("https://finance.yahoo.com/quote/" + q)
	if err != nil {
		return nil, errors.New("invalid stock symbol")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("status code"+strconv.Itoa(res.StatusCode)+": "+res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.New("cannot parse html")
	}

	marketPrice := doc.Find("[data-field=regularMarketPrice][data-symbol=" + q + "]").Text()

	if marketPrice == "" {
		return nil, errors.New("cannot find stock price")
	}

	r := fmt.Sprintf("%s 1 TXT \"%s\"", q, marketPrice)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (c *Stock) Dump() ([]byte, error) {
	return nil, nil
}
