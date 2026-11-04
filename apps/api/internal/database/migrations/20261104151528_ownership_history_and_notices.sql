-- +goose Up
-- +goose StatementBegin
-- One row per change in a company's shareholding, so the page can show how it moved.
CREATE TABLE ownership_history (
    company_id BIGINT NOT NULL REFERENCES companies(id),
    as_of TEXT NOT NULL,
    listed_shares BIGINT,
    public_shares BIGINT,
    public_percent REAL,
    promoter_shares BIGINT,
    promoter_percent REAL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, as_of)
);

-- The dividend notice as NEPSE published it, kept verbatim as the figure's provenance (ADR 0002).
ALTER TABLE corporate_actions ADD COLUMN notice_headline TEXT;
ALTER TABLE corporate_actions ADD COLUMN notice_body TEXT;
ALTER TABLE corporate_actions ADD COLUMN notice_url TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE corporate_actions DROP COLUMN notice_url;
ALTER TABLE corporate_actions DROP COLUMN notice_body;
ALTER TABLE corporate_actions DROP COLUMN notice_headline;
DROP TABLE ownership_history;
-- +goose StatementEnd
