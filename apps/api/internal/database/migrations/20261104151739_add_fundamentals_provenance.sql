-- +goose Up
-- Provenance per ADR 0002: when the exchange published the report and which filing it came from.
ALTER TABLE fundamentals ADD COLUMN published_at TEXT; -- ISO date (YYYY-MM-DD), the filing's submitted date
ALTER TABLE fundamentals ADD COLUMN source_url TEXT;
ALTER TABLE fundamentals ADD COLUMN fiscal_year_np TEXT; -- go-nepse FYNameNepali verbatim, e.g. 2080-2081

-- +goose Down
ALTER TABLE fundamentals DROP COLUMN fiscal_year_np;
ALTER TABLE fundamentals DROP COLUMN source_url;
ALTER TABLE fundamentals DROP COLUMN published_at;
