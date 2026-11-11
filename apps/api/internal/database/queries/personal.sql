-- The owner's personal tables as one file (issue #62). Rows are exported and
-- imported with their ids so holding_id references survive; bank rows travel by
-- symbol, since company ids are the exchange's.

-- name: ExportHoldings :many
-- The table as stored, unlike ListHoldings which prices each row.
SELECT * FROM holdings ORDER BY id;

-- name: ListBankMetrics :many
SELECT b.*, c.symbol FROM bank_metrics b JOIN companies c ON c.id = b.company_id ORDER BY b.id;

-- name: DeleteAllHoldings :exec
DELETE FROM holdings;

-- name: DeleteAllBankMetrics :exec
DELETE FROM bank_metrics;

-- name: ImportHolding :exec
INSERT INTO holdings (id, symbol, status, target_shares, allocation_percent, carry_rs, drawdown_gate, thesis, frozen_reason, review_date, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ImportTransaction :exec
INSERT INTO transactions (id, holding_id, day, kind, quantity, price) VALUES (?, ?, ?, ?, ?, ?);

-- name: ImportNote :exec
INSERT INTO holding_notes (id, holding_id, day, text) VALUES (?, ?, ?, ?);

-- name: ImportGate :exec
INSERT INTO gates (id, holding_id, kind, metric, direction, threshold, periods, text, checked_on)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ImportBankMetric :exec
INSERT INTO bank_metrics (id, company_id, fiscal_year, quarter, npl_percent, car_percent, loan_growth_percent, distributable_profit, published_date, source_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
