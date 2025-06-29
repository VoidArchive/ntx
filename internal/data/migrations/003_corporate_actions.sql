-- +goose Up
-- Add corporate actions table for NEPSE-specific features

-- Corporate actions table for tracking bonus shares, dividends, etc.
CREATE TABLE corporate_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    action_type TEXT NOT NULL CHECK (action_type IN ('bonus', 'dividend', 'rights', 'split')),
    announcement_date DATE,
    record_date DATE,
    ex_date DATE,
    ratio TEXT, -- e.g., "1:5" for bonus shares
    dividend_amount INTEGER, -- paisa for dividend actions
    processed BOOLEAN DEFAULT FALSE,
    processed_date DATETIME,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for corporate actions
CREATE INDEX idx_corporate_actions_symbol ON corporate_actions(symbol);
CREATE INDEX idx_corporate_actions_type ON corporate_actions(action_type);
CREATE INDEX idx_corporate_actions_processed ON corporate_actions(processed);
CREATE INDEX idx_corporate_actions_ex_date ON corporate_actions(ex_date);
CREATE INDEX idx_corporate_actions_symbol_date ON corporate_actions(symbol, ex_date DESC);

-- Trigger to update corporate_actions updated_at timestamp
-- +goose StatementBegin
CREATE TRIGGER corporate_actions_updated_at 
AFTER UPDATE ON corporate_actions
BEGIN
    UPDATE corporate_actions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS corporate_actions_updated_at;
DROP INDEX IF EXISTS idx_corporate_actions_symbol_date;
DROP INDEX IF EXISTS idx_corporate_actions_ex_date;
DROP INDEX IF EXISTS idx_corporate_actions_processed;
DROP INDEX IF EXISTS idx_corporate_actions_type;
DROP INDEX IF EXISTS idx_corporate_actions_symbol;
DROP TABLE IF EXISTS corporate_actions; 