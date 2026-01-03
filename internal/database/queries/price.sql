-- name: UpsertPrice :exec
INSERT INTO prices (symbol, date, open, high, low, close, previous_close, volume, turnover, is_complete)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(symbol, date) DO UPDATE SET
    open = excluded.open,
    high = excluded.high,
    low = excluded.low,
    close = excluded.close,
    previous_close = excluded.previous_close,
    volume = excluded.volume,
    turnover = excluded.turnover,
    is_complete = excluded.is_complete;

-- name: GetLatestPrice :one
SELECT * FROM prices
WHERE symbol = ?
ORDER BY date DESC
LIMIT 1;

-- name: GetLatestPriceDates :many
SELECT symbol, MAX(date) as latest_date FROM prices GROUP BY symbol;

-- name: GetPriceHistory :many
SELECT * FROM prices
WHERE symbol = ?
  AND date >= ?
  AND date <= ?
ORDER BY date ASC;

-- name: GetPricesForDate :many
SELECT * FROM prices
WHERE date = ?
ORDER BY symbol;

-- name: MarkPricesComplete :exec
UPDATE prices SET is_complete = 1 WHERE date = ?;

-- name: Get52WeekHighLow :one
SELECT
    MAX(high) as week_52_high,
    MIN(low) as week_52_low
FROM prices
WHERE symbol = ?
  AND date >= date('now', '-52 weeks');

-- name: GetTopGainers :many
SELECT p.* FROM prices p
INNER JOIN (
    SELECT symbol, MAX(date) as max_date
    FROM prices
    GROUP BY symbol
) latest ON p.symbol = latest.symbol AND p.date = latest.max_date
WHERE p.previous_close IS NOT NULL AND p.close > p.previous_close
ORDER BY ((p.close - p.previous_close) / p.previous_close) DESC
LIMIT ?;

-- name: GetTopGainersBySector :many
SELECT p.* FROM prices p
INNER JOIN companies c ON p.symbol = c.symbol
INNER JOIN (
    SELECT symbol, MAX(date) as max_date
    FROM prices
    GROUP BY symbol
) latest ON p.symbol = latest.symbol AND p.date = latest.max_date
WHERE c.sector = ? AND p.previous_close IS NOT NULL AND p.close > p.previous_close
ORDER BY ((p.close - p.previous_close) / p.previous_close) DESC
LIMIT ?;

-- name: GetTopLosers :many
SELECT p.* FROM prices p
INNER JOIN (
    SELECT symbol, MAX(date) as max_date
    FROM prices
    GROUP BY symbol
) latest ON p.symbol = latest.symbol AND p.date = latest.max_date
WHERE p.previous_close IS NOT NULL AND p.close < p.previous_close
ORDER BY ((p.close - p.previous_close) / p.previous_close) ASC
LIMIT ?;

-- name: GetTopLosersBySector :many
SELECT p.* FROM prices p
INNER JOIN companies c ON p.symbol = c.symbol
INNER JOIN (
    SELECT symbol, MAX(date) as max_date
    FROM prices
    GROUP BY symbol
) latest ON p.symbol = latest.symbol AND p.date = latest.max_date
WHERE c.sector = ? AND p.previous_close IS NOT NULL AND p.close < p.previous_close
ORDER BY ((p.close - p.previous_close) / p.previous_close) ASC
LIMIT ?;
