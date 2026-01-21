-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS portfolios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    stock_symbol TEXT NOT NULL,
    transaction_type TEXT NOT NULL CHECK(transaction_type IN ('BUY', 'SELL')),
    quantity INTEGER NOT NULL CHECK(quantity > 0),
    unit_price REAL NOT NULL CHECK(unit_price > 0),
    transaction_date DATE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_portfolio_id ON transactions(portfolio_id);
CREATE INDEX idx_transactions_stock_symbol ON transactions(stock_symbol);
CREATE INDEX idx_portfolios_user_id ON portfolios(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_portfolios_user_id;
DROP INDEX IF EXISTS idx_transactions_stock_symbol;
DROP INDEX IF EXISTS idx_transactions_portfolio_id;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS portfolios;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
