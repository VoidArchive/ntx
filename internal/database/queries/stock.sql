-- name: UpsertStock :exec
INSERT INTO stocks (symbol, name, sector, last_synced)
VALUES (?, ?, ?, datetime('now'))
ON CONFLICT(symbol) DO UPDATE SET
    name = excluded.name,
    sector = excluded.sector,
    last_synced = excluded.last_synced;

-- name: GetStock :one
SELECT * FROM stocks WHERE symbol = ? LIMIT 1;

-- name: ListStocks :many
SELECT * FROM stocks ORDER BY symbol;

-- name: ListStocksBySector :many
SELECT * FROM stocks WHERE sector = ? ORDER BY symbol;
