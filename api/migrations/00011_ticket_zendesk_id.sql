-- +goose Up

ALTER TABLE tickets ADD COLUMN zendesk_ticket_id BIGINT NOT NULL;
ALTER TABLE tickets ADD CONSTRAINT tickets_zendesk_ticket_id_org_unique UNIQUE (org_id, zendesk_ticket_id);

-- +goose Down

ALTER TABLE tickets DROP CONSTRAINT tickets_zendesk_ticket_id_org_unique;
ALTER TABLE tickets DROP COLUMN zendesk_ticket_id;
