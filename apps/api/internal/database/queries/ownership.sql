-- name: UpsertOwnership :exec
INSERT INTO ownership (company_id, listed_shares, public_shares, public_percent, promoter_shares, promoter_percent, updated_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(company_id) DO UPDATE SET
  listed_shares = excluded.listed_shares,
  public_shares = excluded.public_shares,
  public_percent = excluded.public_percent,
  promoter_shares = excluded.promoter_shares,
  promoter_percent = excluded.promoter_percent,
  updated_at = CURRENT_TIMESTAMP;

-- name: GetOwnership :one
SELECT * FROM ownership WHERE company_id = ?;

-- name: GetOwnershipBySymbol :one
SELECT o.* FROM ownership o
JOIN companies c ON c.id = o.company_id
WHERE c.symbol = ?;

-- name: GetLatestOwnershipHistory :one
SELECT * FROM ownership_history
WHERE company_id = ?
ORDER BY as_of DESC
LIMIT 1;

-- name: UpsertOwnershipHistory :exec
INSERT INTO ownership_history (company_id, as_of, listed_shares, public_shares, public_percent, promoter_shares, promoter_percent)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(company_id, as_of) DO UPDATE SET
  listed_shares = excluded.listed_shares,
  public_shares = excluded.public_shares,
  public_percent = excluded.public_percent,
  promoter_shares = excluded.promoter_shares,
  promoter_percent = excluded.promoter_percent;

-- name: ListOwnershipHistory :many
SELECT * FROM ownership_history
WHERE company_id = ?
ORDER BY as_of;
