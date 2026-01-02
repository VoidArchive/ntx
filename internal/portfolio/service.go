package portfolio

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/meroshare"
	"github.com/voidarchive/ntx/internal/nepse"
)

type ImportResult struct {
	Imported int
	Skipped  int
	Errors   []string
}

type Service struct {
	queries *sqlc.Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		queries: sqlc.New(db),
	}
}

func (s *Service) ImportCSV(ctx context.Context, csvData []byte) (*ImportResult, error) {
	if len(csvData) == 0 {
		return nil, fmt.Errorf("empty CSV data")
	}

	r := csv.NewReader(strings.NewReader(string(csvData)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("no data in CSV")
	}

	result := &ImportResult{
		Imported: 0,
		Skipped:  0,
		Errors:   []string{},
	}

	holdingsMap := make(map[string]struct{})

	for i := 1; i < len(records); i++ {
		rec := records[i]
		tx, err := parseMeroshareRecord(rec)
		if err != nil {
			slog.Debug("failed to parse record", "row", i, "error", err)
			result.Skipped++
			continue
		}

		// Check for duplicate transaction
		exists, err := s.queries.TransactionExists(ctx, sqlc.TransactionExistsParams{
			Symbol:      tx.Symbol,
			Date:        tx.Date,
			Description: tx.Description,
		})
		if err != nil {
			slog.Warn("failed to check duplicate", "row", i, "error", err)
		}
		if exists == 1 {
			slog.Debug("skipping duplicate transaction", "symbol", tx.Symbol, "date", tx.Date)
			result.Skipped++
			continue
		}

		if err := s.queries.UpsertStock(ctx, sqlc.UpsertStockParams{
			Symbol: tx.Symbol,
			Name:   "",
			Sector: 0,
		}); err != nil {
			slog.Warn("failed to upsert stock", "symbol", tx.Symbol, "error", err)
		}

		if err := s.queries.CreateTransaction(ctx, tx); err != nil {
			slog.Warn("failed to create transaction", "row", i, "error", err)
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", i, err))
			continue
		}

		holdingsMap[tx.Symbol] = struct{}{}
		result.Imported++
	}

	if err := s.recalculateHoldings(ctx, holdingsMap); err != nil {
		slog.Error("failed to recalculate holdings", "error", err)
		return result, fmt.Errorf("failed to recalculate holdings: %w", err)
	}

	return result, nil
}

func (s *Service) recalculateHoldings(ctx context.Context, symbols map[string]struct{}) error {
	for symbol := range symbols {
		txs, err := s.queries.ListTransactionsBySymbol(ctx, symbol)
		if err != nil {
			return fmt.Errorf("failed to list transactions for %s: %w", symbol, err)
		}

		var quantity float64
		var totalCostPaisa int64

		for _, tx := range txs {
			switch tx.Type {
			case int64(v1.TransactionType_TRANSACTION_TYPE_BUY),
				int64(v1.TransactionType_TRANSACTION_TYPE_IPO),
				int64(v1.TransactionType_TRANSACTION_TYPE_RIGHTS),
				int64(v1.TransactionType_TRANSACTION_TYPE_BONUS),
				int64(v1.TransactionType_TRANSACTION_TYPE_MERGER_IN),
				int64(v1.TransactionType_TRANSACTION_TYPE_REARRANGEMENT):
				quantity += tx.Quantity
				totalCostPaisa += tx.TotalPaisa
			case int64(v1.TransactionType_TRANSACTION_TYPE_SELL),
				int64(v1.TransactionType_TRANSACTION_TYPE_MERGER_OUT):
				// Quantity is stored as positive for sells, subtract here
				quantity -= tx.Quantity
			case int64(v1.TransactionType_TRANSACTION_TYPE_DEMAT):
				// Demat doesn't affect quantity
			}
		}

		if quantity > 0 {
			avgCostPaisa := int64(float64(totalCostPaisa) / quantity)
			if err := s.queries.UpsertHolding(ctx, sqlc.UpsertHoldingParams{
				Symbol:           symbol,
				Quantity:         quantity,
				AverageCostPaisa: avgCostPaisa,
				TotalCostPaisa:   totalCostPaisa,
			}); err != nil {
				return fmt.Errorf("failed to upsert holding for %s: %w", symbol, err)
			}
		} else {
			if err := s.queries.DeleteHolding(ctx, symbol); err != nil {
				slog.Warn("failed to delete zero-quantity holding", "symbol", symbol, "error", err)
			}
		}
	}

	return nil
}

func parseMeroshareRecord(record []string) (sqlc.CreateTransactionParams, error) {
	// Meroshare CSV: S.N, Scrip, Transaction Date, Credit Quantity, Debit Quantity, Balance, History Description
	if len(record) < 7 {
		return sqlc.CreateTransactionParams{}, fmt.Errorf("invalid record: expected 7 fields, got %d", len(record))
	}

	scrip := strings.TrimSpace(record[1])
	if scrip == "" {
		return sqlc.CreateTransactionParams{}, fmt.Errorf("empty scrip")
	}

	dateStr := strings.TrimSpace(record[2])
	if dateStr == "" {
		return sqlc.CreateTransactionParams{}, fmt.Errorf("empty date")
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return sqlc.CreateTransactionParams{}, fmt.Errorf("invalid date format: %w", err)
	}

	creditQty := meroshare.ParseQuantity(record[3])
	debitQty := meroshare.ParseQuantity(record[4])
	description := strings.TrimSpace(record[6])

	txType := detectTransactionType(description, creditQty, debitQty)

	// Determine quantity based on transaction type
	// All quantities stored as positive; sign is determined by type
	var quantity float64
	switch txType {
	case int64(v1.TransactionType_TRANSACTION_TYPE_BUY),
		int64(v1.TransactionType_TRANSACTION_TYPE_IPO),
		int64(v1.TransactionType_TRANSACTION_TYPE_RIGHTS),
		int64(v1.TransactionType_TRANSACTION_TYPE_BONUS),
		int64(v1.TransactionType_TRANSACTION_TYPE_MERGER_IN),
		int64(v1.TransactionType_TRANSACTION_TYPE_REARRANGEMENT):
		quantity = creditQty
	case int64(v1.TransactionType_TRANSACTION_TYPE_SELL),
		int64(v1.TransactionType_TRANSACTION_TYPE_MERGER_OUT):
		quantity = debitQty // Store as positive
	case int64(v1.TransactionType_TRANSACTION_TYPE_DEMAT):
		quantity = 0
	default:
		// Unknown type - use net quantity
		if creditQty > 0 {
			quantity = creditQty
		} else {
			quantity = debitQty
		}
	}

	if txType != int64(v1.TransactionType_TRANSACTION_TYPE_DEMAT) && quantity == 0 {
		return sqlc.CreateTransactionParams{}, fmt.Errorf("zero quantity transaction")
	}

	txID, err := uuid.NewV7()
	if err != nil {
		return sqlc.CreateTransactionParams{}, fmt.Errorf("failed to generate UUID: %w", err)
	}

	return sqlc.CreateTransactionParams{
		ID:          txID.String(),
		Symbol:      scrip,
		Type:        txType,
		Quantity:    quantity,
		PricePaisa:  0,
		TotalPaisa:  0,
		Date:        dateStr,
		Description: sql.NullString{String: description, Valid: description != ""},
	}, nil
}

func detectTransactionType(desc string, creditQty, debitQty float64) int64 {
	desc = strings.TrimSpace(strings.ToUpper(desc))

	switch {
	case strings.Contains(desc, "INITIAL PUBLIC OFFERING"):
		return int64(v1.TransactionType_TRANSACTION_TYPE_IPO)
	case strings.HasPrefix(desc, "CA-BONUS"):
		return int64(v1.TransactionType_TRANSACTION_TYPE_BONUS)
	case strings.HasPrefix(desc, "CA-MERGER"):
		if debitQty > 0 {
			return int64(v1.TransactionType_TRANSACTION_TYPE_MERGER_OUT)
		}
		return int64(v1.TransactionType_TRANSACTION_TYPE_MERGER_IN)
	case strings.HasPrefix(desc, "CA-RIGHTS"):
		return int64(v1.TransactionType_TRANSACTION_TYPE_RIGHTS)
	case strings.HasPrefix(desc, "CA-REARRANGEMENT"):
		return int64(v1.TransactionType_TRANSACTION_TYPE_REARRANGEMENT)
	case strings.HasPrefix(desc, "ON-CR"):
		return int64(v1.TransactionType_TRANSACTION_TYPE_BUY)
	case strings.HasPrefix(desc, "ON-DR"):
		return int64(v1.TransactionType_TRANSACTION_TYPE_SELL)
	case strings.HasPrefix(desc, "DEM"):
		return int64(v1.TransactionType_TRANSACTION_TYPE_DEMAT)
	default:
		return int64(v1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED)
	}
}

func (s *Service) ListHoldings(ctx context.Context) ([]*v1.Holding, error) {
	dbHoldings, err := s.queries.ListHoldings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list holdings: %w", err)
	}

	holdings := make([]*v1.Holding, 0, len(dbHoldings))
	for _, h := range dbHoldings {
		stock := &v1.Stock{
			Symbol: h.Symbol,
		}

		holding := &v1.Holding{
			Stock:                stock,
			Quantity:             h.Quantity,
			AverageCost:          moneyFromPaisa(h.AverageCostPaisa),
			TotalCost:            moneyFromPaisa(h.TotalCostPaisa),
			CurrentPrice:         moneyFromPaisaOpt(h.CurrentPricePaisa),
			CurrentValue:         moneyFromPaisaOpt(h.CurrentValuePaisa),
			UnrealizedPnl:        moneyFromPaisaOpt(h.UnrealizedPnlPaisa),
			UnrealizedPnlPercent: h.UnrealizedPnlPercent.Float64,
		}
		holdings = append(holdings, holding)
	}

	return holdings, nil
}

