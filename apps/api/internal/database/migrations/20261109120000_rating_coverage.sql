-- +goose Up
-- +goose StatementBegin
ALTER TABLE company_ratings ADD COLUMN total INTEGER NOT NULL DEFAULT 0;
ALTER TABLE company_ratings ADD COLUMN version INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE company_ratings DROP COLUMN version;
ALTER TABLE company_ratings DROP COLUMN total;
-- +goose StatementEnd
