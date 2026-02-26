-- +goose Up
ALTER TABLE tickets ADD COLUMN ai_summary TEXT;
ALTER TABLE tickets ADD COLUMN ai_summary_stale BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE tickets DROP COLUMN ai_summary_stale;
ALTER TABLE tickets DROP COLUMN ai_summary;
