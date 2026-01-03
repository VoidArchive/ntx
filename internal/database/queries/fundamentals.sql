-- name: UpsertFundamentals :exec
INSERT INTO fundamentals (symbol, pe, pb, eps, book_value, market_cap, dividend_yield, roe, shares_outstanding, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
ON CONFLICT(symbol) DO UPDATE SET
    pe = excluded.pe,
    pb = excluded.pb,
    eps = excluded.eps,
    book_value = excluded.book_value,
    market_cap = excluded.market_cap,
    dividend_yield = excluded.dividend_yield,
    roe = excluded.roe,
    shares_outstanding = excluded.shares_outstanding,
    updated_at = excluded.updated_at;

-- name: GetFundamentals :one
SELECT * FROM fundamentals WHERE symbol = ? LIMIT 1;

-- name: ListFundamentals :many
SELECT * FROM fundamentals ORDER BY symbol;
