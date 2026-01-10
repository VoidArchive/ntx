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
