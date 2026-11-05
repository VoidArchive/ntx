package company

import (
	"context"
	"database/sql"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func TestGetFundamentalsCarriesProvenance(t *testing.T) {
	ctx := context.Background()
	q := database.OpenTestDB(t)

	if err := q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
		ID: 131, Name: "Nabil Bank Limited", Symbol: "NABIL", Status: "A",
		Sector: "Commercial Banks", InstrumentType: "Equity",
	}); err != nil {
		t.Fatal(err)
	}

	const url = "https://www.nepalstock.com.np/api/nots/security/fetchFiles?fileLocation=nabil_user%2F2026-11-02%2FQ1.pdf"
	// First sync predates provenance; the second must fill it in place (backfill by re-running the sync).
	rows := []sqlc.UpsertFundamentalParams{
		{CompanyID: 131, FiscalYear: "2026-2027", Quarter: "First Quarter"},
		{
			CompanyID: 131, FiscalYear: "2026-2027", Quarter: "First Quarter",
			PublishedAt:  sql.NullString{String: "2026-11-02", Valid: true},
			SourceUrl:    sql.NullString{String: url, Valid: true},
			FiscalYearNp: sql.NullString{String: "2083-2084", Valid: true},
		},
		{CompanyID: 131, FiscalYear: "2025-2026", Quarter: "Fourth Quarter"},
	}
	for _, r := range rows {
		if err := q.UpsertFundamental(ctx, r); err != nil {
			t.Fatal(err)
		}
	}

	req := connect.NewRequest(&ntxv1.GetFundamentalsRequest{Symbol: "NABIL"})
	res, err := NewCompanyService(q).GetFundamentals(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	history := res.Msg.GetHistory()
	if len(history) != 2 {
		t.Fatalf("history has %d rows, want 2", len(history))
	}

	tests := []struct {
		name                         string
		got                          *ntxv1.Fundamental
		publishedAt, sourceURL, fyNp *string
	}{
		{"recorded", history[0], new("2026-11-02"), new(url), new("2083-2084")},
		{"not recorded", history[1], nil, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOptional(t, "published_at", tt.got.PublishedAt, tt.publishedAt)
			assertOptional(t, "source_url", tt.got.SourceUrl, tt.sourceURL)
			assertOptional(t, "fiscal_year_np", tt.got.FiscalYearNp, tt.fyNp)
		})
	}
}

func assertOptional(t *testing.T, field string, got, want *string) {
	t.Helper()
	if (got == nil) != (want == nil) || (got != nil && *got != *want) {
		t.Errorf("%s = %v, want %v", field, deref(got), deref(want))
	}
}

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func TestFundamentalPeriod(t *testing.T) {
	tests := []struct {
		np, gregorian, quarter, want string
	}{
		{"2083-2084", "2026-2027", "First Quarter", "Q1 2083/84"},
		{"", "2026-2027", "Fourth Quarter", "Q4 2083/84"},
		{"", "2025-2026", "", "FY 2082/83"},
		{"", "unknown", "First Quarter", ""},
	}
	for _, tc := range tests {
		f := sqlc.Fundamental{
			FiscalYear:   tc.gregorian,
			Quarter:      tc.quarter,
			FiscalYearNp: sql.NullString{String: tc.np, Valid: tc.np != ""},
		}
		if got := fundamentalToProto(f).GetPeriod(); got != tc.want {
			t.Fatalf("period(%q, %q, %q) = %q, want %q", tc.np, tc.gregorian, tc.quarter, got, tc.want)
		}
	}
}

func TestListScreener(t *testing.T) {
	ctx := context.Background()
	q := database.OpenTestDB(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []sqlc.UpsertCompanyParams{
		{ID: 131, Name: "Nabil Bank Limited", Symbol: "NABIL", Status: "A", Sector: "Commercial Banks", InstrumentType: "Equity"},
		{ID: 200, Name: "Unfiled Co", Symbol: "UNF", Status: "A", Sector: "Hotels And Tourism", InstrumentType: "Equity"},
	} {
		must(q.UpsertCompany(ctx, c))
	}
	must(q.UpsertPrice(ctx, sqlc.UpsertPriceParams{
		CompanyID: 131, BusinessDate: "2026-11-03",
		ClosePrice: sql.NullFloat64{Float64: 552, Valid: true},
		Turnover:   sql.NullFloat64{Float64: 1e8, Valid: true},
	}))
	must(q.UpsertFundamental(ctx, sqlc.UpsertFundamentalParams{
		CompanyID: 131, FiscalYear: "2026-2027", Quarter: "First Quarter",
		Eps:       sql.NullFloat64{Float64: 34.5, Valid: true},
		BookValue: sql.NullFloat64{Float64: 240, Valid: true},
	}))
	must(q.UpsertOwnership(ctx, sqlc.UpsertOwnershipParams{
		CompanyID: 131, ListedShares: sql.NullInt64{Int64: 270_000_000, Valid: true},
	}))
	must(
		q.UpsertCompanyRating(ctx, sqlc.UpsertCompanyRatingParams{CompanyID: 131, Letter: "B", Holds: 5, Evaluated: 7}),
	)

	res, err := NewCompanyService(q).ListScreener(ctx, connect.NewRequest(&ntxv1.ListScreenerRequest{}))
	must(err)

	// A filed company carries every ratio but yield (no notice); an unfiled one only its identity.
	want := []*ntxv1.ScreenerRow{
		{
			Symbol: "NABIL", Name: "Nabil Bank Limited", Sector: ntxv1.Sector_SECTOR_COMMERCIAL_BANK, Rating: "B",
			Close: proto.Float64(552), Pe: proto.Float64(16), Pb: proto.Float64(2.3), Eps: proto.Float64(34.5),
			MarketCap: proto.Float64(552 * 270_000_000), Turnover: proto.Float64(1e8),
			Period: "Q1 2083/84", CloseDate: "2026-11-03",
		},
		{Symbol: "UNF", Name: "Unfiled Co", Sector: ntxv1.Sector_SECTOR_HOTEL},
	}
	rows := res.Msg.GetRows()
	if len(rows) != len(want) {
		t.Fatalf("rows = %d, want %d", len(rows), len(want))
	}
	for i := range want {
		if !proto.Equal(rows[i], want[i]) {
			t.Errorf("row %d = %v, want %v", i, rows[i], want[i])
		}
	}
}
