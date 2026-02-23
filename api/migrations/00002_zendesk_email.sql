-- +goose Up
ALTER TABLE organizations ADD COLUMN zendesk_email TEXT;

-- +goose Down
ALTER TABLE organizations DROP COLUMN zendesk_email;
