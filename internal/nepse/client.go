// Package nepse provides a wrapper around go-nepse for NTX.
package nepse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voidarchive/go-nepse"
)

type Client struct {
	api *nepse.Client
}

func NewClient() (*Client, error) {
	opts := nepse.DefaultOptions()
	opts.TLSVerification = false // NEPSE server has TLS issues

	api, err := nepse.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return &Client{api: api}, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}

// Company represents a listed company.
type Company struct {
	ID        int32
	Symbol    string
	Name      string
	Sector    string
	MarketCap float64
	Shares    int64
}

// Companies returns all listed companies (excluding mutual funds, bonds, promoter shares).
// Uses Securities API for active status, enriched with sector data from Companies API.
func (c *Client) Companies(ctx context.Context) ([]Company, error) {
	// Get companies for sector info
	companyList, err := c.api.Companies(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch companies: %w", err)
	}

	// Build sector map from companies
	sectorMap := make(map[string]string)
	for _, co := range companyList {
		sectorMap[co.Symbol] = co.SectorName
	}

	// Get securities for active status
	securities, err := c.api.Securities(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch securities: %w", err)
	}

	var out []Company
	for _, s := range securities {
		if s.ActiveStatus != "A" {
			continue
		}
		if isPromoterShare(s.Symbol) {
			continue
		}
		if isDebentureOrBond(s.SecurityName) {
			continue
		}

		sector := sectorMap[s.Symbol]
		if sector == "" {
			sector = "Others"
		}
		if sector == "Mutual Fund" {
			continue
		}

		out = append(out, Company{
			ID:     s.ID,
			Symbol: s.Symbol,
			Name:   s.SecurityName,
			Sector: sector,
		})
	}
	return out, nil
}

// Security represents an actively tradable security.
type Security struct {
	ID     int32
	Symbol string
	Name   string
}

// Securities returns all actively tradable securities.
func (c *Client) Securities(ctx context.Context) ([]Security, error) {
	list, err := c.api.Securities(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch securities: %w", err)
	}

	var out []Security
	for _, s := range list {
		if s.ActiveStatus != "A" {
			continue
		}
		if isPromoterShare(s.Symbol) {
			continue
		}
		if isDebentureOrBond(s.SecurityName) {
			continue
		}

		out = append(out, Security{
			ID:     s.ID,
			Symbol: s.Symbol,
			Name:   s.SecurityName,
		})
	}
	return out, nil
}

// isPromoterShare checks if a symbol is a promoter share.
// Promoter shares typically have suffix "P" or "PO".
func isPromoterShare(symbol string) bool {
	if len(symbol) < 2 {
		return false
	}
	return strings.HasSuffix(symbol, "P") || strings.HasSuffix(symbol, "PO")
}

// isDebentureOrBond checks if a name is a debenture or bond.
// Case-insensitive check for "debenture" or "bond" in the name.
func isDebentureOrBond(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "debenture") || strings.Contains(name, "bond")
}

// CompanyDetail contains full company info with current price.
type CompanyDetail struct {
	ID            int32
	Symbol        string
	Name          string
	Sector        string
	Open          float64
	High          float64
	Low           float64
	Close         float64
	LTP           float64
	PreviousClose float64
	Volume        int64
	Week52High    float64
	Week52Low     float64
	LastUpdated   time.Time
}

// Company returns details for a single company.
func (c *Client) Company(ctx context.Context, symbol string) (*CompanyDetail, error) {
	d, err := c.api.CompanyBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("fetch company %s: %w", symbol, err)
	}

	updated, _ := time.Parse("2006-01-02T15:04:05", d.LastUpdatedDateTime)

	return &CompanyDetail{
		ID:            d.ID,
		Symbol:        d.Symbol,
		Name:          d.SecurityName,
		Sector:        d.SectorName,
		Open:          d.OpenPrice,
		High:          d.HighPrice,
		Low:           d.LowPrice,
		Close:         d.ClosePrice,
		LTP:           d.LastTradedPrice,
		PreviousClose: d.PreviousClose,
		Volume:        d.TotalTradeQuantity,
		Week52High:    d.FiftyTwoWeekHigh,
		Week52Low:     d.FiftyTwoWeekLow,
		LastUpdated:   updated,
	}, nil
}

