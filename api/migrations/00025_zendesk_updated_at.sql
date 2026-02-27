-- +goose Up
ALTER TABLE tickets ADD COLUMN zendesk_updated_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE tickets DROP COLUMN zendesk_updated_at;
