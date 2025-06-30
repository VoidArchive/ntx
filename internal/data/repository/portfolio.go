package repository

import (
	"context"
	"fmt"
	"ntx/internal/database"
)

type portfolioRepository struct {
	queries *database.Queries
}

func NewPortfolioRepository(queries *database.Queries) PortfolioRepository {
	return &portfolioRepository{
		queries: queries,
	}
}

func (r *portfolioRepository) Create(ctx context.Context, req CreatePortfolioRequest) (*database.Portfolios, error) {
	params := database.CreatePortfolioParams{
		Name:        req.Name,
		Description: nullStringFromPtr(req.Description),
		Currency:    nullStringFromPtr(&req.Currency),
	}

	portfolio, err := r.queries.CreatePortfolio(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	return &portfolio, nil
}

func (r *portfolioRepository) GetByID(ctx context.Context, id int64) (*database.Portfolios, error) {
	portfolio, err := r.queries.GetPortfolio(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	return &portfolio, nil
}

func (r *portfolioRepository) List(ctx context.Context) ([]database.Portfolios, error) {
	portfolios, err := r.queries.ListPortfolios(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list portfolios: %w", err)
	}

	return portfolios, nil
}

func (r *portfolioRepository) Update(ctx context.Context, req UpdatePortfolioRequest) (*database.Portfolios, error) {
	params := database.UpdatePortfolioParams{
		ID:          req.ID,
		Name:        req.Name,
		Description: nullStringFromPtr(req.Description),
	}

	portfolio, err := r.queries.UpdatePortfolio(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update portfolio: %w", err)
	}

	return &portfolio, nil
}

func (r *portfolioRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeletePortfolio(ctx, id); err != nil {
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}

	return nil
}

func (r *portfolioRepository) GetStats(ctx context.Context, id int64) (*database.GetPortfolioStatsRow, error) {
	stats, err := r.queries.GetPortfolioStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio stats: %w", err)
	}

	return &stats, nil
}