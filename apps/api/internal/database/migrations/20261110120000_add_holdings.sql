-- +goose Up
-- +goose StatementBegin
-- The owner's holdings and watchlist (GLOSSARY: Holding). A watched company is a
-- row with no transactions. Owner-only state behind one token (ADR 0006).
CREATE TABLE holdings (
    id INTEGER PRIMARY KEY,
    symbol TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'hold',
    target_shares INTEGER NOT NULL DEFAULT 0,
    allocation_percent REAL NOT NULL DEFAULT 0,
    carry_rs REAL NOT NULL DEFAULT 0,
    drawdown_gate INTEGER NOT NULL DEFAULT 0,
    thesis TEXT NOT NULL DEFAULT '',
    frozen_reason TEXT NOT NULL DEFAULT '',
    review_date TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Price is per share as TMS shows it, fees included; a bonus is a zero-price lot.
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY,
    holding_id INTEGER NOT NULL REFERENCES holdings(id) ON DELETE CASCADE,
    day TEXT NOT NULL,
    kind TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price REAL NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE holding_notes (
    id INTEGER PRIMARY KEY,
    holding_id INTEGER NOT NULL REFERENCES holdings(id) ON DELETE CASCADE,
    day TEXT NOT NULL,
    text TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE holding_notes;
DROP TABLE transactions;
DROP TABLE holdings;
-- +goose StatementEnd
