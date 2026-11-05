-- +goose Up
-- +goose StatementBegin
-- One row per company, written daily by the worker from the story engine (GLOSSARY: Rating).
-- letter is '' when fewer than four checks could be evaluated.
CREATE TABLE company_ratings (
    company_id BIGINT PRIMARY KEY REFERENCES companies(id),
    letter TEXT NOT NULL,
    holds INTEGER NOT NULL,
    evaluated INTEGER NOT NULL,
    computed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE company_ratings;
-- +goose StatementEnd
