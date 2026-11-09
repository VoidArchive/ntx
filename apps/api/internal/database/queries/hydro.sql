-- name: UpsertHydroPlant :exec
INSERT INTO hydro_plants (symbol, plant, river, basin, district, mw, lat, lon, status, source_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(symbol, plant) DO UPDATE SET
  river = excluded.river,
  basin = excluded.basin,
  district = excluded.district,
  mw = excluded.mw,
  lat = excluded.lat,
  lon = excluded.lon,
  status = excluded.status,
  source_url = excluded.source_url;

-- Every plant with its company's latest close, latest quarterly EPS (for PE, divided
-- in Go as the screener does) and stored rating. LEFT joins: a plant lists without them.
-- name: ListHydroPlants :many
SELECT hp.*, c.name AS company_name,
  p.close_price, p.business_date AS close_date,
  f.eps, f.fiscal_year, f.quarter,
  r.letter AS rating
FROM hydro_plants hp
LEFT JOIN companies c ON c.symbol = hp.symbol
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
LEFT JOIN company_ratings r ON r.company_id = c.id AND r.version = sqlc.arg('version')
ORDER BY hp.basin, hp.symbol, hp.plant;

-- name: ListHydroPlantsBySymbol :many
SELECT * FROM hydro_plants WHERE symbol = ? ORDER BY plant;

-- name: UpsertRainReading :exec
INSERT INTO hydro_readings (plant_id, day, rain_mm, rain_source)
VALUES (?, ?, ?, ?)
ON CONFLICT(plant_id, day) DO UPDATE SET
  rain_mm = excluded.rain_mm,
  rain_source = excluded.rain_source,
  fetched_at = CURRENT_TIMESTAMP;

-- name: UpsertDischargeReading :exec
INSERT INTO hydro_readings (plant_id, day, discharge_m3s, discharge_source)
VALUES (?, ?, ?, ?)
ON CONFLICT(plant_id, day) DO UPDATE SET
  discharge_m3s = excluded.discharge_m3s,
  discharge_source = excluded.discharge_source,
  fetched_at = CURRENT_TIMESTAMP;

-- name: CountDischargeReadings :one
SELECT COUNT(*) FROM hydro_readings WHERE plant_id = ? AND discharge_m3s IS NOT NULL;

-- Newest day per plant, the reading the state is computed from. The GROUP BY
-- walks the (plant_id, day) index once; a correlated MAX scanned every row.
-- name: ListLatestReadings :many
SELECT r.* FROM hydro_readings r
JOIN (SELECT plant_id, MAX(day) AS day FROM hydro_readings GROUP BY plant_id) m
  ON m.plant_id = r.plant_id AND m.day = r.day;

-- Nearest-rank 95th percentile of each plant's discharge since floor (ten years
-- ago, so the window stays ten years as rows age): the row ranked ceil(0.95 * n)
-- ascending, written in integer arithmetic as SQLite has no ceil.
-- name: ListDischargeP95 :many
SELECT plant_id, discharge_m3s FROM (
  SELECT plant_id, discharge_m3s,
    ROW_NUMBER() OVER (PARTITION BY plant_id ORDER BY discharge_m3s) AS rn,
    COUNT(*) OVER (PARTITION BY plant_id) AS n
  FROM hydro_readings WHERE discharge_m3s IS NOT NULL AND day >= sqlc.arg(floor)
)
WHERE rn = (n * 95 + 99) / 100;

-- A symbol DoED spells differently from NEPSE would list a plant with no company.
-- name: ListHydroPlantsWithoutCompany :many
SELECT hp.symbol, hp.plant FROM hydro_plants hp
LEFT JOIN companies c ON c.symbol = hp.symbol
WHERE c.id IS NULL;

-- name: ListRecentReadings :many
SELECT * FROM hydro_readings WHERE plant_id = ? ORDER BY day DESC LIMIT ?;

-- name: GetLastWeatherRun :one
SELECT CAST(COALESCE(MAX(fetched_at), '') AS TEXT) AS fetched_at FROM hydro_readings;

-- name: ListAllHydroPlants :many
SELECT * FROM hydro_plants ORDER BY id;
