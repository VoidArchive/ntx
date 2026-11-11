package portfolio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// personalVersion names the file's shape; an import of another version is refused.
const personalVersion = 1

// personal is the file ExportPersonal writes and ImportPersonal reads: every
// personal table, rows with their ids so holding_id references survive.
type personal struct {
	Version      int                `json:"version"`
	ExportedAt   string             `json:"exported_at"`
	Holdings     []sqlc.Holding     `json:"holdings"`
	Transactions []sqlc.Transaction `json:"transactions"`
	Notes        []sqlc.HoldingNote `json:"holding_notes"`
	Gates        []sqlc.Gate        `json:"gates"`
	Rules        sqlc.Rule          `json:"rules"`
	BankMetrics  []bankFile         `json:"bank_metrics"`
}

// bankFile travels by symbol, since company ids are the exchange's; a figure not
// recorded is absent, never a zero (ADR 0002).
type bankFile struct {
	ID                  int64    `json:"id"`
	Symbol              string   `json:"symbol"`
	FiscalYear          string   `json:"fiscal_year"`
	Quarter             string   `json:"quarter"`
	NplPercent          *float64 `json:"npl_percent,omitempty"`
	CarPercent          *float64 `json:"car_percent,omitempty"`
	LoanGrowthPercent   *float64 `json:"loan_growth_percent,omitempty"`
	DistributableProfit *float64 `json:"distributable_profit,omitempty"`
	PublishedDate       string   `json:"published_date"`
	SourceURL           string   `json:"source_url"`
}

func (s *Service) ExportPersonal(
	ctx context.Context,
	_ *connect.Request[ntxv1.ExportPersonalRequest],
) (*connect.Response[ntxv1.ExportPersonalResponse], error) {
	f, err := s.exportFile(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.ExportPersonalResponse{Json: b}), nil
}

func (s *Service) exportFile(ctx context.Context) (personal, error) {
	q := s.queries
	f := personal{
		Version: personalVersion, ExportedAt: now().UTC().Format(time.RFC3339),
		Holdings: []sqlc.Holding{}, Transactions: []sqlc.Transaction{}, Notes: []sqlc.HoldingNote{},
		Gates: []sqlc.Gate{}, BankMetrics: []bankFile{},
	}
	var err error
	if f.Holdings, err = q.ListHoldingRows(ctx); err != nil {
		return f, err
	}
	if f.Transactions, err = q.ListTransactions(ctx); err != nil {
		return f, err
	}
	if f.Notes, err = q.ListNotes(ctx); err != nil {
		return f, err
	}
	if f.Gates, err = q.ListGates(ctx); err != nil {
		return f, err
	}
	if f.Rules, err = q.GetRules(ctx); err != nil {
		return f, err
	}
	bank, err := q.ListBankMetrics(ctx)
	if err != nil {
		return f, err
	}
	for _, b := range bank {
		f.BankMetrics = append(f.BankMetrics, bankFile{
			ID: b.ID, Symbol: b.Symbol, FiscalYear: b.FiscalYear, Quarter: b.Quarter,
			NplPercent: optional(b.NplPercent), CarPercent: optional(b.CarPercent),
			LoanGrowthPercent: optional(b.LoanGrowthPercent), DistributableProfit: optional(b.DistributableProfit),
			PublishedDate: b.PublishedDate, SourceURL: b.SourceUrl,
		})
	}
	return f, nil
}

// ImportPersonal replaces every personal table with the file's rows in one
// transaction; a file it cannot take, or a row the schema refuses, changes nothing.
func (s *Service) ImportPersonal(
	ctx context.Context,
	req *connect.Request[ntxv1.ImportPersonalRequest],
) (*connect.Response[ntxv1.ImportPersonalResponse], error) {
	var f personal
	dec := json.NewDecoder(bytes.NewReader(req.Msg.GetJson()))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("not an NTX personal file: %w", err))
	}
	if f.Version != personalVersion {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("file version %d; this NTX writes version %d", f.Version, personalVersion),
		)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := importAll(ctx, s.queries.WithTx(tx), f); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.ImportPersonalResponse{}), nil
}

// importAll wipes and refills inside the caller's transaction. Deleting holdings
// cascades to transactions, notes and gates; a row naming a holding the file lacks
// fails the foreign key and the whole import with it.
func importAll(ctx context.Context, q *sqlc.Queries, f personal) error {
	if err := errors.Join(q.DeleteAllHoldings(ctx), q.DeleteAllBankMetrics(ctx)); err != nil {
		return err
	}
	for _, h := range f.Holdings {
		if err := q.ImportHolding(ctx, sqlc.ImportHoldingParams(h)); err != nil {
			return fmt.Errorf("holding %s: %w", h.Symbol, err)
		}
	}
	for _, t := range f.Transactions {
		if err := q.ImportTransaction(ctx, sqlc.ImportTransactionParams(t)); err != nil {
			return fmt.Errorf("transaction %d: %w", t.ID, err)
		}
	}
	for _, n := range f.Notes {
		if err := q.ImportNote(ctx, sqlc.ImportNoteParams(n)); err != nil {
			return fmt.Errorf("note %d: %w", n.ID, err)
		}
	}
	for _, g := range f.Gates {
		if err := q.ImportGate(ctx, sqlc.ImportGateParams(g)); err != nil {
			return fmt.Errorf("gate %d: %w", g.ID, err)
		}
	}
	for _, b := range f.BankMetrics {
		c, err := q.GetCompany(ctx, normSymbol(b.Symbol))
		if err != nil {
			return fmt.Errorf("bank row for %q: %w", b.Symbol, err)
		}
		if err := q.ImportBankMetric(ctx, sqlc.ImportBankMetricParams{
			ID: b.ID, CompanyID: c.ID, FiscalYear: b.FiscalYear, Quarter: b.Quarter,
			NplPercent: nullable(b.NplPercent), CarPercent: nullable(b.CarPercent),
			LoanGrowthPercent: nullable(b.LoanGrowthPercent), DistributableProfit: nullable(b.DistributableProfit),
			PublishedDate: b.PublishedDate, SourceUrl: b.SourceURL,
		}); err != nil {
			return fmt.Errorf("bank row %d: %w", b.ID, err)
		}
	}
	r := f.Rules
	_, err := q.UpdateRules(ctx, sqlc.UpdateRulesParams{
		SoftCapPercent: r.SoftCapPercent, HardCapPercent: r.HardCapPercent, SectorCapPercent: r.SectorCapPercent,
		MaxPositions: r.MaxPositions, MinEntryRs: r.MinEntryRs, OneHydroPerBasin: r.OneHydroPerBasin,
		CandidateMinRating: r.CandidateMinRating, CandidateDividendRecord: r.CandidateDividendRecord,
		MonthlyBudgetRs: r.MonthlyBudgetRs, MultiplierBelowMedian: r.MultiplierBelowMedian,
		MultiplierBottomQuarter: r.MultiplierBottomQuarter,
	})
	return err
}
