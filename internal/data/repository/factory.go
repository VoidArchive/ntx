package repository

import (
	"database/sql"
	"ntx/internal/database"
)

// Factory creates and manages repository instances
type Factory struct {
	db      *sql.DB
	queries *database.Queries
}

// NewFactory creates a new repository factory
func NewFactory(db *sql.DB) *Factory {
	return &Factory{
		db:      db,
		queries: database.New(db),
	}
}

// NewRepository creates a new Repository instance with all sub-repositories
func (f *Factory) NewRepository() *Repository {
	return &Repository{
		Portfolio:       NewPortfolioRepository(f.queries),
		Holding:         NewHoldingRepository(f.queries),
		Transaction:     NewTransactionRepository(f.queries),
		CorporateAction: NewCorporateActionRepository(f.queries),
	}
}

// NewTransactor creates a new transactor for handling database transactions
func (f *Factory) NewTransactor() Transactor {
	return NewTransactor(f.db)
}

// Close closes the database connection
func (f *Factory) Close() error {
	return f.db.Close()
}

// Services provides all repository services and transaction support
type Services struct {
	Repository *Repository
	Transactor Transactor
	Factory    *Factory
}

// NewServices creates all repository services from a database connection
func NewServices(db *sql.DB) *Services {
	factory := NewFactory(db)
	return &Services{
		Repository: factory.NewRepository(),
		Transactor: factory.NewTransactor(),
		Factory:    factory,
	}
}

// Close closes all database connections
func (s *Services) Close() error {
	return s.Factory.Close()
}