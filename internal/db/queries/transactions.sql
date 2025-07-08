-- name: CreateTransaction :one
INSERT INTO transactions (scrip, date, quantity, price, transaction_type, description)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTransactionsByScripOrderedByDate :many
SELECT * FROM transactions 
WHERE scrip = ? 
ORDER BY date ASC, id ASC;

-- name: GetTransactionsWithoutPrice :many
SELECT * FROM transactions 
WHERE price IS NULL AND transaction_type IN ('IPO', 'REGULAR')
ORDER BY scrip, date ASC;

-- name: UpdateTransactionPrice :exec
UPDATE transactions 
SET price = ? 
WHERE id = ?;

-- name: GetAllTransactions :many
SELECT * FROM transactions 
ORDER BY date ASC, id ASC;

-- name: GetTransactionById :one
SELECT * FROM transactions 
WHERE id = ?;

-- name: GetUniqueScripList :many
SELECT DISTINCT scrip FROM transactions 
ORDER BY scrip;

-- name: GetTransactionsByDateRange :many
SELECT * FROM transactions 
WHERE date BETWEEN ? AND ?
ORDER BY date ASC, id ASC;