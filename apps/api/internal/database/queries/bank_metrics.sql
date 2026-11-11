-- name: UpsertBankMetric :one
INSERT INTO bank_metrics (company_id, fiscal_year, quarter, npl_percent, car_percent, loan_growth_percent, distributable_profit, published_date, source_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(company_id, fiscal_year, quarter) DO UPDATE SET
  npl_percent = excluded.npl_percent,
  car_percent = excluded.car_percent,
  loan_growth_percent = excluded.loan_growth_percent,
  distributable_profit = excluded.distributable_profit,
  published_date = excluded.published_date,
  source_url = excluded.source_url
RETURNING *;

-- name: ListBankMetricsByCompany :many
SELECT * FROM bank_metrics WHERE company_id = ? ORDER BY fiscal_year, quarter;
