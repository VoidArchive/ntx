-- +goose Up
-- +goose StatementBegin
-- The owner's written break conditions per holding (GLOSSARY: Gate). A metric
-- gate names a metric, a direction and a threshold over consecutive periods; the
-- prior directions are ratchets against the previous period's own figure. An event
-- gate is text ticked at each review.
CREATE TABLE gates (
    id INTEGER PRIMARY KEY,
    holding_id INTEGER NOT NULL REFERENCES holdings(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    metric TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT '',
    threshold REAL NOT NULL DEFAULT 0,
    periods INTEGER NOT NULL DEFAULT 1,
    text TEXT NOT NULL DEFAULT '',
    checked_on TEXT NOT NULL DEFAULT ''
);
-- +goose StatementEnd

-- +goose StatementBegin
-- One editable row of portfolio rules; the defaults are the owner's note.
CREATE TABLE rules (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    soft_cap_percent REAL NOT NULL DEFAULT 20,
    hard_cap_percent REAL NOT NULL DEFAULT 25,
    sector_cap_percent REAL NOT NULL DEFAULT 40,
    max_positions INTEGER NOT NULL DEFAULT 8,
    min_entry_rs REAL NOT NULL DEFAULT 15000,
    one_hydro_per_basin INTEGER NOT NULL DEFAULT 1,
    candidate_min_rating TEXT NOT NULL DEFAULT 'B',
    candidate_dividend_record INTEGER NOT NULL DEFAULT 1,
    monthly_budget_rs REAL NOT NULL DEFAULT 12000,
    multiplier_below_median REAL NOT NULL DEFAULT 1.5,
    multiplier_bottom_quarter REAL NOT NULL DEFAULT 2.0
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO rules (id) VALUES (1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rules;
DROP TABLE gates;
-- +goose StatementEnd