// Price represents live price data.
type Price struct {
	Symbol        string
	Open          float64
	High          float64
	Low           float64
	LTP           float64
	PreviousClose float64
	Change        float64
	ChangePercent float64
	Volume        int64
	Turnover      float64
}

// LivePrices returns current prices for all traded securities.
func (c *Client) LivePrices(ctx context.Context) ([]Price, error) {
	entries, err := c.api.LiveMarket(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch live market: %w", err)
	}

	out := make([]Price, len(entries))
	for i, e := range entries {
		change := e.LastTradedPrice - e.PreviousClose
		out[i] = Price{
			Symbol:        e.Symbol,
			Open:          e.OpenPrice,
			High:          e.HighPrice,
			Low:           e.LowPrice,
			LTP:           e.LastTradedPrice,
			PreviousClose: e.PreviousClose,
			Change:        change,
			ChangePercent: e.PercentageChange,
			Volume:        e.TotalTradeQuantity,
			Turnover:      e.TotalTradeValue,
		}
	}
	return out, nil
}

// Candle represents historical OHLCV data.
type Candle struct {
	Date     string
	Open     float64 // Note: NEPSE doesn't provide open in history
	High     float64
	Low      float64
	Close    float64
	Volume   int64
	Turnover float64
}

// PriceHistory returns historical prices for a symbol.
func (c *Client) PriceHistory(ctx context.Context, symbol, from, to string) ([]Candle, error) {
	history, err := c.api.PriceHistoryBySymbol(ctx, symbol, from, to)
	if err != nil {
		return nil, fmt.Errorf("fetch history for %s: %w", symbol, err)
	}

	out := make([]Candle, len(history))
	for i, h := range history {
		out[i] = Candle{
			Date:     h.BusinessDate,
			Open:     h.ClosePrice, // NEPSE doesn't provide open, use close as fallback
			High:     h.HighPrice,
			Low:      h.LowPrice,
			Close:    h.ClosePrice,
			Volume:   h.TotalTradedQuantity,
			Turnover: h.TotalTradedValue,
		}
	}
	return out, nil
}

// MarketStatus represents current market state.
type MarketStatus struct {
	IsOpen bool
	AsOf   string
}

// MarketStatus returns whether the market is open.
func (c *Client) MarketStatus(ctx context.Context) (*MarketStatus, error) {
	s, err := c.api.MarketStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch market status: %w", err)
	}

	return &MarketStatus{
		IsOpen: s.IsMarketOpen(),
		AsOf:   s.AsOf,
	}, nil
}

// Index represents a market index.
type Index struct {
	Name          string
	Value         float64
	Change        float64
	ChangePercent float64
}

// NepseIndex returns the main NEPSE index.
func (c *Client) NepseIndex(ctx context.Context) (*Index, error) {
	idx, err := c.api.NepseIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch nepse index: %w", err)
	}

	return &Index{
		Name:          "NEPSE",
		Value:         idx.IndexValue,
		Change:        idx.PointChange,
		ChangePercent: idx.PercentChange,
	}, nil
}

// SubIndices returns all sector sub-indices.
func (c *Client) SubIndices(ctx context.Context) ([]Index, error) {
	subs, err := c.api.SubIndices(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch sub-indices: %w", err)
	}

	out := make([]Index, len(subs))
	for i, s := range subs {
		out[i] = Index{
			Name:          s.Index,
			Value:         s.Close,
			Change:        s.Change,
			ChangePercent: s.PerChange,
		}
	}
	return out, nil
}

// TopMover represents a top gainer or loser.
type TopMover struct {
	Symbol        string
	LTP           float64
	Change        float64
	ChangePercent float64
	Volume        int64
	Turnover      float64
}

// TopGainers returns top gaining stocks.
func (c *Client) TopGainers(ctx context.Context) ([]TopMover, error) {
	list, err := c.api.TopGainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch top gainers: %w", err)
	}

	out := make([]TopMover, len(list))
	for i, e := range list {
		out[i] = TopMover{
			Symbol:        e.Symbol,
			LTP:           e.LTP,
			Change:        e.PointChange,
			ChangePercent: e.PercentageChange,
		}
	}
	return out, nil
}

