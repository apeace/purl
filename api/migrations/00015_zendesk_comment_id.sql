-- +goose Up

-- Zendesk comment ID for ticket_comments, used to correlate webhook payloads
-- with existing rows. NULL for comments created directly in Purl (not yet synced
-- back from Zendesk).
ALTER TABLE ticket_comments ADD COLUMN zendesk_comment_id BIGINT;
ALTER TABLE ticket_comments ADD CONSTRAINT ticket_comments_ticket_id_zendesk_comment_id_unique
    UNIQUE (ticket_id, zendesk_comment_id);

-- +goose Down

ALTER TABLE ticket_comments DROP CONSTRAINT ticket_comments_ticket_id_zendesk_comment_id_unique;
ALTER TABLE ticket_comments DROP COLUMN zendesk_comment_id;
