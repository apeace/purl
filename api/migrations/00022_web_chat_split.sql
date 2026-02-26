-- +goose Up

-- Add 'chat' channel for individual web chat sub-messages (split from a single
-- Zendesk comment into one row per transcript line).
ALTER TYPE comment_channel ADD VALUE IF NOT EXISTS 'chat';

-- Track ordering of sub-messages within a single Zendesk comment.
-- Regular comments always have 0 (the default).
ALTER TABLE ticket_comments ADD COLUMN zendesk_sub_index SMALLINT NOT NULL DEFAULT 0;

-- Store the parsed speaker name from web chat transcripts (e.g. "Web User", "Bot").
-- Overrides the JOIN-derived author_name in the API response.
ALTER TABLE ticket_comments ADD COLUMN author_display_name TEXT;

-- Replace the unique constraint to allow multiple rows per zendesk_comment_id
-- (one per sub-message). NULL zendesk_comment_id rows (Purl-native comments)
-- are always distinct under PostgreSQL NULL semantics.
ALTER TABLE ticket_comments DROP CONSTRAINT ticket_comments_ticket_id_zendesk_comment_id_unique;
ALTER TABLE ticket_comments ADD CONSTRAINT ticket_comments_ticket_id_zendesk_comment_id_sub_index_unique
    UNIQUE (ticket_id, zendesk_comment_id, zendesk_sub_index);

-- +goose Down

ALTER TABLE ticket_comments DROP CONSTRAINT ticket_comments_ticket_id_zendesk_comment_id_sub_index_unique;
ALTER TABLE ticket_comments ADD CONSTRAINT ticket_comments_ticket_id_zendesk_comment_id_unique
    UNIQUE (ticket_id, zendesk_comment_id);
ALTER TABLE ticket_comments DROP COLUMN author_display_name;
ALTER TABLE ticket_comments DROP COLUMN zendesk_sub_index;
-- Note: PostgreSQL does not support removing enum values, so 'chat' remains after rollback.
