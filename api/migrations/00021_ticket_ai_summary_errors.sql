-- +goose Up
ALTER TABLE tickets ADD COLUMN ai_summary_error_count INT NOT NULL DEFAULT 0;
ALTER TABLE tickets ADD COLUMN ai_summary_last_error TEXT;

-- +goose Down
ALTER TABLE tickets DROP COLUMN ai_summary_last_error;
ALTER TABLE tickets DROP COLUMN ai_summary_error_count;
