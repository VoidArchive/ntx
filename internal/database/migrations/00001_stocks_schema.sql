-- +goose Up
CREATE TABLE stocks (
    symbol TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    sector INTEGER NOT NULL,
    last_synced TEXT NOT NULL
);

-- +goose Down
DROP TABLE stocks;
