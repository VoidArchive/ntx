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
ORDER BY fiscal_year DESC,
  CASE
    WHEN quarter = '' THEN 5
    WHEN quarter = 'Fourth Quarter' THEN 4
    WHEN quarter = 'Third Quarter' THEN 3
    WHEN quarter = 'Second Quarter' THEN 2
    WHEN quarter = 'First Quarter' THEN 1
    ELSE 0
  END DESC
LIMIT 1;

-- name: ListFundamentalsByCompany :many
SELECT * FROM fundamentals
WHERE company_id = ?
ORDER BY fiscal_year DESC,
  CASE
    WHEN quarter = '' THEN 5
    WHEN quarter = 'Fourth Quarter' THEN 4
    WHEN quarter = 'Third Quarter' THEN 3
    WHEN quarter = 'Second Quarter' THEN 2
    WHEN quarter = 'First Quarter' THEN 1
    ELSE 0
  END DESC;

-- name: GetSectorStats :one
SELECT
  COUNT(DISTINCT c.id) as company_count,
  AVG(f.eps) as avg_eps,
  AVG(f.pe_ratio) as avg_pe_ratio,
  AVG(f.book_value) as avg_book_value
FROM companies c
INNER JOIN fundamentals f ON f.company_id = c.id
WHERE c.sector = ?
  AND f.id IN (
    SELECT f2.id FROM fundamentals f2
    WHERE f2.company_id = c.id
    ORDER BY f2.fiscal_year DESC,
      CASE
        WHEN f2.quarter = '' THEN 5
        WHEN f2.quarter = 'Fourth Quarter' THEN 4
        WHEN f2.quarter = 'Third Quarter' THEN 3
        WHEN f2.quarter = 'Second Quarter' THEN 2
        WHEN f2.quarter = 'First Quarter' THEN 1
        ELSE 0
      END DESC
    LIMIT 1
  );