func (s *Service) GetHolding(ctx context.Context, symbol string) (*v1.Holding, error) {
	h, err := s.queries.GetHolding(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get holding %s: %w", symbol, err)
	}

	return &v1.Holding{
		Stock:                &v1.Stock{Symbol: h.Symbol},
		Quantity:             h.Quantity,
		AverageCost:          moneyFromPaisa(h.AverageCostPaisa),
		TotalCost:            moneyFromPaisa(h.TotalCostPaisa),
		CurrentPrice:         moneyFromPaisaOpt(h.CurrentPricePaisa),
		CurrentValue:         moneyFromPaisaOpt(h.CurrentValuePaisa),
		UnrealizedPnl:        moneyFromPaisaOpt(h.UnrealizedPnlPaisa),
		UnrealizedPnlPercent: h.UnrealizedPnlPercent.Float64,
	}, nil
}

func (s *Service) Summary(ctx context.Context) (*v1.PortfolioSummary, error) {
	holdings, err := s.queries.ListHoldings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list holdings: %w", err)
	}

	var totalInvestmentPaisa int64
	var currentValuePaisa int64
	var totalUnrealizedPnlPaisa int64

	for _, h := range holdings {
		totalInvestmentPaisa += h.TotalCostPaisa
		if h.CurrentValuePaisa.Valid {
			currentValuePaisa += h.CurrentValuePaisa.Int64
		}
		if h.UnrealizedPnlPaisa.Valid {
			totalUnrealizedPnlPaisa += h.UnrealizedPnlPaisa.Int64
		}
	}

	var totalUnrealizedPnlPercent float64
	if totalInvestmentPaisa > 0 {
		totalUnrealizedPnlPercent = float64(totalUnrealizedPnlPaisa) / float64(totalInvestmentPaisa) * 100
	}

	summary := &v1.PortfolioSummary{
		TotalInvestment:           moneyFromPaisa(totalInvestmentPaisa),
		CurrentValue:              moneyFromPaisa(currentValuePaisa),
		TotalUnrealizedPnl:        moneyFromPaisa(totalUnrealizedPnlPaisa),
		TotalUnrealizedPnlPercent: totalUnrealizedPnlPercent,
		HoldingsCount:             int32(len(holdings)),
		LastUpdated:               timestamppb.Now(),
	}

	return summary, nil
}

