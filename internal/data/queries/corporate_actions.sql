-- Corporate Actions Operations

-- name: CreateCorporateAction :execresult
INSERT INTO corporate_actions 
(symbol, action_type, announcement_date, record_date, ex_date, ratio, dividend_amount, processed, notes, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetCorporateAction :one
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
WHERE id = ?;

-- name: GetCorporateActionsBySymbol :many
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
WHERE symbol = ? COLLATE NOCASE
ORDER BY ex_date DESC, created_at DESC;

-- name: GetCorporateActionsByType :many
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
WHERE action_type = ?
ORDER BY ex_date DESC, created_at DESC;

-- name: GetPendingCorporateActions :many
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
WHERE processed = FALSE
ORDER BY ex_date ASC, created_at ASC;

-- name: GetUpcomingCorporateActions :many
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
WHERE ex_date BETWEEN date('now') AND date('now', '+' || ? || ' days')
ORDER BY ex_date ASC, created_at ASC;

-- name: GetAllCorporateActions :many
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
ORDER BY ex_date DESC, created_at DESC;

-- name: UpdateCorporateAction :exec
UPDATE corporate_actions
SET action_type = ?, announcement_date = ?, record_date = ?, ex_date = ?,
    ratio = ?, dividend_amount = ?, processed = ?, processed_date = ?, notes = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteCorporateAction :exec
DELETE FROM corporate_actions WHERE id = ?;

-- name: MarkAsProcessed :exec
UPDATE corporate_actions
SET processed = TRUE, processed_date = ?, updated_at = ?
WHERE id = ?;



-- name: GetCorporateActionsByDateRange :many
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
WHERE ex_date BETWEEN ? AND ?
ORDER BY ex_date DESC, created_at DESC;

-- name: GetDividendHistory :many
SELECT id, symbol, action_type, announcement_date, record_date, ex_date, ratio,
       dividend_amount, processed, processed_date, notes, created_at, updated_at
FROM corporate_actions
WHERE action_type = 'dividend' AND dividend_amount > 0
ORDER BY ex_date DESC, created_at DESC; 