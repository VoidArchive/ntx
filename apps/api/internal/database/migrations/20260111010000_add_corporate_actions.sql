-- +goose Up
CREATE TABLE corporate_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    fiscal_year TEXT NOT NULL,
    bonus_percentage REAL DEFAULT 0,
    right_percentage REAL,
    cash_dividend REAL,
    submitted_date TEXT,
    UNIQUE(company_id, fiscal_year, submitted_date)
);

CREATE INDEX idx_corporate_actions_company ON corporate_actions(company_id);

-- +goose Down
DROP TABLE IF EXISTS corporate_actions;
