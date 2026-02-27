-- +goose Up
ALTER TABLE tickets ADD COLUMN ai_title TEXT;
ALTER TABLE tickets ADD COLUMN ai_temperature SMALLINT;

-- +goose Down
ALTER TABLE tickets DROP COLUMN ai_temperature;
ALTER TABLE tickets DROP COLUMN ai_title;
