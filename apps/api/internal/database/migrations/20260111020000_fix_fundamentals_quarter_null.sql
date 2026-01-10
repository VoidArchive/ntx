-- +goose Up
-- Fix NULL quarter causing duplicate rows (SQLite UNIQUE constraint doesn't work with NULL)

-- Create new table with quarter NOT NULL DEFAULT ''
CREATE TABLE fundamentals_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    fiscal_year TEXT NOT NULL,
    quarter TEXT NOT NULL DEFAULT '',
    eps REAL,
    pe_ratio REAL,
    book_value REAL,
    paid_up_capital REAL,
    profit_amount REAL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, fiscal_year, quarter)
);

-- Copy data, deduplicating by taking the first entry per (company_id, fiscal_year, quarter)
INSERT INTO fundamentals_new (company_id, fiscal_year, quarter, eps, pe_ratio, book_value, paid_up_capital, profit_amount, created_at, updated_at)
SELECT company_id, fiscal_year, COALESCE(quarter, ''), eps, pe_ratio, book_value, paid_up_capital, profit_amount, created_at, updated_at
FROM fundamentals
WHERE id IN (
    SELECT MIN(id) FROM fundamentals
    GROUP BY company_id, fiscal_year, COALESCE(quarter, '')
);

-- Drop old table and rename
DROP TABLE fundamentals;
ALTER TABLE fundamentals_new RENAME TO fundamentals;

-- Recreate indexes
CREATE INDEX idx_fundamentals_company_id ON fundamentals(company_id);
CREATE INDEX idx_fundamentals_fiscal_year ON fundamentals(fiscal_year);

-- +goose Down
-- Revert to nullable quarter (will lose deduplication)
CREATE TABLE fundamentals_old (
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

INSERT INTO fundamentals_old (company_id, fiscal_year, quarter, eps, pe_ratio, book_value, paid_up_capital, profit_amount, created_at, updated_at)
SELECT company_id, fiscal_year, NULLIF(quarter, ''), eps, pe_ratio, book_value, paid_up_capital, profit_amount, created_at, updated_at
FROM fundamentals;

DROP TABLE fundamentals;
ALTER TABLE fundamentals_old RENAME TO fundamentals;

CREATE INDEX idx_fundamentals_company_id ON fundamentals(company_id);
CREATE INDEX idx_fundamentals_fiscal_year ON fundamentals(fiscal_year);
