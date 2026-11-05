package worker

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

func TestStoreCompaniesRenamesByIDAndDelistsMissing(t *testing.T) {
	ctx := context.Background()
	w, q := newTestWorker(t)

	day1 := []nepse.Company{
		{
			ID:             131,
			Name:           "Nabil Bank",
			Symbol:         "NABIL",
			Status:         "A",
			Sector:         "Commercial Banks",
			InstrumentType: "Equity",
		},
		{
			ID:             502,
			Name:           "Kumari Bank",
			Symbol:         "KBL",
			Status:         "A",
			Sector:         "Commercial Banks",
			InstrumentType: "Equity",
		},
	}
	if err := w.storeCompanies(ctx, day1); err != nil {
		t.Fatal(err)
	}

	// After a merger the exchange keeps the id and changes the ticker; KBL stops arriving.
	day2 := []nepse.Company{
		{
			ID:             131,
			Name:           "Nabil Bank",
			Symbol:         "NABL",
			Status:         "A",
			Sector:         "Commercial Banks",
			InstrumentType: "Equity",
		},
	}
	if err := w.storeCompanies(ctx, day2); err != nil {
		t.Fatalf("rename with the same id: %v", err)
	}

	renamed, err := q.GetCompany(ctx, "NABL")
	if err != nil {
		t.Fatalf("new symbol not found: %v", err)
	}
	if renamed.ID != 131 {
		t.Fatalf("new symbol row id = %d, want 131", renamed.ID)
	}
	if _, err := q.GetCompany(ctx, "NABIL"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("old symbol still resolves (err %v)", err)
	}

	gone, err := q.GetCompany(ctx, "KBL")
	if err != nil {
		t.Fatalf("delisted company must keep its row: %v", err)
	}
	if gone.Status != "D" {
		t.Fatalf("missing company status = %q, want D", gone.Status)
	}

	listed, err := q.ListCompanies(ctx, sqlc.ListCompaniesParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].Symbol != "NABL" {
		t.Fatalf("listed = %+v, want only NABL", listed)
	}

	if err := w.storeCompanies(ctx, nil); err == nil {
		t.Fatal("an empty feed must not delist every company")
	}
}
