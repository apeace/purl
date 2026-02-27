-- +goose Up
ALTER TABLE tickets ADD COLUMN received_at TIMESTAMPTZ;
ALTER TABLE ticket_comments ADD COLUMN received_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE ticket_comments DROP COLUMN received_at;
ALTER TABLE tickets DROP COLUMN received_at;
