-- name: UpsertDividend :exec
INSERT INTO dividends (symbol, fiscal_year, cash_percent, bonus_percent, headline, published_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(symbol, fiscal_year) DO UPDATE SET
    cash_percent = excluded.cash_percent,
    bonus_percent = excluded.bonus_percent,
    headline = excluded.headline,
    published_at = excluded.published_at;

-- name: GetDividends :many
SELECT * FROM dividends
WHERE symbol = ?
ORDER BY fiscal_year DESC;

-- name: GetLatestDividend :one
SELECT * FROM dividends
WHERE symbol = ?
ORDER BY fiscal_year DESC
LIMIT 1;

-- name: ListDividends :many
SELECT * FROM dividends
ORDER BY fiscal_year DESC, symbol;
