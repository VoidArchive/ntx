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
type personalFile struct {
	Version      int                `json:"version"`
	ExportedAt   string             `json:"exported_at"`
	Holdings     []sqlc.Holding     `json:"holdings"`
	Transactions []sqlc.Transaction `json:"transactions"`
	Notes        []sqlc.HoldingNote `json:"holding_notes"`
	Gates        []sqlc.Gate        `json:"gates"`
	Rules        *sqlc.Rule         `json:"rules"`
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

func (s *Service) exportFile(ctx context.Context) (personalFile, error) {
	q := s.queries
	f := personalFile{
		Version: personalVersion, ExportedAt: now().UTC().Format(time.RFC3339),
		Holdings: []sqlc.Holding{}, Transactions: []sqlc.Transaction{}, Notes: []sqlc.HoldingNote{},
		Gates: []sqlc.Gate{}, BankMetrics: []bankFile{},
	}
	var err error
	if f.Holdings, err = q.ExportHoldings(ctx); err != nil {
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
	// An empty table is "[]", never null: the import refuses a null table as a truncated file.
	f.Holdings, f.Transactions = orEmpty(f.Holdings), orEmpty(f.Transactions)
	f.Notes, f.Gates = orEmpty(f.Notes), orEmpty(f.Gates)
	rules, err := q.GetRules(ctx)
	if err != nil {
		return f, err
	}
	f.Rules = &rules
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

func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// ImportPersonal replaces every personal table with the file's rows in one
// transaction; a file it cannot take, or a row the schema refuses, changes nothing.
func (s *Service) ImportPersonal(
	ctx context.Context,
	req *connect.Request[ntxv1.ImportPersonalRequest],
) (*connect.Response[ntxv1.ImportPersonalResponse], error) {
	// The rules decode over the row as it stands, so a file written before a rule
	// existed keeps that rule's current value rather than zeroing it.
	current, err := s.queries.GetRules(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	f := personalFile{Rules: &current}
	dec := json.NewDecoder(bytes.NewReader(req.Msg.GetJson()))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("not an NTX personal file: %w", err))
	}
	if err := f.complete(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck // a rollback after commit is a no-op
	if err := importAll(ctx, s.queries.WithTx(tx), f); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.ImportPersonalResponse{}), nil
}

// complete refuses a file missing any table, so a truncated or hand-edited file
// cannot wipe the book by omission, and a bank row or rules row the forms would refuse.
func (f personalFile) complete() error {
	if f.Version != personalVersion {
		return fmt.Errorf("file version %d; this NTX writes version %d", f.Version, personalVersion)
	}
	for name, missing := range map[string]bool{
		"holdings": f.Holdings == nil, "transactions": f.Transactions == nil, "holding_notes": f.Notes == nil,
		"gates": f.Gates == nil, "rules": f.Rules == nil, "bank_metrics": f.BankMetrics == nil,
	} {
		if missing {
			return fmt.Errorf("the file has no %q", name)
		}
	}
	if err := validRules(*f.Rules); err != nil {
		return err
	}
	for _, b := range f.BankMetrics {
		if _, _, err := validBankMetric(b.proto()); err != nil {
			return fmt.Errorf("bank row %d: %w", b.ID, err)
		}
	}
	return nil
}

// proto is the row as the form sends it, so the import runs the same checks.
func (b bankFile) proto() *ntxv1.BankMetric {
	return &ntxv1.BankMetric{
		Symbol: b.Symbol, FiscalYear: b.FiscalYear, Quarter: b.Quarter,
		NplPercent: b.NplPercent, CarPercent: b.CarPercent,
		LoanGrowthPercent: b.LoanGrowthPercent, DistributableProfit: b.DistributableProfit,
		PublishedDate: b.PublishedDate, SourceUrl: b.SourceURL,
	}
}

// importAll wipes and refills inside the caller's transaction. Deleting holdings
// cascades to transactions, notes and gates; a row naming a holding the file lacks
// fails the foreign key and the whole import with it.
func importAll(ctx context.Context, q *sqlc.Queries, f personalFile) error {
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
	_, err := q.UpdateRules(ctx, rulesParams(*f.Rules))
	return err
}
