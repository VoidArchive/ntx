package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Quote struct {
	Symbol string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	LTP    float64
	Volume float64
}

func main() {
	targetSymbol := "AHPC"

	quote, err := scrapeQuote(targetSymbol)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found quote for %s:\n", quote.Symbol)
	fmt.Printf("  Open: %.2f\n", quote.Open)
	fmt.Printf("  High: %.2f\n", quote.High)
	fmt.Printf("  Low: %.2f\n", quote.Low)
	fmt.Printf("  Close: %.2f\n", quote.Close)
	fmt.Printf("  LTP: %.2f\n", quote.LTP)
	fmt.Printf("  Volume: %.2f\n", quote.Volume)
}

func scrapeQuote(symbol string) (*Quote, error) {
	var foundQuote *Quote

	c := colly.NewCollector(
		colly.AllowedDomains("www.sharesansar.com"),
	)
	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"

	c.OnHTML("table tr", func(e *colly.HTMLElement) {
		if e.ChildText("td:nth-child(2)") == "" {
			return
		}
		cellSymbol := strings.TrimSpace(e.ChildText("td:nth-child(2)"))

		if cellSymbol != symbol {
			return
		}

		quote := &Quote{
			Symbol: cellSymbol,
		}
		if ltp, err := parseFloat(e.ChildText("td:nth-child(3)")); err == nil {
			quote.LTP = ltp
		}
		if open, err := parseFloat(e.ChildText("td:nth-child(6)")); err == nil {
			quote.Open = open
		}
		if high, err := parseFloat(e.ChildText("td:nth-child(7)")); err == nil {
			quote.High = high
		}
		if low, err := parseFloat(e.ChildText("td:nth-child(8)")); err == nil {
			quote.Low = low
		}
		if volume, err := parseFloat(e.ChildText("td:nth-child(9)")); err == nil {
			quote.Volume = volume
		}
		if close, err := parseFloat(e.ChildText("td:nth-child(10)")); err == nil {
			quote.Close = close
		}
		foundQuote = quote
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping: %s", err.Error())
	})

	err := c.Visit("https://www.sharesansar.com/live-trading")
	if err != nil {
		return nil, fmt.Errorf("failed to visit page: %w", err)
	}

	if foundQuote == nil {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}
	return foundQuote, nil
}

func parseFloat(s string) (float64, error) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if cleaned == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseFloat(cleaned, 64)
}

func parseInt(s string) (int64, error) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if cleaned == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseInt(cleaned, 10, 64)
}
