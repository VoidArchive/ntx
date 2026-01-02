-- name: UpsertHolding :exec
INSERT INTO holdings (symbol, quantity, average_cost_paisa, total_cost_paisa, realized_pnl_paisa, last_updated)
VALUES (?, ?, ?, ?, ?, datetime('now'))
ON CONFLICT(symbol) DO UPDATE SET
    quantity = excluded.quantity,
    average_cost_paisa = excluded.average_cost_paisa,
    total_cost_paisa = excluded.total_cost_paisa,
    realized_pnl_paisa = excluded.realized_pnl_paisa,
    last_updated = excluded.last_updated;

-- name: GetHolding :one
SELECT * FROM holdings WHERE symbol = ? LIMIT 1;

-- name: ListHoldings :many
SELECT * FROM holdings ORDER BY symbol;

-- name: UpdateHoldingPrices :exec
UPDATE holdings
SET current_price_paisa = ?,
    current_value_paisa = ?,
    unrealized_pnl_paisa = ?,
    unrealized_pnl_percent = ?,
    last_updated = datetime('now')
WHERE symbol = ?;

-- name: DeleteHolding :exec
DELETE FROM holdings WHERE symbol = ?;

-- name: UpdateHoldingCosts :exec
UPDATE holdings
SET average_cost_paisa = ?,
    total_cost_paisa = ?,
    last_updated = datetime('now')
WHERE symbol = ?;
