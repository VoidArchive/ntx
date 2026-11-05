-- name: ListSectorCounts :many
SELECT sector, COUNT(*) AS company_count
FROM companies
WHERE status = 'A'
GROUP BY sector
ORDER BY company_count DESC, sector;

-- Stored profit is year-to-date, so a quarter's YTD is comparable with the same quarter's YTD a year earlier.
-- name: GetSectorProfitYoY :one
SELECT
  CAST(COALESCE(SUM(f.profit_amount), 0) AS REAL) AS current_profit,
  CAST(COALESCE(SUM(prev.profit_amount), 0) AS REAL) AS previous_profit
FROM companies c
JOIN fundamentals f ON f.id = (
  SELECT f2.id FROM fundamentals f2
  WHERE f2.company_id = c.id AND f2.quarter != ''
  ORDER BY f2.fiscal_year DESC,
    CASE f2.quarter WHEN 'Fourth Quarter' THEN 4 WHEN 'Third Quarter' THEN 3 WHEN 'Second Quarter' THEN 2 ELSE 1 END DESC
  LIMIT 1
)
JOIN fundamentals prev ON prev.company_id = c.id
  AND prev.quarter = f.quarter
  AND prev.fiscal_year = (CAST(substr(f.fiscal_year, 1, 4) AS INTEGER) - 1) || '-' || substr(f.fiscal_year, 1, 4)
WHERE c.sector = ? AND f.profit_amount IS NOT NULL AND prev.profit_amount IS NOT NULL;

-- Change columns are NULL on backfilled rows, so the change comes from consecutive closes.
-- name: ListDayChanges :many
SELECT
  c.sector, c.symbol, c.name,
  CAST(p.close_price AS REAL) AS close,
  CAST((p.close_price - prev.close_price) * 100.0 / prev.close_price AS REAL) AS change_percent
FROM companies c
JOIN prices p ON p.company_id = c.id AND p.business_date = (SELECT MAX(business_date) FROM prices)
JOIN prices prev ON prev.id = (
  SELECT p2.id FROM prices p2
  WHERE p2.company_id = c.id AND p2.business_date < p.business_date AND p2.close_price IS NOT NULL
  ORDER BY p2.business_date DESC LIMIT 1
)
WHERE p.close_price IS NOT NULL AND prev.close_price != 0
ORDER BY change_percent DESC, c.symbol;

-- The first backfill set every created_at to one day, so every unpublished row reads "seen" on that day.
-- name: ListFiledRecently :many
SELECT
  c.id, c.symbol, c.name, c.sector,
  CAST(date(MAX(COALESCE(f.published_at, f.created_at))) AS TEXT) AS filed_on,
  CAST(COALESCE(date(MAX(f.published_at)) >= date(MAX(COALESCE(f.published_at, f.created_at))), 0) AS INTEGER) AS published
FROM companies c
JOIN fundamentals f ON f.company_id = c.id
WHERE date(COALESCE(f.published_at, f.created_at)) >= date('now', '-30 days')
GROUP BY c.id
ORDER BY filed_on DESC, c.symbol
LIMIT ?;
