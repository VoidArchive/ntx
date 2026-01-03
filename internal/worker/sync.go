package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/market"
)

// Sector name mapping from NEPSE API to proto enum integers.
var sectorMap = map[string]int64{
	"Commercial Banks":             1,  // SECTOR_COMMERCIAL_BANK
	"Development Banks":            2,  // SECTOR_DEVELOPMENT_BANK
	"Finance":                      3,  // SECTOR_FINANCE
	"Microfinance":                 4,  // SECTOR_MICROFINANCE
	"Life Insurance":               5,  // SECTOR_LIFE_INSURANCE
	"Non Life Insurance":           6,  // SECTOR_NON_LIFE_INSURANCE
	"Hydro Power":                  7,  // SECTOR_HYDROPOWER
	"Manufacturing And Processing": 8,  // SECTOR_MANUFACTURING
	"Hotels And Tourism":           9,  // SECTOR_HOTEL
	"Trading":                      10, // SECTOR_TRADING
	"Investment":                   11, // SECTOR_INVESTMENT
	"Mutual Fund":                  12, // SECTOR_MUTUAL_FUND
	"Others":                       13, // SECTOR_OTHERS
}

func sectorToInt(name string) int64 {
	if id, ok := sectorMap[name]; ok {
		return id
	}
	return 0 // SECTOR_UNSPECIFIED
}

// syncCompanies fetches all companies from NEPSE and upserts to the database.
func (w *Worker) syncCompanies(ctx context.Context) error {
	start := time.Now()

	companies, err := w.nepse.Companies(ctx)
	if err != nil {
		return err
	}

	for _, c := range companies {
		err := w.queries.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
			Symbol:      c.Symbol,
			Name:        c.Name,
			Sector:      sectorToInt(c.Sector),
			Description: "",
			LogoUrl:     "",
		})
		if err != nil {
			slog.Error("failed to upsert company", "symbol", c.Symbol, "error", err)
			continue
		}
	}

	slog.Info("company sync complete", "count", len(companies), "duration", time.Since(start))
	return nil
}

// syncPrices fetches live prices from NEPSE and upserts to the database.
func (w *Worker) syncPrices(ctx context.Context) error {
	start := time.Now()
	today := time.Now().In(market.NPT).Format(market.DateFormat)

	prices, err := w.nepse.LivePrices(ctx)
	if err != nil {
		return err
	}

	for _, p := range prices {
		err := w.queries.UpsertPrice(ctx, sqlc.UpsertPriceParams{
			Symbol:        p.Symbol,
			Date:          today,
			Open:          p.Open,
			High:          p.High,
			Low:           p.Low,
			Close:         p.LTP, // LTP is the current price during trading
			PreviousClose: sql.NullFloat64{Float64: p.PreviousClose, Valid: true},
			Volume:        p.Volume,
			Turnover:      sql.NullInt64{Int64: int64(p.Turnover), Valid: true},
			IsComplete:    0, // Not complete until market closes
		})
		if err != nil {
			slog.Error("failed to upsert price", "symbol", p.Symbol, "error", err)
			continue
		}
	}

	slog.Info("price sync complete", "count", len(prices), "duration", time.Since(start))
	return nil
}

// finalSnapshot marks the day's prices as complete and records the trading day.
func (w *Worker) finalSnapshot(ctx context.Context) error {
	start := time.Now()
	today := time.Now().In(market.NPT)
	date := today.Format(market.DateFormat)

	// Do one final price sync to capture closing prices
	if err := w.syncPrices(ctx); err != nil {
		slog.Error("final price sync failed", "error", err)
		// Continue anyway to mark day complete
	}

	// Mark all prices for today as complete
	if err := w.queries.MarkPricesComplete(ctx, date); err != nil {
		return err
	}

	// Record this as a trading day
	if err := w.market.RecordTradingDay(ctx, today, true, "completed"); err != nil {
		return err
	}

	slog.Info("final snapshot complete", "date", date, "duration", time.Since(start))
	return nil
}
