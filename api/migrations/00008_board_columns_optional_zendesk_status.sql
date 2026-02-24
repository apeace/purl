-- +goose Up

-- zendesk_status is only meaningful for the default board's columns.
-- Custom board columns don't map to Zendesk statuses, so make it optional.
ALTER TABLE board_columns ALTER COLUMN zendesk_status DROP NOT NULL;

-- Replace the blanket unique constraint with a partial index so uniqueness
-- is only enforced for rows where zendesk_status is actually set.
ALTER TABLE board_columns DROP CONSTRAINT board_columns_board_id_zendesk_status_key;
CREATE UNIQUE INDEX board_columns_board_id_zendesk_status_key
    ON board_columns (board_id, zendesk_status)
    WHERE zendesk_status IS NOT NULL;

-- +goose Down

DROP INDEX board_columns_board_id_zendesk_status_key;
ALTER TABLE board_columns ADD CONSTRAINT board_columns_board_id_zendesk_status_key
    UNIQUE (board_id, zendesk_status);
ALTER TABLE board_columns ALTER COLUMN zendesk_status SET NOT NULL;
