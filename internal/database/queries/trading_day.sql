-- name: UpsertTradingDay :exec
INSERT INTO trading_days (date, is_open, status)
VALUES (?, ?, ?)
ON CONFLICT(date) DO UPDATE SET
    is_open = excluded.is_open,
    status = excluded.status;

-- name: GetTradingDay :one
SELECT * FROM trading_days WHERE date = ? LIMIT 1;

-- name: GetLastTradingDay :one
SELECT * FROM trading_days
WHERE is_open = 1
ORDER BY date DESC
LIMIT 1;

-- name: ListTradingDays :many
SELECT * FROM trading_days
WHERE date >= ? AND date <= ?
ORDER BY date;
