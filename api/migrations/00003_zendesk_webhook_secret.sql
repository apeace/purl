-- +goose Up
ALTER TABLE organizations ADD COLUMN zendesk_webhook_secret TEXT;
-- +goose Down
ALTER TABLE organizations DROP COLUMN zendesk_webhook_secret;
