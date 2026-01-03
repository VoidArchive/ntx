-- +goose Up

-- Dividends: dividend declarations history
CREATE TABLE dividends (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL REFERENCES companies(symbol) ON DELETE CASCADE,
    fiscal_year TEXT NOT NULL,
    cash_percent REAL NOT NULL DEFAULT 0,
    bonus_percent REAL NOT NULL DEFAULT 0,
    headline TEXT,
    published_at TEXT,
    UNIQUE(symbol, fiscal_year)
);

CREATE INDEX idx_dividends_symbol ON dividends(symbol);

-- +goose Down

DROP TABLE IF EXISTS dividends;
