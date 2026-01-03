package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/market"
)

var sectorMap = map[string]int64{
	"Commercial Banks":             1,
	"Development Banks":            2,
	"Finance":                      3,
	"Microfinance":                 4,
	"Life Insurance":               5,
	"Non Life Insurance":           6,
	"Hydro Power":                  7,
	"Manufacturing And Processing": 8,
	"Hotels And Tourism":           9,
	"Trading":                      10,
	"Investment":                   11,
	"Mutual Fund":                  12,
	"Others":                       13,
}

func sectorToInt(name string) int64 {
	if id, ok := sectorMap[name]; ok {
		return id
	}
	return 0
}

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

// syncFundamentals fetches fundamentals for all companies and upserts to database.
func (w *Worker) syncFundamentals(ctx context.Context) error {
	start := time.Now()

	companies, err := w.nepse.Companies(ctx)
	if err != nil {
		return err
	}

	latestPrices, err := w.queries.GetPricesForDate(ctx, time.Now().In(market.NPT).Format(market.DateFormat))
	if err != nil {
		slog.Error("failed to get latest prices", "error", err)
		return err
	}

	priceMap := make(map[string]float64)
	for _, p := range latestPrices {
		priceMap[p.Symbol] = p.Close
	}

	count := 0
	for _, company := range companies {
		symbol := company.Symbol
		price, hasPrice := priceMap[symbol]

		if !hasPrice {
			continue
		}

		reports, err := w.nepse.Reports(ctx, symbol)
		if err != nil {
			slog.Error("failed to fetch reports", "symbol", symbol, "error", err)
			continue
		}

		if len(reports) == 0 {
			continue
		}

		latest := reports[0]

		dividendYield := 0.0
		roe := 0.0

		dividends, err := w.nepse.Dividends(ctx, symbol)
		if err == nil && len(dividends) > 0 {
			latestDiv := dividends[0]
			latestDivCash := latestDiv.CashPercent + latestDiv.BonusPercent
			if latestDivCash > 0 && price > 0 {
				dividendYield = (latestDivCash / price) * 100
			}
		}

		if latest.BookValue > 0 && latest.Profit > 0 {
			roe = (latest.Profit / latest.BookValue) * 100
		}

		err = w.queries.UpsertFundamentals(ctx, sqlc.UpsertFundamentalsParams{
			Symbol:            symbol,
			Pe:                sql.NullFloat64{Float64: latest.PE, Valid: latest.PE > 0},
			Pb:                sql.NullFloat64{Float64: price / latest.BookValue, Valid: latest.BookValue > 0},
			Eps:               sql.NullFloat64{Float64: latest.EPS, Valid: latest.EPS != 0},
			BookValue:         sql.NullFloat64{Float64: latest.BookValue, Valid: latest.BookValue > 0},
			MarketCap:         sql.NullFloat64{Float64: company.MarketCap, Valid: company.MarketCap > 0},
			DividendYield:     sql.NullFloat64{Float64: dividendYield, Valid: dividendYield > 0},
			Roe:               sql.NullFloat64{Float64: roe, Valid: roe != 0},
			SharesOutstanding: sql.NullInt64{Int64: company.Shares, Valid: company.Shares > 0},
		})
		if err != nil {
			slog.Error("failed to upsert fundamentals", "symbol", symbol, "error", err)
			continue
		}

		count++
	}

	slog.Info("fundamentals sync complete", "count", count, "duration", time.Since(start))
	return nil
}

func (w *Worker) syncPrices(ctx context.Context) error {
	start := time.Now()
	today := time.Now().In(market.NPT).Format(market.DateFormat)

	prices, err := w.nepse.LivePrices(ctx)
	if err != nil {
		return err
	}

	for _, p := range prices {
		highLow, _ := w.queries.Get52WeekHighLow(ctx, p.Symbol)

		week52High := 0.0
		week52Low := 0.0

		if highLow.Week52High != nil {
			if h, ok := highLow.Week52High.(float64); ok {
				week52High = h
			}
		}
		if highLow.Week52Low != nil {
			if l, ok := highLow.Week52Low.(float64); ok {
				week52Low = l
			}
		}

		err := w.queries.UpsertPrice(ctx, sqlc.UpsertPriceParams{
			Symbol:        p.Symbol,
			Date:          today,
			Open:          p.Open,
			High:          p.High,
			Low:           p.Low,
			Close:         p.LTP,
			PreviousClose: sql.NullFloat64{Float64: p.PreviousClose, Valid: true},
			Volume:        p.Volume,
			Turnover:      sql.NullInt64{Int64: int64(p.Turnover), Valid: true},
			IsComplete:    0,
			Week52High:    sql.NullFloat64{Float64: week52High, Valid: week52High > 0},
			Week52Low:     sql.NullFloat64{Float64: week52Low, Valid: week52Low > 0},
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
