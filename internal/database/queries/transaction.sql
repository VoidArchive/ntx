-- name: CreateTransaction :exec
INSERT INTO transactions (id, symbol, type, quantity, price_paisa, total_paisa, date, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTransaction :one
SELECT * FROM transactions WHERE id = ? LIMIT 1;

-- name: ListTransactionsBySymbol :many
SELECT * FROM transactions
WHERE symbol = ?
ORDER BY date DESC;

-- name: ListTransactionsBySymbolChronological :many
SELECT * FROM transactions
WHERE symbol = ?
ORDER BY date ASC, created_at ASC;

-- name: ListTransactions :many
SELECT * FROM transactions
ORDER BY date DESC
LIMIT ? OFFSET ?;

-- name: ListTransactionsFiltered :many
SELECT * FROM transactions
WHERE (? = '' OR symbol = ?)
  AND (? = 0 OR type = ?)
ORDER BY date DESC
LIMIT ? OFFSET ?;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions;

-- name: CountTransactionsFiltered :one
SELECT COUNT(*) FROM transactions
WHERE (? = '' OR symbol = ?)
  AND (? = 0 OR type = ?);

-- name: TransactionExists :one
SELECT EXISTS(
    SELECT 1 FROM transactions
    WHERE symbol = ? AND date = ? AND description = ?
) AS exists_flag;

-- name: DeleteAllTransactions :exec
DELETE FROM transactions;

-- name: UpdateTransactionPrices :execrows
UPDATE transactions
SET price_paisa = ?,
    total_paisa = ?
WHERE symbol = ?
  AND type = ?
  AND quantity = ?
  AND total_paisa = 0;

-- name: ListTransactionsWithoutPrices :many
SELECT * FROM transactions
WHERE total_paisa = 0
  AND type IN (1, 2)  -- BUY and SELL only
ORDER BY date ASC;
