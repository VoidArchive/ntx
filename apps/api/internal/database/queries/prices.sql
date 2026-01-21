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

-- name: GetPriceByDate :one
SELECT * FROM prices
WHERE company_id = ? AND business_date = ?;

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

