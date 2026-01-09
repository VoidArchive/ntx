-- +goose Up
CREATE TABLE prices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    business_date TEXT NOT NULL,
    open_price REAL,
    high_price REAL,
    low_price REAL,
    close_price REAL,
    last_traded_price REAL,
    previous_close REAL,
    change_amount REAL,
    change_percent REAL,
    volume BIGINT,
    turnover REAL,
    trades INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, business_date)
);

CREATE INDEX idx_prices_company_id ON prices(company_id);
CREATE INDEX idx_prices_business_date ON prices(business_date);

-- +goose Down
DROP TABLE IF EXISTS prices;
