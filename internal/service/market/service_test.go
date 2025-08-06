package market_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/voidarchive/ntx/internal/domain/models"
	"github.com/voidarchive/ntx/internal/service/market"
)

// --- stub quote source -------------------------------------------------------

type stubSource struct {
	quotes []*models.Quote
	err    error
}

func (s stubSource) GetAllQuotes() ([]*models.Quote, error) {
	return s.quotes, s.err
}

// --- tests -------------------------------------------------------------------

func TestService_GetLiveQuotes(t *testing.T) {
	ctx := context.Background()

	wantQuotes := []*models.Quote{
		{Symbol: "NABIL", LTP: 500.0},
		{Symbol: "NLIC", LTP: 1500.5},
	}

	cases := []struct {
		name    string
		source  stubSource
		want    []*models.Quote
		wantErr bool
	}{
		{
			name:   "happy path",
			source: stubSource{quotes: wantQuotes},
			want:   wantQuotes,
		},
		{
			name:    "source returns error",
			source:  stubSource{err: errors.New("boom")},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		// capture
		t.Run(tc.name, func(t *testing.T) {
			svc := market.New(tc.source)

			got, err := svc.GetLiveQuotes(ctx)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("quotes mismatch:\nwant %#v\ngot  %#v", tc.want, got)
			}
		})
	}
}

func TestNewWithShareSansar(t *testing.T) {
	svc := market.NewWithShareSansar()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}
