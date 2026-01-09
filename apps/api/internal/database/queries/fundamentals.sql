-- name: UpsertFundamental :exec
INSERT INTO fundamentals (company_id, fiscal_year, quarter, eps, pe_ratio, book_value, paid_up_capital, profit_amount, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(company_id, fiscal_year, quarter) DO UPDATE SET
  eps = excluded.eps,
  pe_ratio = excluded.pe_ratio,
  book_value = excluded.book_value,
  paid_up_capital = excluded.paid_up_capital,
  profit_amount = excluded.profit_amount,
  updated_at = CURRENT_TIMESTAMP;

-- name: GetLatestFundamental :one
SELECT * FROM fundamentals
WHERE company_id = ?
ORDER BY fiscal_year DESC, quarter DESC NULLS FIRST
LIMIT 1;

-- name: ListFundamentalsByCompany :many
SELECT * FROM fundamentals
WHERE company_id = ?
ORDER BY fiscal_year DESC, quarter DESC NULLS FIRST;
