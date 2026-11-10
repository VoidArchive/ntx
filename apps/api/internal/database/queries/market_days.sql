-- name: CountBreadth :one
-- Compares each company's close with its own previous priced day, because
-- backfilled rows carry no change fields and illiquid stocks skip sessions.
SELECT
    CAST(COUNT(*) AS INTEGER) AS priced,
    CAST(COALESCE(SUM(t.close_price > p.close_price), 0) AS INTEGER) AS advances,
    CAST(COALESCE(SUM(t.close_price < p.close_price), 0) AS INTEGER) AS declines,
    CAST(COALESCE(SUM(t.close_price = p.close_price), 0) AS INTEGER) AS unchanged
FROM prices t
LEFT JOIN prices p ON p.company_id = t.company_id
    AND p.business_date = (
        SELECT MAX(business_date) FROM prices
        WHERE company_id = t.company_id AND business_date < t.business_date
    )
WHERE t.business_date = ?;

-- name: UpsertMarketDay :exec
INSERT INTO market_days (
    business_date, index_close, index_change, index_change_percent,
    turnover, traded_shares, transactions, scrips_traded, market_cap,
    advances, declines, unchanged
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(business_date) DO UPDATE SET
    index_close = excluded.index_close,
    index_change = excluded.index_change,
    index_change_percent = excluded.index_change_percent,
    turnover = excluded.turnover,
    traded_shares = excluded.traded_shares,
    transactions = excluded.transactions,
    scrips_traded = excluded.scrips_traded,
    market_cap = excluded.market_cap,
    advances = excluded.advances,
    declines = excluded.declines,
    unchanged = excluded.unchanged;

-- name: GetLatestMarketDay :one
SELECT * FROM market_days
ORDER BY business_date DESC
LIMIT 1;

-- name: ListIndexCloses :many
-- The index's trailing range for the order sheet's multiplier.
SELECT business_date, index_close FROM market_days
WHERE business_date >= ?
ORDER BY business_date;
