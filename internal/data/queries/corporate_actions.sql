-- name: CreateCorporateAction :one
INSERT INTO corporate_actions (symbol, action_type, announcement_date, record_date, execution_date, ratio_from, ratio_to, amount_paisa, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetCorporateAction :one
SELECT * FROM corporate_actions WHERE id = ?;

-- name: ListCorporateActions :many
SELECT * FROM corporate_actions ORDER BY record_date DESC, created_at DESC;

-- name: ListCorporateActionsBySymbol :many
SELECT * FROM corporate_actions 
WHERE symbol = ? 
ORDER BY record_date DESC, created_at DESC;

-- name: ListCorporateActionsByType :many
SELECT * FROM corporate_actions 
WHERE action_type = ? 
ORDER BY record_date DESC, created_at DESC;

-- name: ListCorporateActionsByDateRange :many
SELECT * FROM corporate_actions 
WHERE record_date BETWEEN ? AND ?
ORDER BY record_date DESC, created_at DESC;

-- name: UpdateCorporateAction :one
UPDATE corporate_actions 
SET symbol = ?, action_type = ?, announcement_date = ?, record_date = ?, 
    execution_date = ?, ratio_from = ?, ratio_to = ?, amount_paisa = ?, notes = ?
WHERE id = ?
RETURNING *;

-- name: DeleteCorporateAction :exec
DELETE FROM corporate_actions WHERE id = ?;

-- name: GetPendingCorporateActions :many
SELECT * FROM corporate_actions 
WHERE execution_date IS NULL OR execution_date > date('now')
ORDER BY record_date ASC;

-- name: GetCorporateActionsBySymbolAndDate :many
SELECT * FROM corporate_actions 
WHERE symbol = ? AND record_date <= ? 
ORDER BY record_date DESC;