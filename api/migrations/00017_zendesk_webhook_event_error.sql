-- +goose Up

ALTER TABLE zendesk_webhook_events ADD COLUMN last_error TEXT;

-- +goose Down

ALTER TABLE zendesk_webhook_events DROP COLUMN last_error;
