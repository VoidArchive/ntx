-- name: InsertGate :one
INSERT INTO gates (holding_id, kind, metric, direction, threshold, periods, text, checked_on)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateGate :one
UPDATE gates SET kind = ?, metric = ?, direction = ?, threshold = ?, periods = ?, text = ?, checked_on = ?
WHERE id = ? AND holding_id = ?
RETURNING *;

-- name: DeleteGate :exec
DELETE FROM gates WHERE id = ?;

-- name: ListGates :many
SELECT * FROM gates ORDER BY holding_id, id;

-- name: GetRules :one
SELECT * FROM rules WHERE id = 1;

-- name: UpdateRules :one
UPDATE rules SET
    soft_cap_percent = ?, hard_cap_percent = ?, sector_cap_percent = ?, max_positions = ?,
    min_entry_rs = ?, one_hydro_per_basin = ?, candidate_min_rating = ?, candidate_dividend_record = ?,
    monthly_budget_rs = ?, multiplier_below_median = ?, multiplier_bottom_quarter = ?,
    goal_rs = ?, fund_sip_rs = ?, npl_ceiling_percent = ?, car_floor_percent = ?,
    distributable_above_zero = ?, max_pe_to_sector_median = ?
WHERE id = 1
RETURNING *;
