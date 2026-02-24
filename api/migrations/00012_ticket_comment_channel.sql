-- +goose Up

CREATE TYPE comment_channel AS ENUM ('email', 'sms', 'voice', 'internal', 'web');
ALTER TABLE ticket_comments ADD COLUMN channel comment_channel NOT NULL DEFAULT 'email';

-- +goose Down

ALTER TABLE ticket_comments DROP COLUMN channel;
DROP TYPE comment_channel;
