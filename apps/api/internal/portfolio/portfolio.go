// Package portfolio provides portfolio service implementation.
package portfolio

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// ContextKey is a type for context keys.
type ContextKey string

// UserIDKey is the context key for user ID.
const UserIDKey ContextKey = "user_id"

// PortfolioService implements the PortfolioService gRPC service.
type PortfolioService struct {
	queries *sqlc.Queries
}

// NewPortfolioService creates a new PortfolioService.
func NewPortfolioService(queries *sqlc.Queries) *PortfolioService {
	return &PortfolioService{queries: queries}
}

// getUserID extracts user ID from context (set by auth middleware).
func getUserID(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	if !ok || userID == 0 {
		return 0, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	return userID, nil
}

// ListPortfolios returns all portfolios for the authenticated user.
func (s *PortfolioService) ListPortfolios(
	ctx context.Context,
	_ *connect.Request[ntxv1.ListPortfoliosRequest],
) (*connect.Response[ntxv1.ListPortfoliosResponse], error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	portfolios, err := s.queries.ListPortfoliosByUser(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := make([]*ntxv1.Portfolio, len(portfolios))
	for i, p := range portfolios {
		createdAt := ""
		if p.CreatedAt.Valid {
			createdAt = p.CreatedAt.Time.Format(time.RFC3339)
		}
		result[i] = &ntxv1.Portfolio{
			Id:        p.ID,
			Name:      p.Name,
			CreatedAt: createdAt,
		}
	}

	return connect.NewResponse(&ntxv1.ListPortfoliosResponse{
		Portfolios: result,
	}), nil
}

// CreatePortfolio creates a new portfolio.
func (s *PortfolioService) CreatePortfolio(
	ctx context.Context,
	req *connect.Request[ntxv1.CreatePortfolioRequest],
) (*connect.Response[ntxv1.CreatePortfolioResponse], error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name is required"))
	}

	portfolio, err := s.queries.CreatePortfolio(ctx, sqlc.CreatePortfolioParams{
		UserID: userID,
		Name:   req.Msg.Name,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	createdAt := ""
	if portfolio.CreatedAt.Valid {
		createdAt = portfolio.CreatedAt.Time.Format(time.RFC3339)
	}

	return connect.NewResponse(&ntxv1.CreatePortfolioResponse{
		Portfolio: &ntxv1.Portfolio{
			Id:        portfolio.ID,
			Name:      portfolio.Name,
			CreatedAt: createdAt,
		},
	}), nil
}

// AddTransaction adds a transaction to a portfolio.
func (s *PortfolioService) AddTransaction(
	ctx context.Context,
	req *connect.Request[ntxv1.AddTransactionRequest],
) (*connect.Response[ntxv1.AddTransactionResponse], error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Verify portfolio belongs to user
	_, err = s.queries.GetPortfolio(ctx, sqlc.GetPortfolioParams{
		ID:     req.Msg.PortfolioId,
		UserID: userID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("portfolio not found"))
	}

	// Validate input
	if req.Msg.StockSymbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("stock_symbol is required"))
	}
	if req.Msg.Quantity <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("quantity must be positive"))
	}
	if req.Msg.UnitPrice <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("unit_price must be positive"))
	}

	transactionType := "BUY"
	if req.Msg.TransactionType == ntxv1.TransactionType_TRANSACTION_TYPE_SELL {
		transactionType = "SELL"
	}

	transactionDate, err := time.Parse("2006-01-02", req.Msg.TransactionDate)
	if err != nil {
		transactionDate = time.Now()
	}

	tx, err := s.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		PortfolioID:     req.Msg.PortfolioId,
		StockSymbol:     req.Msg.StockSymbol,
		TransactionType: transactionType,
		Quantity:        req.Msg.Quantity,
		UnitPrice:       req.Msg.UnitPrice,
		TransactionDate: transactionDate,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.AddTransactionResponse{
		Transaction: &ntxv1.Transaction{
			Id:              tx.ID,
			PortfolioId:     tx.PortfolioID,
			StockSymbol:     tx.StockSymbol,
			TransactionType: req.Msg.TransactionType,
			Quantity:        tx.Quantity,
			UnitPrice:       tx.UnitPrice,
			TransactionDate: tx.TransactionDate.Format("2006-01-02"),
		},
	}), nil
}

