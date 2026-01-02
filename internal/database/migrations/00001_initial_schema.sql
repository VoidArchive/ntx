-- +goose Up
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    type INTEGER NOT NULL,
    quantity REAL NOT NULL,
    price_paisa INTEGER NOT NULL,
    total_paisa INTEGER NOT NULL,
    date TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_symbol ON transactions(symbol);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);

CREATE TABLE IF NOT EXISTS holdings (
    symbol TEXT PRIMARY KEY,
    quantity REAL NOT NULL,
    average_cost_paisa INTEGER NOT NULL,
    total_cost_paisa INTEGER NOT NULL,
    current_price_paisa INTEGER,
    current_value_paisa INTEGER,
    unrealized_pnl_paisa INTEGER,
    unrealized_pnl_percent REAL,
    last_updated TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS stocks (
    symbol TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sector INTEGER NOT NULL,
    last_synced TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_symbol;
DROP INDEX IF EXISTS idx_transactions_date;
DROP TABLE IF EXISTS stocks;
DROP TABLE IF EXISTS holdings;
DROP TABLE IF EXISTS transactions;
