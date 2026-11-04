-- name: UpsertCorporateAction :exec
INSERT INTO corporate_actions (company_id, fiscal_year, bonus_percentage, right_percentage, cash_dividend, submitted_date, notice_headline, notice_body, notice_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(company_id, fiscal_year, submitted_date) DO UPDATE SET
  bonus_percentage = excluded.bonus_percentage,
  right_percentage = excluded.right_percentage,
  cash_dividend = excluded.cash_dividend,
  notice_headline = excluded.notice_headline,
  notice_body = excluded.notice_body,
  notice_url = excluded.notice_url;

-- name: ListCorporateActionsByCompany :many
SELECT * FROM corporate_actions
WHERE company_id = ?
ORDER BY submitted_date DESC;

-- name: GetCorporateActionsBySymbol :many
SELECT ca.* FROM corporate_actions ca
JOIN companies c ON c.id = ca.company_id
WHERE c.symbol = ?
ORDER BY ca.submitted_date DESC;

-- name: GetLatestCorporateAction :one
SELECT ca.* FROM corporate_actions ca
JOIN companies c ON c.id = ca.company_id
WHERE c.symbol = ?
ORDER BY ca.submitted_date DESC
LIMIT 1;
