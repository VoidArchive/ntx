-- name: UpsertHolding :one
INSERT INTO holdings (
    symbol, status, target_shares, allocation_percent, carry_rs,
    drawdown_gate, thesis, frozen_reason, review_date
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(symbol) DO UPDATE SET
    status = excluded.status,
    target_shares = excluded.target_shares,
    allocation_percent = excluded.allocation_percent,
    carry_rs = excluded.carry_rs,
    drawdown_gate = excluded.drawdown_gate,
    thesis = excluded.thesis,
    frozen_reason = excluded.frozen_reason,
    review_date = excluded.review_date
RETURNING *;

-- name: GetHolding :one
SELECT * FROM holdings WHERE symbol = ?;

-- name: DeleteHolding :exec
DELETE FROM holdings WHERE symbol = ?;

-- name: ListHoldings :many
-- A holding without a company row is a mutual fund, priced by the caller.
SELECT h.*, c.id AS company_id, c.name AS company_name, c.sector,
    (SELECT close_price FROM prices p WHERE p.company_id = c.id
        ORDER BY p.business_date DESC LIMIT 1) AS close_price
FROM holdings h
LEFT JOIN companies c ON c.symbol = h.symbol
ORDER BY h.symbol;

-- name: InsertTransaction :one
INSERT INTO transactions (holding_id, day, kind, quantity, price)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListTransactions :many
SELECT * FROM transactions ORDER BY holding_id, day, id;

-- name: InsertNote :one
INSERT INTO holding_notes (holding_id, day, text)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListNotes :many
SELECT * FROM holding_notes ORDER BY holding_id, day DESC, id DESC;
