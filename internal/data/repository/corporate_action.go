package repository

import (
	"context"
	"fmt"
	"ntx/internal/database"
	"time"
)

type corporateActionRepository struct {
	queries *database.Queries
}

func NewCorporateActionRepository(queries *database.Queries) CorporateActionRepository {
	return &corporateActionRepository{
		queries: queries,
	}
}

func (r *corporateActionRepository) Create(ctx context.Context, req CreateCorporateActionRequest) (*database.CorporateActions, error) {
	params := database.CreateCorporateActionParams{
		Symbol:           req.Symbol,
		ActionType:       req.ActionType,
		AnnouncementDate: req.AnnouncementDate,
		RecordDate:       req.RecordDate,
		ExecutionDate:    nullTimeFromPtr(req.ExecutionDate),
		RatioFrom:        nullInt64FromPtr(req.RatioFrom),
		RatioTo:          nullInt64FromPtr(req.RatioTo),
		AmountPaisa:      nullInt64FromPtr(req.AmountPaisa),
		Notes:            nullStringFromPtr(req.Notes),
	}

	action, err := r.queries.CreateCorporateAction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create corporate action: %w", err)
	}

	return &action, nil
}

func (r *corporateActionRepository) GetByID(ctx context.Context, id int64) (*database.CorporateActions, error) {
	action, err := r.queries.GetCorporateAction(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get corporate action: %w", err)
	}

	return &action, nil
}

func (r *corporateActionRepository) List(ctx context.Context) ([]database.CorporateActions, error) {
	actions, err := r.queries.ListCorporateActions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list corporate actions: %w", err)
	}

	return actions, nil
}

func (r *corporateActionRepository) ListBySymbol(ctx context.Context, symbol string) ([]database.CorporateActions, error) {
	actions, err := r.queries.ListCorporateActionsBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to list corporate actions by symbol: %w", err)
	}

	return actions, nil
}

func (r *corporateActionRepository) ListByType(ctx context.Context, actionType string) ([]database.CorporateActions, error) {
	actions, err := r.queries.ListCorporateActionsByType(ctx, actionType)
	if err != nil {
		return nil, fmt.Errorf("failed to list corporate actions by type: %w", err)
	}

	return actions, nil
}

func (r *corporateActionRepository) ListByDateRange(ctx context.Context, req ListCorporateActionsByDateRangeRequest) ([]database.CorporateActions, error) {
	params := database.ListCorporateActionsByDateRangeParams{
		FromRecordDate: req.StartDate,
		ToRecordDate:   req.EndDate,
	}

	actions, err := r.queries.ListCorporateActionsByDateRange(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list corporate actions by date range: %w", err)
	}

	return actions, nil
}

func (r *corporateActionRepository) Update(ctx context.Context, req UpdateCorporateActionRequest) (*database.CorporateActions, error) {
	params := database.UpdateCorporateActionParams{
		ID:               req.ID,
		Symbol:           req.Symbol,
		ActionType:       req.ActionType,
		AnnouncementDate: req.AnnouncementDate,
		RecordDate:       req.RecordDate,
		ExecutionDate:    nullTimeFromPtr(req.ExecutionDate),
		RatioFrom:        nullInt64FromPtr(req.RatioFrom),
		RatioTo:          nullInt64FromPtr(req.RatioTo),
		AmountPaisa:      nullInt64FromPtr(req.AmountPaisa),
		Notes:            nullStringFromPtr(req.Notes),
	}

	action, err := r.queries.UpdateCorporateAction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update corporate action: %w", err)
	}

	return &action, nil
}

func (r *corporateActionRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteCorporateAction(ctx, id); err != nil {
		return fmt.Errorf("failed to delete corporate action: %w", err)
	}

	return nil
}

func (r *corporateActionRepository) GetPending(ctx context.Context) ([]database.CorporateActions, error) {
	actions, err := r.queries.GetPendingCorporateActions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending corporate actions: %w", err)
	}

	return actions, nil
}

func (r *corporateActionRepository) GetBySymbolAndDate(ctx context.Context, symbol string, date time.Time) ([]database.CorporateActions, error) {
	params := database.GetCorporateActionsBySymbolAndDateParams{
		Symbol:     symbol,
		RecordDate: date,
	}

	actions, err := r.queries.GetCorporateActionsBySymbolAndDate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get corporate actions by symbol and date: %w", err)
	}

	return actions, nil
}