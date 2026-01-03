-- name: UpsertCompany :exec
INSERT INTO companies (symbol, name, sector, description, logo_url, last_synced)
VALUES (?, ?, ?, ?, ?, datetime('now'))
ON CONFLICT(symbol) DO UPDATE SET
    name = excluded.name,
    sector = excluded.sector,
    description = excluded.description,
    logo_url = excluded.logo_url,
    last_synced = excluded.last_synced;

-- name: UpsertCompanyBasic :exec
-- Updates company info without overwriting description/logo
INSERT INTO companies (symbol, name, sector, description, logo_url, last_synced)
VALUES (?, ?, ?, '', '', datetime('now'))
ON CONFLICT(symbol) DO UPDATE SET
    name = excluded.name,
    sector = excluded.sector,
    last_synced = excluded.last_synced;

-- name: GetCompany :one
SELECT * FROM companies WHERE symbol = ? LIMIT 1;

-- name: ListCompanies :many
SELECT * FROM companies ORDER BY symbol;

-- name: ListCompaniesBySector :many
SELECT * FROM companies WHERE sector = ? ORDER BY symbol;

-- name: SearchCompanies :many
SELECT * FROM companies
WHERE symbol LIKE ? OR name LIKE ?
ORDER BY symbol
LIMIT ?;

-- name: CountCompanies :one
SELECT COUNT(*) FROM companies;

-- name: UpdateCompanyDescription :exec
UPDATE companies SET description = ? WHERE symbol = ?;
