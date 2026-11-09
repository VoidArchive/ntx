-- name: UpsertCompanyRating :exec
INSERT INTO company_ratings (company_id, letter, holds, evaluated, total, version, computed_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(company_id) DO UPDATE SET
  letter = excluded.letter,
  holds = excluded.holds,
  evaluated = excluded.evaluated,
  total = excluded.total,
  version = excluded.version,
  computed_at = excluded.computed_at;

-- One row per listed company with the inputs of the screener columns; ratios
-- are divided in Go so a zero never reaches SQL. Every join is LEFT so a
-- company without a filing, a close or a rating still lists.
-- shortcut: the "latest quarterly row" subquery repeats story.sql and sector.sql; fold all three into one view when the quarter ordering next changes.
-- name: ListScreener :many
SELECT c.id, c.symbol, c.name, c.sector,
  o.listed_shares,
  p.close_price, p.turnover,
  prev.close_price AS prev_close,
  f.eps, f.book_value,
  ca.cash_dividend,
  r.letter AS rating, r.holds, r.evaluated, r.total,
  f.fiscal_year, f.quarter, f.fiscal_year_np, p.business_date
FROM companies c
LEFT JOIN ownership o ON o.company_id = c.id
LEFT JOIN prices p ON p.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.close_price IS NOT NULL
  ORDER BY p2.business_date DESC LIMIT 1
)
LEFT JOIN prices prev ON prev.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.business_date < p.business_date AND p2.close_price IS NOT NULL
  ORDER BY p2.business_date DESC LIMIT 1
)
LEFT JOIN fundamentals f ON f.id = (
  SELECT f2.id FROM fundamentals f2
  WHERE f2.company_id = c.id AND f2.quarter != ''
  ORDER BY f2.fiscal_year DESC,
    CASE f2.quarter WHEN 'Fourth Quarter' THEN 4 WHEN 'Third Quarter' THEN 3 WHEN 'Second Quarter' THEN 2 ELSE 1 END DESC
  LIMIT 1
)
LEFT JOIN corporate_actions ca ON ca.id = (
  SELECT ca2.id FROM corporate_actions ca2
  WHERE ca2.company_id = c.id
  ORDER BY ca2.fiscal_year DESC, ca2.submitted_date DESC
  LIMIT 1
)
LEFT JOIN company_ratings r ON r.company_id = c.id AND r.version = sqlc.arg('version')
WHERE c.status = 'A'
ORDER BY c.symbol;

-- Listed companies without a rating from the current rule set: every one after a
-- deploy that changed the rules, the remainder after a sync that stopped part-way.
-- name: CountUnratedCompanies :one
SELECT COUNT(*) FROM companies c
WHERE c.status = 'A'
  AND NOT EXISTS (SELECT 1 FROM company_ratings r WHERE r.company_id = c.id AND r.version = ?);
