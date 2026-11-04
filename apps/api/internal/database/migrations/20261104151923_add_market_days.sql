-- +goose Up
-- +goose StatementBegin
-- One row per trading session, written by the post-close job (see GLOSSARY: Market day).
CREATE TABLE market_days (
    business_date TEXT PRIMARY KEY,
    index_close REAL NOT NULL,
    index_change REAL NOT NULL,
    index_change_percent REAL NOT NULL,
    turnover REAL NOT NULL,
    traded_shares BIGINT NOT NULL,
    transactions BIGINT NOT NULL,
    scrips_traded INTEGER NOT NULL,
    market_cap REAL NOT NULL,
    advances INTEGER NOT NULL,
    declines INTEGER NOT NULL,
    unchanged INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE market_days;
-- +goose StatementEnd
