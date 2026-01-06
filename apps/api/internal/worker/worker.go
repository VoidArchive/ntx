// Package worker
package worker

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

type Worker struct {
	nepse   *nepse.Client
	queries *sqlc.Queries
}

func New(client *nepse.Client, queries *sqlc.Queries) *Worker {
	return &Worker{
		nepse:   client,
		queries: queries,
	}
}

func (w *Worker) SyncCompanies(ctx context.Context) error {
	companies, err := w.nepse.Companies(ctx)
	if err != nil {
		return fmt.Errorf("nepse companies: %w", err)
	}

	for _, c := range companies {
		params := sqlc.UpsertCompanyParams{
			ID:             c.ID,
			Name:           c.Name,
			Symbol:         c.Symbol,
			Status:         c.Status,
			Email:          nullString(c.Email),
			Website:        nullString(c.Website),
			Sector:         c.Sector,
			InstrumentType: c.InstrumentType,
		}
		if err := w.queries.UpsertCompany(ctx, params); err != nil {
			return fmt.Errorf("upsert company %q: %w", c.Symbol, err)
		}
	}
	return nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
