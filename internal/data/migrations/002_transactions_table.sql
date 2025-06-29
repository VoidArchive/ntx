-- +goose Up
-- Add transactions table for trade history

-- Transactions table for tracking buy/sell history
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('buy', 'sell', 'bonus', 'rights', 'split')),
    symbol TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price INTEGER NOT NULL CHECK (price > 0), -- paisa per share
    total_amount INTEGER NOT NULL CHECK (total_amount > 0), -- paisa
    fees INTEGER DEFAULT 0 CHECK (fees >= 0), -- paisa
    date DATE NOT NULL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for transactions
CREATE INDEX idx_transactions_symbol ON transactions(symbol);
CREATE INDEX idx_transactions_date ON transactions(date DESC);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_symbol_date ON transactions(symbol, date DESC);

-- Trigger to update transactions updated_at timestamp
-- +goose StatementBegin
CREATE TRIGGER transactions_updated_at 
AFTER UPDATE ON transactions
BEGIN
    UPDATE transactions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS transactions_updated_at;
DROP INDEX IF EXISTS idx_transactions_symbol_date;
DROP INDEX IF EXISTS idx_transactions_type;
DROP INDEX IF EXISTS idx_transactions_date;
DROP INDEX IF EXISTS idx_transactions_symbol;
DROP TABLE IF EXISTS transactions; 