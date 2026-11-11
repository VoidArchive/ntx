-- +goose Up
-- +goose StatementBegin
-- The owner's lender ceilings, the PE-to-sector rule and the goal (issue #67). The
-- CAR floor of 12 sits above NRB's 11% minimum so a thinly capitalised bank is
-- turned away; the owner can move every one of these.
ALTER TABLE rules ADD COLUMN goal_rs REAL NOT NULL DEFAULT 500000;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE rules ADD COLUMN fund_sip_rs REAL NOT NULL DEFAULT 4000;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE rules ADD COLUMN npl_ceiling_percent REAL NOT NULL DEFAULT 5;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE rules ADD COLUMN car_floor_percent REAL NOT NULL DEFAULT 12;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE rules ADD COLUMN distributable_above_zero INTEGER NOT NULL DEFAULT 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE rules ADD COLUMN max_pe_to_sector_median REAL NOT NULL DEFAULT 1.25;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rules DROP COLUMN goal_rs;
ALTER TABLE rules DROP COLUMN fund_sip_rs;
ALTER TABLE rules DROP COLUMN npl_ceiling_percent;
ALTER TABLE rules DROP COLUMN car_floor_percent;
ALTER TABLE rules DROP COLUMN distributable_above_zero;
ALTER TABLE rules DROP COLUMN max_pe_to_sector_median;
-- +goose StatementEnd