// TopLosers returns top losing stocks.
func (c *Client) TopLosers(ctx context.Context) ([]TopMover, error) {
	list, err := c.api.TopLosers(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch top losers: %w", err)
	}

	out := make([]TopMover, len(list))
	for i, e := range list {
		out[i] = TopMover{
			Symbol:        e.Symbol,
			LTP:           e.LTP,
			Change:        e.PointChange,
			ChangePercent: e.PercentageChange,
		}
	}
	return out, nil
}

// CompanyProfile represents company contact and profile info.
type CompanyProfile struct {
	Name          string
	Email         string
	Profile       string
	ContactPerson string
	Address       string
	Phone         string
}

// CompanyProfile returns profile info for a company.
func (c *Client) CompanyProfile(ctx context.Context, symbol string) (*CompanyProfile, error) {
	p, err := c.api.CompanyProfileBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("fetch profile for %s: %w", symbol, err)
	}

	return &CompanyProfile{
		Name:          p.CompanyName,
		Email:         p.CompanyEmail,
		Profile:       p.CompanyProfile,
		ContactPerson: p.CompanyContactPerson,
		Address:       p.AddressField,
		Phone:         p.PhoneNumber,
	}, nil
}

// Report represents a quarterly or annual financial report.
type Report struct {
	ReportType    string // "annual" or "quarterly"
	FiscalYear    string
	Quarter       int
	EPS           float64
	PE            float64
	BookValue     float64 // Net worth per share
	PaidUpCapital float64
	Profit        float64
	PublishedAt   string
}

// Reports returns quarterly and annual reports for a company.
func (c *Client) Reports(ctx context.Context, symbol string) ([]Report, error) {
	reports, err := c.api.ReportsBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("fetch reports for %s: %w", symbol, err)
	}

	out := make([]Report, 0, len(reports))
	for _, r := range reports {
		if r.FiscalReport == nil {
			continue
		}

		report := Report{
			EPS:           r.FiscalReport.EPSValue,
			PE:            r.FiscalReport.PEValue,
			BookValue:     r.FiscalReport.NetWorthPerShare,
			PaidUpCapital: r.FiscalReport.PaidUpCapital,
			Profit:        r.FiscalReport.ProfitAmount,
			PublishedAt:   r.ModifiedDate,
		}

		if r.FiscalReport.FinancialYear != nil {
			report.FiscalYear = r.FiscalReport.FinancialYear.FYNameNepali
		}

		if r.IsAnnual() {
			report.ReportType = "annual"
			report.Quarter = 0
		} else {
			report.ReportType = "quarterly"
			if r.FiscalReport.QuarterMaster != nil {
				switch r.FiscalReport.QuarterMaster.QuarterName {
				case "First Quarter":
					report.Quarter = 1
				case "Second Quarter":
					report.Quarter = 2
				case "Third Quarter":
					report.Quarter = 3
				case "Fourth Quarter":
					report.Quarter = 4
				}
			}
		}

		out = append(out, report)
	}

	return out, nil
}

// Dividend represents a dividend declaration.
type Dividend struct {
	FiscalYear   string
	CashPercent  float64
	BonusPercent float64
	Headline     string
	PublishedAt  string
}

// Dividends returns dividend history for a company.
func (c *Client) Dividends(ctx context.Context, symbol string) ([]Dividend, error) {
	divs, err := c.api.DividendsBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("fetch dividends for %s: %w", symbol, err)
	}

	out := make([]Dividend, 0, len(divs))
	for _, d := range divs {
		if d.CompanyNews == nil || d.CompanyNews.DividendsNotice == nil {
			continue
		}

		notice := d.CompanyNews.DividendsNotice
		div := Dividend{
			CashPercent:  notice.CashDividend,
			BonusPercent: notice.BonusShare,
			Headline:     d.CompanyNews.NewsHeadline,
			PublishedAt:  d.ModifiedDate,
		}

		if notice.FinancialYear != nil {
			div.FiscalYear = notice.FinancialYear.FYNameNepali
		}

		out = append(out, div)
	}

	return out, nil
}