// GetPortfolioSummary returns the portfolio summary with holdings.
func (s *PortfolioService) GetPortfolioSummary(
	ctx context.Context,
	req *connect.Request[ntxv1.GetPortfolioSummaryRequest],
) (*connect.Response[ntxv1.GetPortfolioSummaryResponse], error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Verify portfolio belongs to user
	portfolio, err := s.queries.GetPortfolio(ctx, sqlc.GetPortfolioParams{
		ID:     req.Msg.PortfolioId,
		UserID: userID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("portfolio not found"))
	}

	// Get aggregated holdings
	holdingsData, err := s.queries.GetHoldingsByPortfolio(ctx, req.Msg.PortfolioId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Fetch current prices for all holdings
	priceMap, err := s.fetchCurrentPrices(ctx, holdingsData)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var holdings []*ntxv1.Holding
	var totalInvested, totalCurrentValue float64

	for _, h := range holdingsData {
		qty := h.NetQuantity.Float64
		if qty <= 0 {
			continue
		}

		avgBuyPrice := 0.0
		if h.TotalBuyQuantity.Float64 > 0 {
			avgBuyPrice = h.TotalBuyCost.Float64 / h.TotalBuyQuantity.Float64
		}

		currentPrice := priceMap[h.StockSymbol]
		totalValue := qty * currentPrice
		invested := qty * avgBuyPrice
		profitLoss := totalValue - invested
		profitLossPercent := 0.0
		if invested > 0 {
			profitLossPercent = (profitLoss / invested) * 100
		}

		holdings = append(holdings, &ntxv1.Holding{
			StockSymbol:       h.StockSymbol,
			Quantity:          int64(qty),
			AvgBuyPrice:       avgBuyPrice,
			CurrentPrice:      currentPrice,
			TotalValue:        totalValue,
			ProfitLoss:        profitLoss,
			ProfitLossPercent: profitLossPercent,
		})

		totalInvested += invested
		totalCurrentValue += totalValue
	}

	totalPL := totalCurrentValue - totalInvested
	totalPLPercent := 0.0
	if totalInvested > 0 {
		totalPLPercent = (totalPL / totalInvested) * 100
	}

	return connect.NewResponse(&ntxv1.GetPortfolioSummaryResponse{
		Summary: &ntxv1.PortfolioSummary{
			PortfolioId:            portfolio.ID,
			PortfolioName:          portfolio.Name,
			Holdings:               holdings,
			TotalInvested:          totalInvested,
			TotalCurrentValue:      totalCurrentValue,
			TotalProfitLoss:        totalPL,
			TotalProfitLossPercent: totalPLPercent,
		},
	}), nil
}

// fetchCurrentPrices fetches current prices for the given holdings.
func (s *PortfolioService) fetchCurrentPrices(ctx context.Context, holdings []sqlc.GetHoldingsByPortfolioRow) (map[string]float64, error) {
	prices := make(map[string]float64)

	for _, h := range holdings {
		// Try to get price from our database by symbol
		price, err := s.queries.GetLatestPriceBySymbol(ctx, h.StockSymbol)
		if err != nil {
			// If no price found, use 0
			prices[h.StockSymbol] = 0
			continue
		}

		// Use LTP if available, otherwise close
		if price.LastTradedPrice.Valid {
			prices[h.StockSymbol] = price.LastTradedPrice.Float64
		} else if price.ClosePrice.Valid {
			prices[h.StockSymbol] = price.ClosePrice.Float64
		} else {
			prices[h.StockSymbol] = 0
		}
	}

	return prices, nil
}
