-- +goose Up

-- Add 52-week high/low to prices table for better performance
ALTER TABLE prices ADD COLUMN week_52_high REAL;
ALTER TABLE prices ADD COLUMN week_52_low REAL;

-- +goose Down

ALTER TABLE prices DROP COLUMN week_52_low;
ALTER TABLE prices DROP COLUMN week_52_high;
