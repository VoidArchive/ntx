-- name: UpsertCompany :exec
INSERT INTO companies (id, name, symbol, status, email, website, sector, instrument_type, updated_at)
VALUES (?,?,?,?,?,?,?,?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
  name = excluded.name,
  symbol = excluded.symbol,
  status = excluded.status,
  email = excluded.email,
  website = excluded.website,
  sector = excluded.sector,
  instrument_type = excluded.instrument_type,
  updated_at = CURRENT_TIMESTAMP;

-- Companies that stop arriving in the NEPSE feed keep their row (and history) but leave every list.
-- name: MarkCompaniesDelistedExcept :exec
UPDATE companies SET status = 'D', updated_at = CURRENT_TIMESTAMP
WHERE status != 'D' AND id NOT IN (sqlc.slice('ids'));

-- name: GetCompany :one
SELECT * FROM companies WHERE symbol = ?;

-- name: ListCompanies :many
SELECT c.*, o.listed_shares
FROM companies c
LEFT JOIN ownership o ON c.id = o.company_id
WHERE c.status = 'A'
ORDER by c.symbol LIMIT ? OFFSET ?;

-- name: ListCompaniesBySector :many
SELECT * FROM companies
WHERE status = 'A' AND sector = ? AND (symbol LIKE ? OR name LIKE ?)
ORDER BY symbol
LIMIT ? OFFSET ?;

-- name: SearchCompanies :many
SELECT * FROM companies
WHERE status = 'A' AND (symbol LIKE ? COLLATE NOCASE
OR name LIKE ? COLLATE NOCASE)
ORDER BY symbol
LIMIT ? OFFSET ?;
