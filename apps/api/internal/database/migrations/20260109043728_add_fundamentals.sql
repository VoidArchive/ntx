-- +goose Up
CREATE TABLE fundamentals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    fiscal_year TEXT NOT NULL,
    quarter TEXT,
    eps REAL,
    pe_ratio REAL,
    book_value REAL,
    paid_up_capital REAL,
    profit_amount REAL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, fiscal_year, quarter)
);

CREATE INDEX idx_fundamentals_company_id ON fundamentals(company_id);
CREATE INDEX idx_fundamentals_fiscal_year ON fundamentals(fiscal_year);

-- +goose Down
DROP TABLE IF EXISTS fundamentals;
