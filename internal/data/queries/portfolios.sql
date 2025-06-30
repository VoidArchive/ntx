-- name: CreatePortfolio :one
INSERT INTO portfolios (name, description, currency)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetPortfolio :one
SELECT * FROM portfolios WHERE id = ?;

-- name: ListPortfolios :many
SELECT * FROM portfolios ORDER BY created_at DESC;

-- name: UpdatePortfolio :one
UPDATE portfolios 
SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeletePortfolio :exec
DELETE FROM portfolios WHERE id = ?;

-- name: GetPortfolioStats :one
SELECT 
    p.id,
    p.name,
    COUNT(h.id) as holding_count,
    COALESCE(SUM(h.quantity * h.average_cost_paisa), 0) as total_cost_paisa,
    COALESCE(SUM(h.quantity * COALESCE(h.last_price_paisa, h.average_cost_paisa)), 0) as total_value_paisa
FROM portfolios p
LEFT JOIN holdings h ON p.id = h.portfolio_id
WHERE p.id = ?
GROUP BY p.id, p.name;