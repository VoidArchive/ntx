package portfolio

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

// bankFigure is the column a bank metric gate reads from the owner's row.
func bankFigure(m sqlc.BankMetric, metric string) sql.NullFloat64 {
	switch metric {
	case "npl_percent":
		return m.NplPercent
	case "car_percent":
		return m.CarPercent
	case "loan_growth_percent":
		return m.LoanGrowthPercent
	case "distributable_profit":
		return m.DistributableProfit
	}
	return sql.NullFloat64{}
}

// bankSeries is a bank metric by quarter, oldest first. A quarter the exchange has a
// filing for but the owner has no figure for is a missing point, so the gate reads
// Needs data for that quarter rather than judging the last entered one.
func bankSeries(fundamentals []sqlc.Fundamental, bank []sqlc.BankMetric, metric string) []point {
	byRank := map[int]point{}
	for _, m := range bank {
		y, err := fiscal.ParseYear(m.FiscalYear)
		if err != nil {
			continue
		}
		q, err := fiscal.ParseQuarter(m.Quarter)
		if err != nil || q == fiscal.Annual {
			continue
		}
		p := point{Period: q.String() + " " + y.String(), Source: m.SourceUrl, Missing: true}
		if v := bankFigure(m, metric); v.Valid {
			p.Value, p.Missing = v.Float64, false
		}
		byRank[filing{year: y, quarter: q}.rank()] = p
	}
	for _, f := range filings(fundamentals) {
		if _, ok := byRank[f.rank()]; !ok && f.quarter != fiscal.Annual {
			byRank[f.rank()] = point{Period: f.quarter.String() + " " + f.year.String(), Missing: true}
		}
	}
	var out []point
	for _, r := range slices.Sorted(maps.Keys(byRank)) {
		out = append(out, byRank[r])
	}
	return out
}

func (s *Service) UpsertBankMetric(
	ctx context.Context,
	req *connect.Request[ntxv1.UpsertBankMetricRequest],
) (*connect.Response[ntxv1.UpsertBankMetricResponse], error) {
	m := req.Msg.GetMetric()
	c, err := s.queries.GetCompany(ctx, normSymbol(m.GetSymbol()))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no company %q", m.GetSymbol()))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	y, q, err := validBankMetric(m)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	row, err := s.queries.UpsertBankMetric(ctx, sqlc.UpsertBankMetricParams{
		CompanyID: c.ID, FiscalYear: y.String(), Quarter: q.String(),
		NplPercent:          nullable(m.NplPercent),
		CarPercent:          nullable(m.CarPercent),
		LoanGrowthPercent:   nullable(m.LoanGrowthPercent),
		DistributableProfit: nullable(m.DistributableProfit),
		PublishedDate:       m.GetPublishedDate(),
		SourceUrl:           strings.TrimSpace(m.GetSourceUrl()),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.UpsertBankMetricResponse{Metric: bankToProto(c.Symbol, row)}), nil
}

// validBankMetric refuses a row without its full provenance (ADR 0002): a fiscal
// year, a quarter, the filing date, an http(s) source, and at least one figure.
func validBankMetric(m *ntxv1.BankMetric) (fiscal.Year, fiscal.Quarter, error) {
	y, err := fiscal.ParseYear(m.GetFiscalYear())
	if err != nil {
		return 0, 0, err
	}
	q, err := fiscal.ParseQuarter(m.GetQuarter())
	if err != nil || q == fiscal.Annual {
		return 0, 0, fmt.Errorf("a bank row names one quarter, Q1 to Q4, not %q", m.GetQuarter())
	}
	if _, err := time.Parse(time.DateOnly, m.GetPublishedDate()); err != nil {
		return 0, 0, errors.New("a bank row needs the date the filing was published")
	}
	u, err := url.Parse(strings.TrimSpace(m.GetSourceUrl()))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return 0, 0, errors.New("a bank row needs the filing's URL as its source")
	}
	if m.NplPercent == nil && m.CarPercent == nil && m.LoanGrowthPercent == nil && m.DistributableProfit == nil {
		return 0, 0, errors.New("a bank row needs at least one figure")
	}
	return y, q, nil
}

func nullable(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}

func optional(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}

func bankToProto(symbol string, m sqlc.BankMetric) *ntxv1.BankMetric {
	return &ntxv1.BankMetric{
		Symbol: symbol, FiscalYear: m.FiscalYear, Quarter: m.Quarter,
		NplPercent:          optional(m.NplPercent),
		CarPercent:          optional(m.CarPercent),
		LoanGrowthPercent:   optional(m.LoanGrowthPercent),
		DistributableProfit: optional(m.DistributableProfit),
		PublishedDate:       m.PublishedDate,
		SourceUrl:           m.SourceUrl,
	}
}

// bankRows is a holding's owner-entered bank figures for the story, oldest first.
func (s *Service) bankRows(ctx context.Context, h *ntxv1.Holding) ([]*ntxv1.BankMetric, error) {
	if h.GetIsFund() {
		return nil, nil
	}
	c, err := s.queries.GetCompany(ctx, h.GetSymbol())
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListBankMetricsByCompany(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(rows, func(a, b sqlc.BankMetric) int {
		return cmp.Or(cmp.Compare(a.FiscalYear, b.FiscalYear), cmp.Compare(a.Quarter, b.Quarter))
	})
	var out []*ntxv1.BankMetric
	for _, r := range rows {
		out = append(out, bankToProto(c.Symbol, r))
	}
	return out, nil
}
