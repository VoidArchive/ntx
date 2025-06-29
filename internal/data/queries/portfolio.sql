-- Portfolio Operations

-- name: CreateHolding :execresult
INSERT INTO portfolio (symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetHolding :one
SELECT id, symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at
FROM portfolio
WHERE id = ?;

-- name: GetHoldingBySymbol :one
SELECT id, symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at
FROM portfolio
WHERE symbol = ? COLLATE NOCASE;

-- name: GetAllHoldings :many
SELECT id, symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at
FROM portfolio
ORDER BY symbol;

-- name: UpdateHolding :exec
UPDATE portfolio
SET quantity = ?, avg_cost = ?, purchase_date = ?, notes = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteHolding :exec
DELETE FROM portfolio WHERE id = ?;

-- name: DeleteHoldingBySymbol :exec
DELETE FROM portfolio WHERE symbol = ? COLLATE NOCASE;

-- name: GetPortfolioValue :one
SELECT COALESCE(SUM(quantity * avg_cost), 0) as total_value
FROM portfolio;

-- name: GetHoldingsWithValues :many
SELECT 
    p.id, p.symbol, p.quantity, p.avg_cost, p.purchase_date, p.notes, p.created_at, p.updated_at,
    COALESCE(m.last_price, 0) as current_price,
    (p.quantity * p.avg_cost) as total_cost,
    (p.quantity * COALESCE(m.last_price, p.avg_cost)) as current_value
FROM portfolio p
LEFT JOIN (
    SELECT symbol, last_price,
           ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) as rn
    FROM market_data
) m ON p.symbol = m.symbol COLLATE NOCASE AND m.rn = 1
ORDER BY p.symbol;

-- name: GetPortfolioSummary :one
SELECT 
    COUNT(*) as total_holdings,
    COALESCE(SUM(quantity * avg_cost), 0) as total_invested,
    COALESCE(SUM(quantity * COALESCE(m.last_price, avg_cost)), 0) as current_value
FROM portfolio p
LEFT JOIN (
    SELECT symbol, last_price,
           ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) as rn
    FROM market_data
) m ON p.symbol = m.symbol COLLATE NOCASE AND m.rn = 1; 