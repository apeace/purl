-- +goose Up
ALTER TABLE tickets ADD COLUMN resolved_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE tickets DROP COLUMN resolved_at;
