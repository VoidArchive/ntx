-- name: UpsertCompanyRating :exec
INSERT INTO company_ratings (company_id, letter, holds, evaluated, computed_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(company_id) DO UPDATE SET
  letter = excluded.letter,
  holds = excluded.holds,
  evaluated = excluded.evaluated,
  computed_at = excluded.computed_at;

-- One row per listed company with the inputs of the screener columns; ratios
-- are divided in Go so a zero never reaches SQL. Every join is LEFT so a
-- company without a filing, a close or a rating still lists.
-- shortcut: the "latest quarterly row" subquery repeats story.sql and sector.sql; fold all three into one view when the quarter ordering next changes.
-- name: ListScreener :many
SELECT c.symbol, c.name, c.sector,
  o.listed_shares,
  p.close_price, p.turnover,
  f.eps, f.book_value,
  ca.cash_dividend,
  r.letter AS rating,
  f.fiscal_year, f.quarter, f.fiscal_year_np, p.business_date
FROM companies c
LEFT JOIN ownership o ON o.company_id = c.id
LEFT JOIN prices p ON p.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.close_price IS NOT NULL
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
LEFT JOIN company_ratings r ON r.company_id = c.id
WHERE c.status = 'A'
ORDER BY c.symbol;

-- name: CountCompanyRatings :one
SELECT COUNT(*) FROM company_ratings;
