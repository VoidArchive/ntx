-- name: GetSectorSummary :many
-- Returns aggregated sector-level data for ListSectors endpoint
SELECT
    c.sector,
    COUNT(*) as stock_count,
    COALESCE(SUM(p.turnover), 0) as turnover,
    COALESCE(SUM(CASE WHEN p.close > p.previous_close THEN 1 ELSE 0 END), 0) as gainers,
    COALESCE(SUM(CASE WHEN p.close < p.previous_close THEN 1 ELSE 0 END), 0) as losers
FROM companies c
LEFT JOIN (
    SELECT p1.*
    FROM prices p1
    INNER JOIN (
        SELECT symbol, MAX(date) as max_date
        FROM prices
        GROUP BY symbol
    ) p2 ON p1.symbol = p2.symbol AND p1.date = p2.max_date
) p ON c.symbol = p.symbol
GROUP BY c.sector
ORDER BY c.sector;
