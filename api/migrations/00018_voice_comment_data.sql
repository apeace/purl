-- +goose Up

ALTER TABLE ticket_comments ADD COLUMN call_id              BIGINT;
ALTER TABLE ticket_comments ADD COLUMN recording_url        TEXT;
ALTER TABLE ticket_comments ADD COLUMN transcription_text   TEXT;
ALTER TABLE ticket_comments ADD COLUMN transcription_status TEXT;
ALTER TABLE ticket_comments ADD COLUMN call_duration        INTEGER;
ALTER TABLE ticket_comments ADD COLUMN call_from            TEXT;
ALTER TABLE ticket_comments ADD COLUMN call_to              TEXT;
ALTER TABLE ticket_comments ADD COLUMN answered_by_name     TEXT;
ALTER TABLE ticket_comments ADD COLUMN call_location        TEXT;
ALTER TABLE ticket_comments ADD COLUMN call_started_at      TIMESTAMPTZ;

-- +goose Down

ALTER TABLE ticket_comments DROP COLUMN call_started_at;
ALTER TABLE ticket_comments DROP COLUMN call_location;
ALTER TABLE ticket_comments DROP COLUMN answered_by_name;
ALTER TABLE ticket_comments DROP COLUMN call_to;
ALTER TABLE ticket_comments DROP COLUMN call_from;
ALTER TABLE ticket_comments DROP COLUMN call_duration;
ALTER TABLE ticket_comments DROP COLUMN transcription_status;
ALTER TABLE ticket_comments DROP COLUMN transcription_text;
ALTER TABLE ticket_comments DROP COLUMN recording_url;
ALTER TABLE ticket_comments DROP COLUMN call_id;
