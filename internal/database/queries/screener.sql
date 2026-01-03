-- name: GetScreenerData :many
-- Returns all data needed for screening: company + latest price + fundamentals
-- Filtering and sorting will be done in Go for flexibility
SELECT
    c.symbol,
    c.name,
    c.sector,
    c.description,
    c.logo_url,
    p.date as price_date,
    p.open,
    p.high,
    p.low,
    p.close,
    p.previous_close,
    p.volume,
    p.turnover,
    p.week_52_high,
    p.week_52_low,
    f.pe,
    f.pb,
    f.eps,
    f.book_value,
    f.market_cap,
    f.dividend_yield,
    f.roe,
    f.shares_outstanding
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
LEFT JOIN fundamentals f ON c.symbol = f.symbol
ORDER BY c.symbol;

-- name: GetScreenerDataBySector :many
SELECT
    c.symbol,
    c.name,
    c.sector,
    c.description,
    c.logo_url,
    p.date as price_date,
    p.open,
    p.high,
    p.low,
    p.close,
    p.previous_close,
    p.volume,
    p.turnover,
    p.week_52_high,
    p.week_52_low,
    f.pe,
    f.pb,
    f.eps,
    f.book_value,
    f.market_cap,
    f.dividend_yield,
    f.roe,
    f.shares_outstanding
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
LEFT JOIN fundamentals f ON c.symbol = f.symbol
WHERE c.sector = ?
ORDER BY c.symbol;
