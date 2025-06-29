-- +goose Up
-- Initial schema: portfolio and market_data tables

-- Portfolio holdings table (using integer paisa storage)
CREATE TABLE portfolio (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL UNIQUE,
    quantity INTEGER NOT NULL CHECK (quantity >= 0),
    avg_cost INTEGER NOT NULL CHECK (avg_cost > 0), -- paisa per share
    purchase_date DATE,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index on symbol for fast lookups
CREATE INDEX idx_portfolio_symbol ON portfolio(symbol);
CREATE INDEX idx_portfolio_purchase_date ON portfolio(purchase_date);

-- Market data table (using integer paisa storage)
CREATE TABLE market_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    last_price INTEGER NOT NULL CHECK (last_price > 0), -- paisa
    change_amount INTEGER DEFAULT 0, -- paisa
    change_percent INTEGER DEFAULT 0, -- basis points
    volume INTEGER DEFAULT 0 CHECK (volume >= 0),
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, timestamp)
);

-- Create indexes for market data
CREATE INDEX idx_market_data_symbol ON market_data(symbol);
CREATE INDEX idx_market_data_timestamp ON market_data(timestamp);
CREATE INDEX idx_market_data_symbol_timestamp ON market_data(symbol, timestamp DESC);

-- Trigger to update portfolio updated_at timestamp
-- +goose StatementBegin
CREATE TRIGGER portfolio_updated_at 
AFTER UPDATE ON portfolio
BEGIN
    UPDATE portfolio SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS portfolio_updated_at;
DROP INDEX IF EXISTS idx_market_data_symbol_timestamp;
DROP INDEX IF EXISTS idx_market_data_timestamp;
DROP INDEX IF EXISTS idx_market_data_symbol;
DROP INDEX IF EXISTS idx_portfolio_purchase_date;
DROP INDEX IF EXISTS idx_portfolio_symbol;
DROP TABLE IF EXISTS market_data;
DROP TABLE IF EXISTS portfolio; 