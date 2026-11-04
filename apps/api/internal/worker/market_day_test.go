package worker

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

func TestStoreMarketDayCountsBreadthFromCloses(t *testing.T) {
	ctx := context.Background()
	w, q := newTestWorker(t)

	// Closes on the previous session and on the day under test. Change fields
	// stay NULL, as on the backfill path. ILLIQ skipped 11-03, so its previous
	// close is from 11-02; NEWCO has no previous close and is not counted.
	closes := []struct {
		id          int64
		symbol      string
		prev, today float64
		prevDate    string
	}{
		{1, "UPA", 100, 110, "2026-11-03"},
		{2, "UPB", 200, 201, "2026-11-03"},
		{3, "DOWN", 300, 290, "2026-11-03"},
		{4, "FLAT", 400, 400, "2026-11-03"},
		{5, "ILLIQ", 500, 450, "2026-11-02"},
		{6, "NEWCO", 0, 50, ""},
	}
	for _, c := range closes {
		if err := q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
			ID: c.id, Name: c.symbol, Symbol: c.symbol, Status: "A", Sector: "Others", InstrumentType: "Equity",
		}); err != nil {
			t.Fatal(err)
		}
		if c.prevDate != "" {
			insertClose(t, q, c.id, c.prevDate, c.prev)
		}
		insertClose(t, q, c.id, "2026-11-04", c.today)
	}

	feed := nepse.MarketDay{IndexClose: 2650.5, IndexChange: -12.3, Turnover: 4.2e9, ScripsTraded: 6}

	if err := w.storeMarketDay(ctx, "2026-11-05", feed); err != nil {
		t.Fatal(err)
	}
	if _, err := q.GetLatestMarketDay(ctx); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("a day with no price rows wrote a market day (err %v)", err)
	}

	if err := w.storeMarketDay(ctx, "2026-11-04", feed); err != nil {
		t.Fatal(err)
	}
	got, err := q.GetLatestMarketDay(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.BusinessDate != "2026-11-04" || got.Advances != 2 || got.Declines != 2 || got.Unchanged != 1 {
		t.Fatalf("got %s advances %d declines %d unchanged %d, want 2026-11-04 2/2/1",
			got.BusinessDate, got.Advances, got.Declines, got.Unchanged)
	}
	if got.IndexClose != 2650.5 || got.IndexChange != -12.3 || got.Turnover != 4.2e9 {
		t.Fatalf("feed figures not stored: %+v", got)
	}
}

func insertClose(t *testing.T, q *sqlc.Queries, companyID int64, date string, close float64) {
	t.Helper()
	if err := q.UpsertPrice(context.Background(), sqlc.UpsertPriceParams{
		CompanyID: companyID, BusinessDate: date, ClosePrice: nullFloat64(close),
	}); err != nil {
		t.Fatal(err)
	}
}
