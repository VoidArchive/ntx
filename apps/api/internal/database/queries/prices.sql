-- name: UpsertPrice :exec
INSERT INTO prices (
    company_id, business_date, open_price, high_price, low_price, close_price,
    last_traded_price, previous_close, change_amount, change_percent,
    volume, turnover, trades
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(company_id, business_date) DO UPDATE SET
    open_price = excluded.open_price,
    high_price = excluded.high_price,
    low_price = excluded.low_price,
    close_price = excluded.close_price,
    last_traded_price = excluded.last_traded_price,
    previous_close = excluded.previous_close,
    change_amount = excluded.change_amount,
    change_percent = excluded.change_percent,
    volume = excluded.volume,
    turnover = excluded.turnover,
    trades = excluded.trades;

-- name: GetLatestPrice :one
SELECT * FROM prices
WHERE company_id = ?
ORDER BY business_date DESC
LIMIT 1;

-- name: ListPricesByCompany :many
SELECT * FROM prices
WHERE company_id = ?
ORDER BY business_date DESC
LIMIT ? OFFSET ?;

-- name: ListLatestPrices :many
WITH LatestDates AS (
    SELECT company_id, MAX(business_date) as max_date
    FROM prices
    GROUP BY company_id
)
SELECT p.*
FROM prices p
JOIN LatestDates ld ON p.company_id = ld.company_id AND p.business_date = ld.max_date;

-- name: GetLatestPriceBySymbol :one
SELECT p.*, c.sector as company_sector, c.id as company_id FROM prices p
JOIN companies c ON p.company_id = c.id
WHERE c.symbol = ?
ORDER BY p.business_date DESC
LIMIT 1;


-- The last close of each calendar month for one company since a day, for the
-- goal chart's value by month (issue #67).
-- name: ListMonthEndCloses :many
SELECT substr(p.business_date, 1, 7) AS month, p.close_price
FROM prices p
WHERE p.company_id = ? AND p.business_date >= ? AND p.close_price IS NOT NULL
  AND p.business_date = (
    SELECT MAX(p2.business_date) FROM prices p2
    WHERE p2.company_id = p.company_id AND p2.close_price IS NOT NULL
      AND substr(p2.business_date, 1, 7) = substr(p.business_date, 1, 7)
  )
ORDER BY month;