func (s *Service) ListTransactions(
	ctx context.Context,
	symbol string,
	txType v1.TransactionType,
	limit int32,
	offset int32,
) ([]*v1.Transaction, int32, error) {
	params := sqlc.ListTransactionsFilteredParams{
		Column1: symbol,
		Symbol:  symbol,
		Column3: int64(txType),
		Type:    int64(txType),
		Limit:   int64(limit),
		Offset:  int64(offset),
	}

	txs, err := s.queries.ListTransactionsFiltered(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	count, err := s.queries.CountTransactionsFiltered(ctx, sqlc.CountTransactionsFilteredParams{
		Column1: symbol,
		Symbol:  symbol,
		Column3: int64(txType),
		Type:    int64(txType),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	transactions := make([]*v1.Transaction, 0, len(txs))
	for _, tx := range txs {
		date, err := time.Parse("2006-01-02", tx.Date)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse transaction date %s: %w", tx.ID, err)
		}

		transaction := &v1.Transaction{
			Id:          tx.ID,
			Symbol:      tx.Symbol,
			Type:        v1.TransactionType(tx.Type),
			Quantity:    tx.Quantity,
			Price:       moneyFromPaisa(tx.PricePaisa),
			Total:       moneyFromPaisa(tx.TotalPaisa),
			Date:        timestamppb.New(date),
			Description: tx.Description.String,
		}
		transactions = append(transactions, transaction)
	}

	return transactions, int32(count), nil
}

func moneyFromPaisa(paisa int64) *v1.Money {
	return &v1.Money{
		Paisa: paisa,
	}
}

func moneyFromPaisaOpt(paisa sql.NullInt64) *v1.Money {
	if !paisa.Valid {
		return nil
	}
	return &v1.Money{
		Paisa: paisa.Int64,
	}
}

// SyncResult contains the result of syncing prices.
type SyncResult struct {
	Updated int
	Failed  int
	Errors  []string
}

// WACCImportResult contains the result of importing WACC data.
type WACCImportResult struct {
	Updated int
	Skipped int
}

// ImportWACC updates holdings with cost data from Meroshare's WACC Report.
func (s *Service) ImportWACC(ctx context.Context, csvData []byte) (*WACCImportResult, error) {
	wacs, err := meroshare.ParseWACCReport(strings.NewReader(string(csvData)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse WACC report: %w", err)
	}

	result := &WACCImportResult{}

	for _, wac := range wacs {
		// Skip entries with zero quantity (sold holdings)
		if wac.Quantity <= 0 {
			result.Skipped++
			continue
		}

		// Check if holding exists
		holding, err := s.queries.GetHolding(ctx, wac.Symbol)
		if err != nil {
			// Holding doesn't exist, skip
			result.Skipped++
			continue
		}

		// Convert to paisa (1 NPR = 100 paisa)
		avgCostPaisa := int64(wac.Rate * 100)
		totalCostPaisa := int64(wac.TotalCost * 100)

		if err := s.queries.UpdateHoldingCosts(ctx, sqlc.UpdateHoldingCostsParams{
			AverageCostPaisa: avgCostPaisa,
			TotalCostPaisa:   totalCostPaisa,
			Symbol:           wac.Symbol,
		}); err != nil {
			slog.Warn("failed to update holding costs", "symbol", wac.Symbol, "error", err)
			continue
		}

		// Recalculate P&L if current price exists
		if holding.CurrentPricePaisa.Valid {
			currentValuePaisa := int64(holding.Quantity * float64(holding.CurrentPricePaisa.Int64))
			unrealizedPnlPaisa := currentValuePaisa - totalCostPaisa

			var unrealizedPnlPercent float64
			if totalCostPaisa > 0 {
				unrealizedPnlPercent = float64(unrealizedPnlPaisa) / float64(totalCostPaisa) * 100
			}

			if err := s.queries.UpdateHoldingPrices(ctx, sqlc.UpdateHoldingPricesParams{
				CurrentPricePaisa:    holding.CurrentPricePaisa,
				CurrentValuePaisa:    sql.NullInt64{Int64: currentValuePaisa, Valid: true},
				UnrealizedPnlPaisa:   sql.NullInt64{Int64: unrealizedPnlPaisa, Valid: true},
				UnrealizedPnlPercent: sql.NullFloat64{Float64: unrealizedPnlPercent, Valid: true},
				Symbol:               wac.Symbol,
			}); err != nil {
				slog.Warn("failed to update holding P&L", "symbol", wac.Symbol, "error", err)
			}
		}

		result.Updated++
	}

	return result, nil
}

// SyncPrices fetches live prices from NEPSE and updates holdings.
func (s *Service) SyncPrices(ctx context.Context) (*SyncResult, error) {
	client, err := nepse.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create NEPSE client: %w", err)
	}
	defer client.Close()

	holdings, err := s.queries.ListHoldings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list holdings: %w", err)
	}

	if len(holdings) == 0 {
		return &SyncResult{}, nil
	}

	symbols := make([]string, 0, len(holdings))
	for _, h := range holdings {
		symbols = append(symbols, h.Symbol)
	}

	prices, err := client.GetPrices(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}

	result := &SyncResult{}

	for _, h := range holdings {
		price, ok := prices[h.Symbol]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: price not found", h.Symbol))
			continue
		}

		// Convert LTP to paisa (1 NPR = 100 paisa)
		currentPricePaisa := int64(price.LTP * 100)
		currentValuePaisa := int64(h.Quantity * float64(currentPricePaisa))
		unrealizedPnlPaisa := currentValuePaisa - h.TotalCostPaisa

		var unrealizedPnlPercent float64
		if h.TotalCostPaisa > 0 {
			unrealizedPnlPercent = float64(unrealizedPnlPaisa) / float64(h.TotalCostPaisa) * 100
		}

		if err := s.queries.UpdateHoldingPrices(ctx, sqlc.UpdateHoldingPricesParams{
			CurrentPricePaisa:   sql.NullInt64{Int64: currentPricePaisa, Valid: true},
			CurrentValuePaisa:   sql.NullInt64{Int64: currentValuePaisa, Valid: true},
			UnrealizedPnlPaisa:  sql.NullInt64{Int64: unrealizedPnlPaisa, Valid: true},
			UnrealizedPnlPercent: sql.NullFloat64{Float64: unrealizedPnlPercent, Valid: true},
			Symbol:              h.Symbol,
		}); err != nil {
			slog.Warn("failed to update holding prices", "symbol", h.Symbol, "error", err)
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", h.Symbol, err))
			continue
		}

		result.Updated++
	}

	return result, nil
}
