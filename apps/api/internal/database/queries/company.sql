-- name: UpsertCompany :exec
INSERT INTO companies (id, name, symbol, status, email, website, sector, instrument_type, updated_at)
VALUES (?,?,?,?,?,?,?,?, CURRENT_TIMESTAMP)
ON CONFLICT(symbol) DO UPDATE SET
  name = excluded.name,
  status = excluded.status,
  email = excluded.email,
  website = excluded.website,
  sector = excluded.sector,
  instrument_type = excluded.instrument_type,
  updated_at = CURRENT_TIMESTAMP;

-- name: GetCompany :one 
SELECT * FROM companies WHERE symbol = ?;

-- name: ListCompanies :many
SELECT c.*, o.listed_shares
FROM companies c
LEFT JOIN ownership o ON c.id = o.company_id
ORDER by c.symbol LIMIT ? OFFSET ?;

-- name: ListCompaniesBySector :many
SELECT * FROM companies
WHERE sector = ? AND (symbol LIKE ? OR name LIKE ?)
ORDER BY symbol
LIMIT ? OFFSET ?;

-- name: SearchCompanies :many
SELECT * FROM companies
WHERE symbol LIKE ? COLLATE NOCASE
OR name LIKE ? COLLATE NOCASE
ORDER BY symbol
LIMIT ? OFFSET ?;

-- name: CountCompanies :one
SELECT COUNT(*) FROM companies;

-- name: CountCompaniesBySector :one
SELECT COUNT(*) FROM companies WHERE sector = ?;

-- name: CountCompaniesBySearch :one
SELECT COUNT(*) FROM companies
WHERE symbol LIKE ? COLLATE NOCASE
OR name LIKE ? COLLATE NOCASE;



