-- Sector distributions for the story strips: one value per company in the
-- sector, from its latest quarterly row and its latest close. Divisions by
-- zero are filtered so every value is a number.
-- shortcut: the "latest quarterly row" subquery repeats here and in sector.sql; replace it with a latest_quarterly_fundamentals view when the quarter ordering next changes.

-- Every quarterly row of every listed company with the company's latest close,
-- for the story engine to price earnings on a trailing four-quarter basis; one
-- sector or, with a NULL sector, all of them.
-- name: ListEarningsRows :many
SELECT sqlc.embed(f), p.close_price AS close
FROM companies c
JOIN fundamentals f ON f.company_id = c.id AND f.quarter != ''
LEFT JOIN prices p ON p.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.close_price IS NOT NULL
  ORDER BY p2.business_date DESC LIMIT 1
)
WHERE c.status = 'A' AND (sqlc.narg('sector') IS NULL OR c.sector = sqlc.narg('sector'))
ORDER BY c.id, f.fiscal_year, f.quarter;

-- name: ListSectorPB :many
SELECT CAST(p.close_price / f.book_value AS REAL) AS pb
FROM companies c
JOIN fundamentals f ON f.id = (
  SELECT f2.id FROM fundamentals f2
  WHERE f2.company_id = c.id AND f2.quarter != ''
  ORDER BY f2.fiscal_year DESC,
    CASE f2.quarter WHEN 'Fourth Quarter' THEN 4 WHEN 'Third Quarter' THEN 3 WHEN 'Second Quarter' THEN 2 ELSE 1 END DESC
  LIMIT 1
)
JOIN prices p ON p.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.close_price IS NOT NULL
  ORDER BY p2.business_date DESC LIMIT 1
)
WHERE c.status = 'A' AND c.sector = ? AND f.book_value != 0;

-- name: ListSectorEPS :many
SELECT CAST(f.eps AS REAL) AS eps
FROM companies c
JOIN fundamentals f ON f.id = (
  SELECT f2.id FROM fundamentals f2
  WHERE f2.company_id = c.id AND f2.quarter != ''
  ORDER BY f2.fiscal_year DESC,
    CASE f2.quarter WHEN 'Fourth Quarter' THEN 4 WHEN 'Third Quarter' THEN 3 WHEN 'Second Quarter' THEN 2 ELSE 1 END DESC
  LIMIT 1
)
WHERE c.status = 'A' AND c.sector = ? AND f.eps IS NOT NULL;

-- Cash dividend is a percentage of the Rs 100 par, so percent yield = cash% * 100 / close.
-- name: ListSectorYield :many
SELECT CAST(COALESCE(ca.cash_dividend, 0) * 100.0 / p.close_price AS REAL) AS yield
FROM companies c
JOIN corporate_actions ca ON ca.id = (
  SELECT ca2.id FROM corporate_actions ca2
  WHERE ca2.company_id = c.id
  ORDER BY ca2.fiscal_year DESC, ca2.submitted_date DESC
  LIMIT 1
)
JOIN prices p ON p.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.close_price IS NOT NULL
  ORDER BY p2.business_date DESC LIMIT 1
)
WHERE c.status = 'A' AND c.sector = ? AND p.close_price != 0;

-- name: GetLatestCorporateActionByCompany :one
SELECT * FROM corporate_actions
WHERE company_id = ?
ORDER BY fiscal_year DESC, submitted_date DESC
LIMIT 1;

-- name: ListSectorPeers :many
SELECT c.id, c.symbol, c.name, CAST(p.close_price AS REAL) AS close, f.eps, f.book_value
FROM companies c
JOIN fundamentals f ON f.id = (
  SELECT f2.id FROM fundamentals f2
  WHERE f2.company_id = c.id AND f2.quarter != ''
  ORDER BY f2.fiscal_year DESC,
    CASE f2.quarter WHEN 'Fourth Quarter' THEN 4 WHEN 'Third Quarter' THEN 3 WHEN 'Second Quarter' THEN 2 ELSE 1 END DESC
  LIMIT 1
)
JOIN prices p ON p.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.close_price IS NOT NULL
  ORDER BY p2.business_date DESC LIMIT 1
)
WHERE c.status = 'A' AND c.sector = ?;
