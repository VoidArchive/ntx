package services

import (
	"context"
	"fmt"
	"ntx/internal/data/repository"
	"ntx/internal/portfolio/models"
	"time"
)

// PortfolioService handles portfolio business operations
type PortfolioService struct {
	repo       *repository.Repository
	transactor repository.Transactor
	calculator *CalculatorService
}

// NewPortfolioService creates a new portfolio service
func NewPortfolioService(repo *repository.Repository, transactor repository.Transactor) *PortfolioService {
	return &PortfolioService{
		repo:       repo,
		transactor: transactor,
		calculator: NewCalculatorService(repo),
	}
}

// CreatePortfolio creates a new portfolio with validation
func (s *PortfolioService) CreatePortfolio(ctx context.Context, req CreatePortfolioRequest) (*models.Portfolio, error) {
	// Validate request
	if err := (&req).Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create portfolio
	portfolioData, err := s.repo.Portfolio.Create(ctx, repository.CreatePortfolioRequest{
		Name:        req.Name,
		Description: req.Description,
		Currency:    req.Currency,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	return s.calculator.convertPortfolio(portfolioData), nil
}

// GetPortfolioWithStats retrieves portfolio with complete statistics
func (s *PortfolioService) GetPortfolioWithStats(ctx context.Context, portfolioID int64) (*models.PortfolioStats, error) {
	return s.calculator.CalculatePortfolioStats(ctx, portfolioID)
}

// GetPortfolioHoldings retrieves all holdings with calculated metrics
func (s *PortfolioService) GetPortfolioHoldings(ctx context.Context, portfolioID int64) ([]models.Holding, error) {
	return s.calculator.CalculateHoldingsForPortfolio(ctx, portfolioID)
}

// ExecuteTransaction processes a buy/sell transaction with all business logic
func (s *PortfolioService) ExecuteTransaction(ctx context.Context, req ExecuteTransactionRequest) (*TransactionResult, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var result *TransactionResult

	// Execute in transaction
	err := s.transactor.WithTx(ctx, func(ctx context.Context, repo *repository.Repository) error {
		// Create transaction record
		transaction := &models.Transaction{
			PortfolioID:     req.PortfolioID,
			Symbol:          req.Symbol,
			TransactionType: req.TransactionType,
			Quantity:        req.Quantity,
			Price:           req.Price,
			Commission:      req.Commission,
			Tax:             req.Tax,
			TransactionDate: req.TransactionDate,
			Notes:           req.Notes,
		}

		// Get current holding to determine if update or create
		currentHolding, err := repo.Holding.GetBySymbol(ctx, req.PortfolioID, req.Symbol)
		var existingHolding *models.Holding
		if err == nil {
			h := s.calculator.convertHolding(currentHolding)
			h.CalculateMetrics()
			existingHolding = h
		}

		// Calculate impact
		impact := s.calculator.CalculateTransactionImpact(transaction, existingHolding)

		// Validate transaction is possible
		if err := s.validateTransaction(impact); err != nil {
			return fmt.Errorf("transaction validation failed: %w", err)
		}

		// Create transaction in database
		transactionData, err := repo.Transaction.Create(ctx, repository.CreateTransactionRequest{
			PortfolioID:     req.PortfolioID,
			Symbol:          req.Symbol,
			TransactionType: req.TransactionType,
			Quantity:        req.Quantity,
			PricePaisa:      req.Price.Paisa,
			CommissionPaisa: req.Commission.Paisa,
			TaxPaisa:        req.Tax.Paisa,
			TransactionDate: req.TransactionDate,
			Notes:           req.Notes,
		})
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// Update or create holding
		if impact.NewHolding != nil {
			if existingHolding != nil {
				// Update existing holding
				_, err = repo.Holding.Update(ctx, repository.UpdateHoldingRequest{
					ID:               existingHolding.ID,
					Quantity:         impact.NewHolding.Quantity,
					AverageCostPaisa: impact.NewHolding.AverageCost.Paisa,
					LastPricePaisa:   getLastPricePaisa(impact.NewHolding.LastPrice),
				})
			} else {
				// Create new holding
				_, err = repo.Holding.Create(ctx, repository.CreateHoldingRequest{
					PortfolioID:      req.PortfolioID,
					Symbol:           req.Symbol,
					Quantity:         impact.NewHolding.Quantity,
					AverageCostPaisa: impact.NewHolding.AverageCost.Paisa,
					LastPricePaisa:   getLastPricePaisa(impact.NewHolding.LastPrice),
				})
			}
			if err != nil {
				return fmt.Errorf("failed to update holding: %w", err)
			}
		} else if req.TransactionType == "sell" && existingHolding != nil {
			// Complete sell - remove holding
			if err := repo.Holding.Delete(ctx, existingHolding.ID); err != nil {
				return fmt.Errorf("failed to delete holding: %w", err)
			}
		}

		// Prepare result
		result = &TransactionResult{
			TransactionID:     transactionData.ID,
			Impact:            *impact,
			TransactionAmount: req.Price.MultiplyInt(req.Quantity).Add(req.Commission).Add(req.Tax),
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateHoldingPrice updates the last price for a holding
func (s *PortfolioService) UpdateHoldingPrice(ctx context.Context, portfolioID int64, symbol string, price models.Money) error {
	return s.repo.Holding.UpdatePrice(ctx, portfolioID, symbol, price.Paisa)
}

// validateTransaction validates that a transaction can be executed
func (s *PortfolioService) validateTransaction(impact *TransactionImpact) error {
	if impact.Transaction.IsSell() {
		if impact.CurrentHolding == nil {
			return fmt.Errorf("cannot sell %s: no holding exists", impact.Symbol)
		}
		if impact.CurrentHolding.Quantity < impact.Transaction.Quantity {
			return fmt.Errorf("cannot sell %d shares of %s: only %d shares available",
				impact.Transaction.Quantity, impact.Symbol, impact.CurrentHolding.Quantity)
		}
	}
	return nil
}

// Request and response types
type CreatePortfolioRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Currency    string  `json:"currency"`
}

func (r *CreatePortfolioRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("portfolio name is required")
	}
	if len(r.Name) > 100 {
		return fmt.Errorf("portfolio name too long (max 100 characters)")
	}
	if r.Currency == "" {
		r.Currency = "NPR"
	}
	return nil
}

type ExecuteTransactionRequest struct {
	PortfolioID     int64        `json:"portfolio_id"`
	Symbol          string       `json:"symbol"`
	TransactionType string       `json:"transaction_type"` // "buy" or "sell"
	Quantity        int64        `json:"quantity"`
	Price           models.Money `json:"price"`
	Commission      models.Money `json:"commission"`
	Tax             models.Money `json:"tax"`
	TransactionDate time.Time    `json:"transaction_date"`
	Notes           *string      `json:"notes,omitempty"`
}

func (r ExecuteTransactionRequest) Validate() error {
	if r.PortfolioID <= 0 {
		return fmt.Errorf("portfolio ID is required")
	}
	if r.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if r.TransactionType != "buy" && r.TransactionType != "sell" {
		return fmt.Errorf("transaction type must be 'buy' or 'sell'")
	}
	if r.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if r.Price.IsZero() || r.Price.IsNegative() {
		return fmt.Errorf("price must be positive")
	}
	if r.Commission.IsNegative() {
		return fmt.Errorf("commission cannot be negative")
	}
	if r.Tax.IsNegative() {
		return fmt.Errorf("tax cannot be negative")
	}
	if r.TransactionDate.IsZero() {
		return fmt.Errorf("transaction date is required")
	}
	return nil
}

type TransactionResult struct {
	TransactionID     int64             `json:"transaction_id"`
	Impact            TransactionImpact `json:"impact"`
	TransactionAmount models.Money      `json:"transaction_amount"`
}

// Helper functions
func getLastPricePaisa(price *models.Money) *int64 {
	if price == nil {
		return nil
	}
	return &price.Paisa
}

