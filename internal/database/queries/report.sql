-- name: InsertReport :exec
INSERT INTO reports (symbol, type, fiscal_year, quarter, revenue, net_income, eps, book_value, npl_ratio, published_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(symbol, type, fiscal_year, quarter) DO UPDATE SET
    revenue = excluded.revenue,
    net_income = excluded.net_income,
    eps = excluded.eps,
    book_value = excluded.book_value,
    npl_ratio = excluded.npl_ratio,
    published_at = excluded.published_at;

-- name: GetReports :many
SELECT * FROM reports
WHERE symbol = ?
ORDER BY fiscal_year DESC, quarter DESC
LIMIT ?;

-- name: GetReportsByType :many
SELECT * FROM reports
WHERE symbol = ? AND type = ?
ORDER BY fiscal_year DESC, quarter DESC
LIMIT ?;

-- name: GetLatestReport :one
SELECT * FROM reports
WHERE symbol = ?
ORDER BY fiscal_year DESC, quarter DESC
LIMIT 1;
