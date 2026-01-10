-- +goose Up
CREATE TABLE ownership (
    company_id BIGINT PRIMARY KEY REFERENCES companies(id),
    listed_shares BIGINT,
    public_shares BIGINT,
    public_percent REAL,
    promoter_shares BIGINT,
    promoter_percent REAL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS ownership;
