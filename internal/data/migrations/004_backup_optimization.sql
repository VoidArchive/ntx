-- +goose Up
-- Add backup tracking and optimize schema

-- Backup tracking table
CREATE TABLE backup_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    backup_path TEXT NOT NULL,
    backup_size INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    restored_at DATETIME,
    notes TEXT
);

-- Create index for backup history
CREATE INDEX idx_backup_history_created_at ON backup_history(created_at DESC);

-- Add validation constraints and optimize existing tables

-- Update portfolio table constraints
CREATE TABLE portfolio_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL UNIQUE COLLATE NOCASE,
    quantity INTEGER NOT NULL CHECK (quantity >= 0),
    avg_cost INTEGER NOT NULL CHECK (avg_cost > 0),
    purchase_date DATE,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old portfolio table
INSERT INTO portfolio_new SELECT * FROM portfolio;

-- Drop old table and rename new one
DROP TABLE portfolio;
ALTER TABLE portfolio_new RENAME TO portfolio;

-- Recreate indexes and triggers for portfolio
CREATE INDEX idx_portfolio_symbol ON portfolio(symbol);
CREATE INDEX idx_portfolio_purchase_date ON portfolio(purchase_date);

-- +goose StatementBegin
CREATE TRIGGER portfolio_updated_at 
AFTER UPDATE ON portfolio
BEGIN
    UPDATE portfolio SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- Add cleanup view for stale market data
CREATE VIEW stale_market_data AS
SELECT symbol, MAX(timestamp) as latest_timestamp
FROM market_data 
WHERE timestamp < datetime('now', '-1 day')
GROUP BY symbol;

-- +goose Down
DROP VIEW IF EXISTS stale_market_data;
DROP TRIGGER IF EXISTS portfolio_updated_at;
DROP INDEX IF EXISTS idx_portfolio_purchase_date;
DROP INDEX IF EXISTS idx_portfolio_symbol;
DROP INDEX IF EXISTS idx_backup_history_created_at;
DROP TABLE IF EXISTS backup_history;

-- Note: Reverting portfolio table changes is complex and risky
-- In production, consider making this irreversible 