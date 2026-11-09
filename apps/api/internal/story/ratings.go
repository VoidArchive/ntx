package story

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// RatingVersion names the rule set; bump it when a check or the letter rule
// changes so the startup fill recomputes every stored letter.
const RatingVersion = 2

// SyncRatings stores every listed company's Rating for the screener, which
// must not build 281 stories per request. The company page still computes live.
func SyncRatings(ctx context.Context, queries *sqlc.Queries) error {
	companies, err := queries.ListCompanies(ctx, sqlc.ListCompaniesParams{Limit: 1000})
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}
	svc := NewService(queries)
	failed := 0
	for _, c := range companies {
		if err := ctx.Err(); err != nil {
			return err
		}
		in, err := svc.load(ctx, c.Symbol)
		if err != nil {
			failed++
			slog.Warn("rating skipped", "symbol", c.Symbol, "err", err)
			continue
		}
		r := Build(in).Rating
		if err := queries.UpsertCompanyRating(ctx, sqlc.UpsertCompanyRatingParams{
			CompanyID: c.ID, Letter: r.Letter,
			Holds: int64(r.Holds), Evaluated: int64(r.Evaluated), Total: int64(r.Total), Version: RatingVersion,
		}); err != nil {
			return fmt.Errorf("store rating %s: %w", c.Symbol, err)
		}
	}
	slog.Info("ratings stored", "companies", len(companies)-failed, "failed", failed)
	return nil
}
