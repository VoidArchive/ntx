package story

import (
	"database/sql"
	"slices"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// Hand-made: the local database has no ownership rows yet.
func observation(asOf string, promoter, public float64) sqlc.OwnershipHistory {
	return sqlc.OwnershipHistory{
		CompanyID:       131,
		AsOf:            asOf,
		PromoterPercent: sql.NullFloat64{Float64: promoter, Valid: true},
		PublicPercent:   sql.NullFloat64{Float64: public, Valid: true},
	}
}

func TestNabilOwnership(t *testing.T) {
	tests := []struct {
		name string
		rows []sqlc.OwnershipHistory
		want []string
	}{
		{
			"changed",
			[]sqlc.OwnershipHistory{
				observation("2026-05-01", 60, 40),
				observation("2026-11-01", 51, 49),
			},
			[]string{
				"Promoters hold 51.00% and the public 49.00%.",
				"The promoter share moved from 60.00% to 51.00% on 2026-11-01.",
			},
		},
		{
			"changed long ago",
			[]sqlc.OwnershipHistory{
				observation("2024-01-01", 60, 40),
				observation("2025-06-01", 51, 49),
			},
			[]string{
				"Promoters hold 51.00% and the public 49.00%.",
				"The promoter share moved from 60.00% to 51.00% on 2025-06-01.",
			},
		},
		{
			"single",
			[]sqlc.OwnershipHistory{observation("2026-11-01", 51, 49)},
			[]string{
				"Promoters hold 51.00% and the public 49.00%.",
				"No change since NTX first recorded the split on 2026-11-01.",
			},
		},
		{"none", nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := Build(Input{Company: nabil, Ownership: tc.rows})
			got := texts(Story{Claims: claimsOf(s, "ownership")})
			if !slices.Equal(got, tc.want) {
				t.Fatalf("claims:\n got %q\nwant %q", got, tc.want)
			}
			if len(s.Ownership) != len(tc.rows) {
				t.Errorf("observations = %d, want %d", len(s.Ownership), len(tc.rows))
			}
			if len(got) > 0 && claimsOf(s, "ownership")[0].Period != tc.rows[len(tc.rows)-1].AsOf {
				t.Errorf("period = %q", claimsOf(s, "ownership")[0].Period)
			}
		})
	}
}

func TestNabilFunds(t *testing.T) {
	tests := []struct {
		funds *Funds
		want  []string
	}{
		{
			&Funds{
				Count:        12,
				Total:        3.2e9,
				LargestName:  "NIBL Sahabhagita Fund",
				LargestValue: 319592955,
				AsOf:         "2083 Ashwin",
			},
			[]string{
				"12 mutual funds hold Rs 320.00 crore of its shares; the largest is NIBL Sahabhagita Fund with Rs 31.96 crore.",
			},
		},
		{
			&Funds{
				Count:        1,
				Total:        319592955,
				LargestName:  "NIBL Sahabhagita Fund",
				LargestValue: 319592955,
				AsOf:         "2083 Ashwin",
			},
			[]string{"One mutual fund, NIBL Sahabhagita Fund, holds Rs 31.96 crore of its shares."},
		},
		{&Funds{}, nil},
		{nil, nil},
	}
	for _, tc := range tests {
		s := Build(Input{Company: nabil, Funds: tc.funds})
		got := claimsOf(s, "funds")
		if !slices.Equal(texts(Story{Claims: got}), tc.want) {
			t.Errorf("funds %+v: got %q, want %q", tc.funds, texts(Story{Claims: got}), tc.want)
		}
		if len(got) > 0 && (got[0].Period != "2083 Ashwin" || got[0].Anchor != "funds") {
			t.Errorf("funds claim = %+v", got[0])
		}
	}
}
