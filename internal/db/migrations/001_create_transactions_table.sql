-- +goose Up
CREATE TABLE transactions (
    id integer PRIMARY KEY AUTOINCREMENT,
    scrip text NOT NULL,
    date text NOT NULL,
    quantity integer NOT NULL, -- positive for buy, negative for sell
    price integer, -- price in paisa (NULL if not entered)
    transaction_type text NOT NULL, -- 'IPO', 'BONUS', 'REGULAR', 'RIGHTS', etc.
    description text, -- original meroshare description
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_transactions_scrip ON transactions (scrip);

CREATE INDEX idx_transactions_date ON transactions (date);

CREATE INDEX idx_transactions_scrip_date ON transactions (scrip, date);

-- +goose Down
DROP TABLE transactions;


