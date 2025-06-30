-- +goose Up
CREATE TABLE portfolios (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    currency TEXT DEFAULT 'NPR',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE holdings (
    id INTEGER PRIMARY KEY,
    portfolio_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    average_cost_paisa INTEGER NOT NULL,
    last_price_paisa INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id),
    UNIQUE(portfolio_id, symbol)
);

CREATE TABLE transactions (
    id INTEGER PRIMARY KEY,
    portfolio_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('buy', 'sell')),
    quantity INTEGER NOT NULL,
    price_paisa INTEGER NOT NULL,
    commission_paisa INTEGER DEFAULT 0,
    tax_paisa INTEGER DEFAULT 0,
    transaction_date DATE NOT NULL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

CREATE TABLE corporate_actions (
    id INTEGER PRIMARY KEY,
    symbol TEXT NOT NULL,
    action_type TEXT NOT NULL CHECK (action_type IN ('bonus', 'dividend', 'split', 'rights')),
    announcement_date DATE NOT NULL,
    record_date DATE NOT NULL,
    execution_date DATE,
    ratio_from INTEGER,
    ratio_to INTEGER,
    amount_paisa INTEGER,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_holdings_portfolio_id ON holdings(portfolio_id);
CREATE INDEX idx_holdings_symbol ON holdings(symbol);
CREATE INDEX idx_transactions_portfolio_id ON transactions(portfolio_id);  
CREATE INDEX idx_transactions_symbol ON transactions(symbol);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_corporate_actions_symbol ON corporate_actions(symbol);
CREATE INDEX idx_corporate_actions_record_date ON corporate_actions(record_date);

-- +goose Down
DROP TABLE IF EXISTS corporate_actions;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS holdings;
DROP TABLE IF EXISTS portfolios;