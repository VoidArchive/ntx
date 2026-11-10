-- +goose Up
-- +goose StatementBegin
-- The Dividend record check's verdict beside the letter, so the owner's Candidates
-- filter reads a stored row instead of building every story per request.
ALTER TABLE company_ratings ADD COLUMN dividend_record TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE company_ratings DROP COLUMN dividend_record;
-- +goose StatementEnd
