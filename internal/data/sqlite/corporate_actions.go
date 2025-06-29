package sqlite

import (
	"context"
	"database/sql"
	"time"

	"ntx/internal/data/models"
	"ntx/internal/data/repository"
)

// CorporateActionsRepository implements corporate actions data operations for SQLite
type CorporateActionsRepository struct {
	db QueryExecutor
}

// NewCorporateActionsRepository creates a new corporate actions repository
func NewCorporateActionsRepository(db QueryExecutor) *CorporateActionsRepository {
	return &CorporateActionsRepository{db: db}
}

// CreateCorporateAction creates a new corporate action record
func (r *CorporateActionsRepository) CreateCorporateAction(ctx context.Context, action *models.CorporateAction) error {
	if action == nil || !action.IsValid() {
		return repository.NewInvalidDataError("invalid corporate action data", nil)
	}

	query := `
		INSERT INTO corporate_actions 
		(symbol, action_type, announcement_date, record_date, ex_date, ratio, dividend_amount, processed, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if action.CreatedAt.IsZero() {
		action.CreatedAt = now
	}
	if action.UpdatedAt.IsZero() {
		action.UpdatedAt = now
	}

	result, err := r.db.ExecContext(ctx, query,
		action.Symbol,
		string(action.ActionType),
		action.AnnouncementDate,
		action.RecordDate,
		action.ExDate,
		action.Ratio,
		action.DividendAmount.Paisa(),
		action.Processed,
		action.Notes,
		action.CreatedAt,
		action.UpdatedAt,
	)

	if err != nil {
		return repository.NewInternalError("failed to create corporate action", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return repository.NewInternalError("failed to get insert ID", err)
	}

	action.ID = id
	return nil
}

// GetCorporateAction retrieves a corporate action by ID
func (r *CorporateActionsRepository) GetCorporateAction(ctx context.Context, id int64) (*models.CorporateAction, error) {
	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		WHERE id = ?
	`

	action := &models.CorporateAction{}
	var actionType string
	var ratio, notes sql.NullString
	var processedDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&action.ID,
		&action.Symbol,
		&actionType,
		&action.AnnouncementDate,
		&action.RecordDate,
		&action.ExDate,
		&ratio,
		&action.DividendAmount,
		&action.Processed,
		&processedDate,
		&notes,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.NewNotFoundError("corporate action", id)
		}
		return nil, repository.NewInternalError("failed to get corporate action", err)
	}

	action.ActionType = models.CorporateActionType(actionType)
	if ratio.Valid {
		action.Ratio = ratio.String
	}
	if processedDate.Valid {
		action.ProcessedDate = processedDate.Time
	}
	if notes.Valid {
		action.Notes = notes.String
	}

	// Convert integer values back to proper types
	action.DividendAmount = models.Money(action.DividendAmount)

	return action, nil
}

// GetCorporateActionsBySymbol retrieves all corporate actions for a symbol
func (r *CorporateActionsRepository) GetCorporateActionsBySymbol(ctx context.Context, symbol string) ([]models.CorporateAction, error) {
	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		WHERE symbol = ? COLLATE NOCASE
		ORDER BY ex_date DESC, created_at DESC
	`

	return r.queryCorporateActions(ctx, query, symbol)
}

// GetCorporateActionsByType retrieves corporate actions by type
func (r *CorporateActionsRepository) GetCorporateActionsByType(ctx context.Context, actionType models.CorporateActionType) ([]models.CorporateAction, error) {
	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		WHERE action_type = ?
		ORDER BY ex_date DESC, created_at DESC
	`

	return r.queryCorporateActions(ctx, query, string(actionType))
}

// GetPendingCorporateActions retrieves unprocessed corporate actions past ex-date
func (r *CorporateActionsRepository) GetPendingCorporateActions(ctx context.Context) ([]models.CorporateAction, error) {
	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		WHERE processed = FALSE AND ex_date <= ?
		ORDER BY ex_date ASC, created_at ASC
	`

	return r.queryCorporateActions(ctx, query, time.Now())
}

// GetUpcomingCorporateActions retrieves corporate actions with future ex-dates
func (r *CorporateActionsRepository) GetUpcomingCorporateActions(ctx context.Context, days int) ([]models.CorporateAction, error) {
	futureDate := time.Now().AddDate(0, 0, days)
	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		WHERE ex_date BETWEEN ? AND ?
		ORDER BY ex_date ASC, created_at ASC
	`

	return r.queryCorporateActions(ctx, query, time.Now(), futureDate)
}

// GetAllCorporateActions retrieves all corporate actions
func (r *CorporateActionsRepository) GetAllCorporateActions(ctx context.Context) ([]models.CorporateAction, error) {
	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		ORDER BY ex_date DESC, created_at DESC
	`

	return r.queryCorporateActions(ctx, query)
}

// UpdateCorporateAction updates an existing corporate action
func (r *CorporateActionsRepository) UpdateCorporateAction(ctx context.Context, action *models.CorporateAction) error {
	if action == nil || !action.IsValid() {
		return repository.NewInvalidDataError("invalid corporate action data", nil)
	}

	query := `
		UPDATE corporate_actions 
		SET symbol = ?, action_type = ?, announcement_date = ?, record_date = ?, ex_date = ?, 
		    ratio = ?, dividend_amount = ?, processed = ?, processed_date = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	action.UpdatedAt = time.Now()
	var processedDate interface{}
	if !action.ProcessedDate.IsZero() {
		processedDate = action.ProcessedDate
	}

	result, err := r.db.ExecContext(ctx, query,
		action.Symbol,
		string(action.ActionType),
		action.AnnouncementDate,
		action.RecordDate,
		action.ExDate,
		action.Ratio,
		action.DividendAmount.Paisa(),
		action.Processed,
		processedDate,
		action.Notes,
		action.UpdatedAt,
		action.ID,
	)

	if err != nil {
		return repository.NewInternalError("failed to update corporate action", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("corporate action", action.ID)
	}

	return nil
}

// DeleteCorporateAction deletes a corporate action by ID
func (r *CorporateActionsRepository) DeleteCorporateAction(ctx context.Context, id int64) error {
	query := "DELETE FROM corporate_actions WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return repository.NewInternalError("failed to delete corporate action", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("corporate action", id)
	}

	return nil
}

