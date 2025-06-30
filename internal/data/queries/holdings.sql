-- name: CreateHolding :one
INSERT INTO holdings (portfolio_id, symbol, quantity, average_cost_paisa, last_price_paisa)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetHolding :one
SELECT * FROM holdings WHERE id = ?;

-- name: GetHoldingBySymbol :one
SELECT * FROM holdings WHERE portfolio_id = ? AND symbol = ?;

-- name: ListHoldingsByPortfolio :many
SELECT * FROM holdings WHERE portfolio_id = ? ORDER BY symbol;

-- name: UpdateHolding :one
UPDATE holdings 
SET quantity = ?, average_cost_paisa = ?, last_price_paisa = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateHoldingPrice :exec
UPDATE holdings 
SET last_price_paisa = ?, updated_at = CURRENT_TIMESTAMP
WHERE portfolio_id = ? AND symbol = ?;

-- name: DeleteHolding :exec
DELETE FROM holdings WHERE id = ?;

-- name: GetHoldingValue :one
SELECT 
    h.*,
    (h.quantity * h.average_cost_paisa) as total_cost_paisa,
    (h.quantity * COALESCE(h.last_price_paisa, h.average_cost_paisa)) as total_value_paisa,
    (h.quantity * COALESCE(h.last_price_paisa, h.average_cost_paisa)) - (h.quantity * h.average_cost_paisa) as unrealized_pnl_paisa
FROM holdings h
WHERE h.id = ?;

-- name: ListHoldingsWithValue :many
SELECT 
    h.*,
    (h.quantity * h.average_cost_paisa) as total_cost_paisa,
    (h.quantity * COALESCE(h.last_price_paisa, h.average_cost_paisa)) as total_value_paisa,
    (h.quantity * COALESCE(h.last_price_paisa, h.average_cost_paisa)) - (h.quantity * h.average_cost_paisa) as unrealized_pnl_paisa
FROM holdings h
WHERE h.portfolio_id = ?
ORDER BY h.symbol;