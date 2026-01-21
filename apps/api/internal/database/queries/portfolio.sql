-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at FROM users WHERE email = ?;

-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES (?, ?)
RETURNING id, email, password_hash, created_at;

-- name: ListPortfoliosByUser :many
SELECT id, user_id, name, created_at FROM portfolios WHERE user_id = ? ORDER BY created_at DESC;

-- name: GetPortfolio :one
SELECT id, user_id, name, created_at FROM portfolios WHERE id = ? AND user_id = ?;

-- name: CreatePortfolio :one
INSERT INTO portfolios (user_id, name)
VALUES (?, ?)
RETURNING id, user_id, name, created_at;

-- name: DeletePortfolio :exec
DELETE FROM portfolios WHERE id = ? AND user_id = ?;

-- name: ListTransactionsByPortfolio :many
SELECT id, portfolio_id, stock_symbol, transaction_type, quantity, unit_price, transaction_date, created_at
FROM transactions
WHERE portfolio_id = ?
ORDER BY transaction_date DESC, created_at DESC;

-- name: ListTransactionsBySymbol :many
SELECT id, portfolio_id, stock_symbol, transaction_type, quantity, unit_price, transaction_date, created_at
FROM transactions
WHERE portfolio_id = ? AND stock_symbol = ?
ORDER BY transaction_date DESC, created_at DESC;

-- name: GetTransaction :one
SELECT id, portfolio_id, stock_symbol, transaction_type, quantity, unit_price, transaction_date, created_at
FROM transactions
WHERE id = ?;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = ?;

-- name: CreateTransaction :one
INSERT INTO transactions (portfolio_id, stock_symbol, transaction_type, quantity, unit_price, transaction_date)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, portfolio_id, stock_symbol, transaction_type, quantity, unit_price, transaction_date, created_at;

-- name: GetHoldingsByPortfolio :many
SELECT
    stock_symbol,
    SUM(CASE WHEN transaction_type = 'BUY' THEN quantity ELSE -quantity END) as net_quantity,
    SUM(CASE WHEN transaction_type = 'BUY' THEN quantity * unit_price ELSE 0 END) as total_buy_cost,
    SUM(CASE WHEN transaction_type = 'BUY' THEN quantity ELSE 0 END) as total_buy_quantity
FROM transactions
WHERE portfolio_id = ?
GROUP BY stock_symbol
HAVING net_quantity > 0;

