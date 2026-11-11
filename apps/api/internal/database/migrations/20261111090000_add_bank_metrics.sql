-- +goose Up
-- +goose StatementBegin
-- The owner's quarterly bank figures, each with its filing date and source (ADR 0002).
-- Written by the owner's form in v1; a later scraper writes the same rows. A NULL
-- figure is not recorded, never a zero. fiscal_year is the canonical "2082/83",
-- quarter "Q1".."Q4".
CREATE TABLE bank_metrics (
    id INTEGER PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    fiscal_year TEXT NOT NULL,
    quarter TEXT NOT NULL,
    npl_percent REAL,
    car_percent REAL,
    loan_growth_percent REAL,
    distributable_profit REAL,
    published_date TEXT NOT NULL,
    source_url TEXT NOT NULL,
    UNIQUE (company_id, fiscal_year, quarter)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE bank_metrics;
-- +goose StatementEnd
