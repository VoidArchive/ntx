package nepse

import (
	"context"
	"fmt"

	"github.com/voidarchive/go-nepse"
)

// Client wraps the go-nepse client for fetching market data.
type Client struct {
	client *nepse.Client
}

// Price represents the last traded price for a symbol.
type Price struct {
	Symbol string
	LTP    float64 // Last traded price in NPR
}

// NewClient creates a new NEPSE client.
func NewClient() (*Client, error) {
	opts := nepse.DefaultOptions()
	// NEPSE server has TLS issues
	opts.TLSVerification = false

	client, err := nepse.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create NEPSE client: %w", err)
	}

	return &Client{client: client}, nil
}

// Close releases resources used by the client.
func (c *Client) Close() error {
	return c.client.Close()
}

// GetPrice fetches the current price for a single symbol.
func (c *Client) GetPrice(ctx context.Context, symbol string) (*Price, error) {
	details, err := c.client.CompanyBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}

	return &Price{
		Symbol: symbol,
		LTP:    details.LastTradedPrice,
	}, nil
}

// GetPrices fetches current prices for multiple symbols.
// Returns a map of symbol -> price. Missing symbols are omitted.
func (c *Client) GetPrices(ctx context.Context, symbols []string) (map[string]*Price, error) {
	prices := make(map[string]*Price, len(symbols))

	for _, symbol := range symbols {
		price, err := c.GetPrice(ctx, symbol)
		if err != nil {
			// Log but continue - some symbols may not have prices
			continue
		}
		prices[symbol] = price
	}

	return prices, nil
}
