-- +goose Up

-- Companies: NEPSE listed securities
CREATE TABLE companies (
    symbol TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sector INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    last_synced TEXT NOT NULL
);

-- Fundamentals: valuation metrics, updated daily
CREATE TABLE fundamentals (
    symbol TEXT PRIMARY KEY REFERENCES companies(symbol) ON DELETE CASCADE,
    pe REAL,
    pb REAL,
    eps REAL,
    book_value REAL,
    market_cap REAL,
    dividend_yield REAL,
    roe REAL,
    shares_outstanding INTEGER,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Prices: daily OHLCV data
CREATE TABLE prices (
    symbol TEXT NOT NULL REFERENCES companies(symbol) ON DELETE CASCADE,
    date TEXT NOT NULL,
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    previous_close REAL,
    volume INTEGER NOT NULL,
    turnover INTEGER,
    is_complete INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (symbol, date)
);

CREATE INDEX idx_prices_date ON prices(date);
CREATE INDEX idx_prices_symbol_date ON prices(symbol, date DESC);

-- Reports: quarterly and annual financials
CREATE TABLE reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL REFERENCES companies(symbol) ON DELETE CASCADE,
    type INTEGER NOT NULL, -- 1=quarterly, 2=annual
    fiscal_year INTEGER NOT NULL,
    quarter INTEGER NOT NULL, -- 0 for annual, 1-4 for quarterly
    revenue REAL,
    net_income REAL,
    eps REAL,
    book_value REAL,
    npl_ratio REAL,
    published_at TEXT,
    UNIQUE(symbol, type, fiscal_year, quarter)
);

CREATE INDEX idx_reports_symbol ON reports(symbol);

-- Trading days: track market open/close status
CREATE TABLE trading_days (
    date TEXT PRIMARY KEY,
    is_open INTEGER NOT NULL,
    status TEXT NOT NULL
);

-- +goose Down

DROP TABLE IF EXISTS trading_days;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS prices;
DROP TABLE IF EXISTS fundamentals;
DROP TABLE IF EXISTS companies;
