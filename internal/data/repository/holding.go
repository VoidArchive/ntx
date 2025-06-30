package repository

import (
	"context"
	"fmt"
	"ntx/internal/database"
)

type holdingRepository struct {
	queries *database.Queries
}

func NewHoldingRepository(queries *database.Queries) HoldingRepository {
	return &holdingRepository{
		queries: queries,
	}
}

func (r *holdingRepository) Create(ctx context.Context, req CreateHoldingRequest) (*database.Holdings, error) {
	params := database.CreateHoldingParams{
		PortfolioID:       req.PortfolioID,
		Symbol:           req.Symbol,
		Quantity:         req.Quantity,
		AverageCostPaisa: req.AverageCostPaisa,
		LastPricePaisa:   nullInt64FromPtr(req.LastPricePaisa),
	}

	holding, err := r.queries.CreateHolding(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create holding: %w", err)
	}

	return &holding, nil
}

func (r *holdingRepository) GetByID(ctx context.Context, id int64) (*database.Holdings, error) {
	holding, err := r.queries.GetHolding(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get holding: %w", err)
	}

	return &holding, nil
}

func (r *holdingRepository) GetBySymbol(ctx context.Context, portfolioID int64, symbol string) (*database.Holdings, error) {
	params := database.GetHoldingBySymbolParams{
		PortfolioID: portfolioID,
		Symbol:     symbol,
	}

	holding, err := r.queries.GetHoldingBySymbol(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get holding by symbol: %w", err)
	}

	return &holding, nil
}

func (r *holdingRepository) ListByPortfolio(ctx context.Context, portfolioID int64) ([]database.Holdings, error) {
	holdings, err := r.queries.ListHoldingsByPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to list holdings: %w", err)
	}

	return holdings, nil
}

func (r *holdingRepository) ListWithValue(ctx context.Context, portfolioID int64) ([]database.ListHoldingsWithValueRow, error) {
	holdings, err := r.queries.ListHoldingsWithValue(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to list holdings with value: %w", err)
	}

	return holdings, nil
}

func (r *holdingRepository) Update(ctx context.Context, req UpdateHoldingRequest) (*database.Holdings, error) {
	params := database.UpdateHoldingParams{
		ID:               req.ID,
		Quantity:         req.Quantity,
		AverageCostPaisa: req.AverageCostPaisa,
		LastPricePaisa:   nullInt64FromPtr(req.LastPricePaisa),
	}

	holding, err := r.queries.UpdateHolding(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update holding: %w", err)
	}

	return &holding, nil
}

func (r *holdingRepository) UpdatePrice(ctx context.Context, portfolioID int64, symbol string, pricePaisa int64) error {
	params := database.UpdateHoldingPriceParams{
		PortfolioID:    portfolioID,
		Symbol:        symbol,
		LastPricePaisa: nullInt64FromPtr(&pricePaisa),
	}

	if err := r.queries.UpdateHoldingPrice(ctx, params); err != nil {
		return fmt.Errorf("failed to update holding price: %w", err)
	}

	return nil
}

func (r *holdingRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteHolding(ctx, id); err != nil {
		return fmt.Errorf("failed to delete holding: %w", err)
	}

	return nil
}

func (r *holdingRepository) GetValue(ctx context.Context, id int64) (*database.GetHoldingValueRow, error) {
	value, err := r.queries.GetHoldingValue(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get holding value: %w", err)
	}

	return &value, nil
}