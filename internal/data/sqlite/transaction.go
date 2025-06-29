package sqlite

import (
	"database/sql"
	"ntx/internal/data/repository"
)

// Transaction wraps sql.Tx and provides repository access within a transaction
type Transaction struct {
	tx            *sql.Tx
	portfolioRepo *PortfolioRepository
	marketRepo    *MarketDataRepository
	txRepo        *TransactionRepository
}

// Portfolio returns the portfolio repository within this transaction
func (t *Transaction) Portfolio() repository.PortfolioRepository {
	return t.portfolioRepo
}

// MarketData returns the market data repository within this transaction
func (t *Transaction) MarketData() repository.MarketDataRepository {
	return t.marketRepo
}

// Transactions returns the transaction repository within this transaction
func (t *Transaction) Transactions() repository.TransactionRepository {
	return t.txRepo
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}