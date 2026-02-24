-- +goose Up

ALTER TABLE comments RENAME TO ticket_comments;

ALTER TABLE tickets DROP COLUMN status;
ALTER TABLE tickets DROP COLUMN priority;

DROP TYPE ticket_status;
DROP TYPE ticket_priority;

-- +goose Down

CREATE TYPE ticket_status   AS ENUM ('open', 'in_progress', 'resolved', 'closed');
CREATE TYPE ticket_priority AS ENUM ('low', 'medium', 'high', 'urgent');

ALTER TABLE tickets ADD COLUMN status   ticket_status   NOT NULL DEFAULT 'open';
ALTER TABLE tickets ADD COLUMN priority ticket_priority NOT NULL DEFAULT 'medium';

ALTER TABLE ticket_comments RENAME TO comments;
