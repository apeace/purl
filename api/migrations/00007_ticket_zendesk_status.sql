-- +goose Up
ALTER TABLE tickets ADD COLUMN zendesk_status zendesk_status_category;

-- +goose Down
ALTER TABLE tickets DROP COLUMN zendesk_status;
