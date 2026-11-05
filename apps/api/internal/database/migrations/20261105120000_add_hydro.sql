-- +goose Up
-- +goose StatementBegin
-- One row per plant of a listed hydro company, upserted from data/hydro_plants.csv
-- on worker start (GLOSSARY: Plant; ADR 0004). basin is one of the nine in GLOSSARY.
CREATE TABLE hydro_plants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    plant TEXT NOT NULL,
    river TEXT NOT NULL,
    basin TEXT NOT NULL,
    district TEXT NOT NULL,
    mw REAL NOT NULL,
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    status TEXT NOT NULL,
    source_url TEXT NOT NULL,
    UNIQUE(symbol, plant)
);
-- +goose StatementEnd
-- +goose StatementBegin
-- One row per plant per day (GLOSSARY: Reading). Rain and discharge arrive from two
-- calls, so either half may be NULL until the other lands; forecast days are never rows.
CREATE TABLE hydro_readings (
    plant_id INTEGER NOT NULL REFERENCES hydro_plants(id),
    day TEXT NOT NULL,
    rain_mm REAL,
    discharge_m3s REAL,
    rain_source TEXT,
    discharge_source TEXT,
    fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plant_id, day)
);
-- +goose StatementEnd
-- +goose StatementBegin
-- The discharge percentile sorts each plant's ten years of rows.
CREATE INDEX hydro_readings_discharge ON hydro_readings(plant_id, discharge_m3s);
-- +goose StatementEnd
-- +goose StatementBegin
-- Every page read asks for the newest fetched_at to know whether the weather job has run.
CREATE INDEX hydro_readings_fetched ON hydro_readings(fetched_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE hydro_readings;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE hydro_plants;
-- +goose StatementEnd
