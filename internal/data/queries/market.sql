-- Market Data Operations

-- name: UpsertMarketData :execresult
INSERT OR REPLACE INTO market_data (symbol, last_price, change_amount, change_percent, volume, timestamp)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetMarketData :one
SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
FROM market_data
WHERE symbol = ? COLLATE NOCASE
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetAllMarketData :many
SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
FROM market_data m1
WHERE m1.timestamp = (
    SELECT MAX(m2.timestamp) 
    FROM market_data m2 
    WHERE m2.symbol = m1.symbol COLLATE NOCASE
)
ORDER BY symbol;

-- name: GetMarketDataBatch :many
SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
FROM market_data m1
WHERE m1.symbol IN (sqlc.slice('symbols')) COLLATE NOCASE
AND m1.timestamp = (
    SELECT MAX(m2.timestamp) 
    FROM market_data m2 
    WHERE m2.symbol = m1.symbol COLLATE NOCASE
);

-- name: DeleteMarketData :exec
DELETE FROM market_data 
WHERE symbol = ? COLLATE NOCASE;

-- name: GetHistoricalPrices :many
SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
FROM market_data
WHERE symbol = ? COLLATE NOCASE 
AND timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC;

-- name: GetLatestPrices :many
SELECT symbol, last_price
FROM market_data m1
WHERE m1.symbol IN (sqlc.slice('symbols')) COLLATE NOCASE
AND m1.timestamp = (
    SELECT MAX(m2.timestamp) 
    FROM market_data m2 
    WHERE m2.symbol = m1.symbol COLLATE NOCASE
);

-- name: CleanupStaleData :exec
DELETE FROM market_data 
WHERE timestamp < ?;

-- name: GetDataAge :one
SELECT 
    symbol,
    CASE 
        WHEN MAX(timestamp) IS NULL THEN NULL
        ELSE (julianday('now') - julianday(MAX(timestamp))) * 24 * 60 * 60
    END as age_seconds
FROM market_data 
WHERE symbol = ? COLLATE NOCASE; 