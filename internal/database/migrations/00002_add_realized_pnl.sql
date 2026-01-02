-- +goose Up
ALTER TABLE holdings ADD COLUMN realized_pnl_paisa INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE holdings DROP COLUMN realized_pnl_paisa;
