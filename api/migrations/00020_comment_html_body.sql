-- +goose Up

ALTER TABLE ticket_comments ADD COLUMN html_body TEXT;

-- +goose Down

ALTER TABLE ticket_comments DROP COLUMN html_body;
