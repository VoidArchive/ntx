package worker

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

func newTestWorker(t *testing.T) (*Worker, *sqlc.Queries) {
	t.Helper()
	db, err := database.OpenDB(filepath.Join(t.TempDir(), "market.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	queries := sqlc.New(db)
	return New(nil, queries), queries
}

func TestStoreOwnershipRecordsHistoryOnlyOnChange(t *testing.T) {
	ctx := context.Background()
	w, q := newTestWorker(t)
	if err := q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
		ID: 131, Name: "Nabil Bank", Symbol: "NABIL", Status: "A", Sector: "Commercial Banks", InstrumentType: "Equity",
	}); err != nil {
		t.Fatal(err)
	}

	first := nepse.Ownership{
		ListedShares:    1000,
		PublicShares:    400,
		PublicPercent:   40,
		PromoterShares:  600,
		PromoterPercent: 60,
		AsOf:            "2026-11-02",
	}
	changed := nepse.Ownership{
		ListedShares:    1000,
		PublicShares:    450,
		PublicPercent:   45,
		PromoterShares:  550,
		PromoterPercent: 55,
		AsOf:            "2026-11-04",
	}
	unchanged := changed
	unchanged.AsOf = "2026-11-05"

	steps := []struct {
		name       string
		in         nepse.Ownership
		wantRows   int
		wantLatest string
	}{
		{"first sync records a row", first, 1, "2026-11-02"},
		{"changed values add a row", changed, 2, "2026-11-04"},
		{"unchanged values add nothing", unchanged, 2, "2026-11-04"},
	}
	for _, s := range steps {
		if err := w.storeOwnership(ctx, 131, s.in); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		rows, err := q.ListOwnershipHistory(ctx, 131)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != s.wantRows {
			t.Fatalf("%s: %d history rows, want %d", s.name, len(rows), s.wantRows)
		}
		if got := rows[len(rows)-1].AsOf; got != s.wantLatest {
			t.Fatalf("%s: latest as_of %q, want %q", s.name, got, s.wantLatest)
		}
	}

	latest, err := q.GetOwnership(ctx, 131)
	if err != nil {
		t.Fatal(err)
	}
	if latest.PublicShares.Int64 != 450 || latest.PromoterPercent.Float64 != 55 {
		t.Fatalf("ownership not kept at latest values: %+v", latest)
	}
}
