-- name: CreateTransaction :one
INSERT INTO transactions (portfolio_id, symbol, transaction_type, quantity, price_paisa, commission_paisa, tax_paisa, transaction_date, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions WHERE id = ?;

-- name: ListTransactionsByPortfolio :many
SELECT * FROM transactions 
WHERE portfolio_id = ? 
ORDER BY transaction_date DESC, created_at DESC;

-- name: ListTransactionsBySymbol :many
SELECT * FROM transactions 
WHERE portfolio_id = ? AND symbol = ? 
ORDER BY transaction_date DESC, created_at DESC;

-- name: ListTransactionsByDateRange :many
SELECT * FROM transactions 
WHERE portfolio_id = ? AND transaction_date BETWEEN ? AND ?
ORDER BY transaction_date DESC, created_at DESC;

-- name: UpdateTransaction :one
UPDATE transactions 
SET symbol = ?, transaction_type = ?, quantity = ?, price_paisa = ?, 
    commission_paisa = ?, tax_paisa = ?, transaction_date = ?, notes = ?
WHERE id = ?
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = ?;

-- name: GetTransactionSummary :one
SELECT 
    symbol,
    SUM(CASE WHEN transaction_type = 'buy' THEN quantity ELSE -quantity END) as net_quantity,
    SUM(CASE WHEN transaction_type = 'buy' THEN quantity * price_paisa ELSE 0 END) as total_buy_value_paisa,
    SUM(CASE WHEN transaction_type = 'sell' THEN quantity * price_paisa ELSE 0 END) as total_sell_value_paisa,
    SUM(commission_paisa + tax_paisa) as total_fees_paisa
FROM transactions 
WHERE portfolio_id = ? AND symbol = ?
GROUP BY symbol;

-- name: GetPortfolioTransactionStats :one
SELECT 
    COUNT(*) as total_transactions,
    SUM(CASE WHEN transaction_type = 'buy' THEN 1 ELSE 0 END) as buy_count,
    SUM(CASE WHEN transaction_type = 'sell' THEN 1 ELSE 0 END) as sell_count,
    SUM(CASE WHEN transaction_type = 'buy' THEN quantity * price_paisa ELSE 0 END) as total_invested_paisa,
    SUM(CASE WHEN transaction_type = 'sell' THEN quantity * price_paisa ELSE 0 END) as total_realized_paisa,
    SUM(commission_paisa + tax_paisa) as total_fees_paisa
FROM transactions 
WHERE portfolio_id = ?;