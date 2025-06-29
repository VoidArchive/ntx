-- Transaction Operations

-- name: CreateTransaction :execresult
INSERT INTO transactions (type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTransaction :one
SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
FROM transactions
WHERE id = ?;

-- name: GetTransactionsBySymbol :many
SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
FROM transactions
WHERE symbol = ? COLLATE NOCASE
ORDER BY date DESC, created_at DESC;

-- name: GetTransactionsByDateRange :many
SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
FROM transactions
WHERE date BETWEEN ? AND ?
ORDER BY date DESC, created_at DESC;

-- name: GetAllTransactions :many
SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
FROM transactions
ORDER BY date DESC, created_at DESC;

-- name: UpdateTransaction :exec
UPDATE transactions 
SET type = ?, symbol = ?, quantity = ?, price = ?, total_amount = ?, fees = ?, date = ?, notes = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = ?;

-- name: GetTransactionSummary :one
SELECT 
    symbol,
    COUNT(*) as total_transactions,
    SUM(CASE WHEN type = 'buy' THEN quantity ELSE -quantity END) as net_quantity,
    SUM(CASE WHEN type = 'buy' THEN total_amount ELSE 0 END) as total_invested,
    SUM(CASE WHEN type = 'sell' THEN total_amount ELSE 0 END) as total_received,
    SUM(fees) as total_fees
FROM transactions
WHERE symbol = ? COLLATE NOCASE;

-- name: CalculateAverageCost :one
SELECT 
    CASE 
        WHEN SUM(CASE WHEN type = 'buy' THEN quantity ELSE 0 END) = 0 THEN 0
        ELSE SUM(CASE WHEN type = 'buy' THEN total_amount ELSE 0 END) / 
             SUM(CASE WHEN type = 'buy' THEN quantity ELSE 0 END)
    END as avg_cost
FROM transactions
WHERE symbol = ? COLLATE NOCASE;

-- name: GetTransactionsByType :many
SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
FROM transactions
WHERE type = ?
ORDER BY date DESC, created_at DESC;

-- name: GetRecentTransactions :many
SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
FROM transactions
ORDER BY date DESC, created_at DESC
LIMIT ?; 