package stock

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Stock struct{}

// New returns a new instance of Stock.
func New() *Stock {
	return &Stock{}
}

func FindStockPrice(symbol string) (string, error) {
	return findStockPriceByUrl(stockPriceUrl(symbol))
}

func findStockPriceByUrl(stockPriceUrl string) (string, error) {
	doc, err := goquery.NewDocument(stockPriceUrl)
	if err != nil {
		return "", errors.New("Your search produces no matches.")
	}

	selection := doc.Find(".wsod_last span")

	if len(selection.Nodes) == 0 {
		return "", errors.New("Your search produces no matches.")
	}

	stockPrice := strings.TrimSpace(strings.Replace(selection.Text(), ",", "", -1))[0:6]

	return stockPrice, nil
}

func stockPriceUrl(symbol string) string {
	return "http://money.cnn.com/quote/quote.html?symb=" + url.QueryEscape(symbol)
}

// Query returns the stock price for a given symbol
func (c *Stock) Query(q string) ([]string, error) {
	res, err := FindStockPrice(q)
	if err != nil {
		return nil, errors.New("invalid stock symbol")
	}

	if res == "" {
		return nil, errors.New("cannot find stock price")
	}

	r := fmt.Sprintf("%s 1 TXT \"%s\"", q, res)
	return []string{r}, nil
}

// Dump is not implemented in this package.
func (c *Stock) Dump() ([]byte, error) {
	return nil, nil
}