// MarkAsProcessed marks a corporate action as processed
func (r *CorporateActionsRepository) MarkAsProcessed(ctx context.Context, id int64) error {
	query := `
		UPDATE corporate_actions 
		SET processed = TRUE, processed_date = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return repository.NewInternalError("failed to mark corporate action as processed", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("corporate action", id)
	}

	return nil
}

// GetCorporateActionSummaries returns summarized corporate action information
func (r *CorporateActionsRepository) GetCorporateActionSummaries(ctx context.Context, filters models.CorporateActionFilters) ([]models.CorporateActionSummary, error) {
	actions, err := r.getFilteredCorporateActions(ctx, filters)
	if err != nil {
		return nil, err
	}

	summaries := make([]models.CorporateActionSummary, len(actions))
	for i, action := range actions {
		summaries[i] = action.ToSummary()
	}

	return summaries, nil
}

// GetCorporateActionsForPortfolio retrieves corporate actions for portfolio symbols
func (r *CorporateActionsRepository) GetCorporateActionsForPortfolio(ctx context.Context, symbols []string) (map[string][]models.CorporateAction, error) {
	if len(symbols) == 0 {
		return make(map[string][]models.CorporateAction), nil
	}

	// Build query with placeholders for symbols
	placeholders := make([]interface{}, len(symbols))
	queryArgs := ""
	for i, symbol := range symbols {
		if i > 0 {
			queryArgs += ","
		}
		queryArgs += "?"
		placeholders[i] = symbol
	}

	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		WHERE symbol IN (` + queryArgs + `) COLLATE NOCASE
		ORDER BY symbol, ex_date DESC
	`

	actions, err := r.queryCorporateActions(ctx, query, placeholders...)
	if err != nil {
		return nil, err
	}

	// Group by symbol
	result := make(map[string][]models.CorporateAction)
	for _, action := range actions {
		result[action.Symbol] = append(result[action.Symbol], action)
	}

	return result, nil
}

// getFilteredCorporateActions retrieves corporate actions based on filters
func (r *CorporateActionsRepository) getFilteredCorporateActions(ctx context.Context, filters models.CorporateActionFilters) ([]models.CorporateAction, error) {
	query := `
		SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio, 
		       dividend_amount, processed, processed_date, notes, created_at, updated_at
		FROM corporate_actions
		WHERE 1=1
	`
	
	var args []interface{}

	// Apply filters
	if filters.Symbol != "" {
		query += " AND symbol = ? COLLATE NOCASE"
		args = append(args, filters.Symbol)
	}

	if filters.ActionType != "" {
		query += " AND action_type = ?"
		args = append(args, string(filters.ActionType))
	}

	if filters.ProcessedOnly {
		query += " AND processed = TRUE"
	}

	if filters.UnprocessedOnly {
		query += " AND processed = FALSE"
	}

	if filters.RequiresAction {
		query += " AND processed = FALSE AND ex_date <= ?"
		args = append(args, time.Now())
	}

	if !filters.FromDate.IsZero() {
		query += " AND ex_date >= ?"
		args = append(args, filters.FromDate)
	}

	if !filters.ToDate.IsZero() {
		query += " AND ex_date <= ?"
		args = append(args, filters.ToDate)
	}

	query += " ORDER BY ex_date DESC, created_at DESC"

	return r.queryCorporateActions(ctx, query, args...)
}

// queryCorporateActions is a helper function to execute corporate action queries
func (r *CorporateActionsRepository) queryCorporateActions(ctx context.Context, query string, args ...interface{}) ([]models.CorporateAction, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, repository.NewInternalError("failed to query corporate actions", err)
	}
	defer rows.Close()

	var actions []models.CorporateAction
	for rows.Next() {
		action := models.CorporateAction{}
		var actionType string
		var ratio, notes sql.NullString
		var processedDate sql.NullTime

		err := rows.Scan(
			&action.ID,
			&action.Symbol,
			&actionType,
			&action.AnnouncementDate,
			&action.RecordDate,
			&action.ExDate,
			&ratio,
			&action.DividendAmount,
			&action.Processed,
			&processedDate,
			&notes,
			&action.CreatedAt,
			&action.UpdatedAt,
		)

		if err != nil {
			return nil, repository.NewInternalError("failed to scan corporate action row", err)
		}

		action.ActionType = models.CorporateActionType(actionType)
		if ratio.Valid {
			action.Ratio = ratio.String
		}
		if processedDate.Valid {
			action.ProcessedDate = processedDate.Time
		}
		if notes.Valid {
			action.Notes = notes.String
		}

		// Convert integer values back to proper types
		action.DividendAmount = models.Money(action.DividendAmount)

		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating corporate action rows", err)
	}

	return actions, nil
}