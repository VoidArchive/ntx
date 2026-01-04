-- +goose Up
CREATE TABLE companies (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    symbol TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    email TEXT,
    website TEXT,
    sector TEXT NOT NULL,
    instrument_type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_companies_sector ON companies(sector);
CREATE INDEX idx_companies_name ON companies(name);

-- +goose Down
DROP TABLE IF EXISTS companies;

