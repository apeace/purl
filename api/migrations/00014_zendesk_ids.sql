-- +goose Up

-- Fix agent email uniqueness: was global, should be per-org.
-- Also enables two different orgs to have agents with the same email address.
ALTER TABLE agents DROP CONSTRAINT agents_email_key;
ALTER TABLE agents ADD CONSTRAINT agents_org_id_email_unique UNIQUE (org_id, email);

-- Zendesk user ID for customers and agents, used to correlate webhook payloads
-- with existing rows and enable idempotent upserts.
ALTER TABLE customers ADD COLUMN zendesk_user_id BIGINT;
ALTER TABLE customers ADD CONSTRAINT customers_org_id_zendesk_user_id_unique UNIQUE (org_id, zendesk_user_id);

ALTER TABLE agents ADD COLUMN zendesk_user_id BIGINT;
ALTER TABLE agents ADD CONSTRAINT agents_org_id_zendesk_user_id_unique UNIQUE (org_id, zendesk_user_id);

-- +goose Down

ALTER TABLE customers DROP CONSTRAINT customers_org_id_zendesk_user_id_unique;
ALTER TABLE customers DROP COLUMN zendesk_user_id;

ALTER TABLE agents DROP CONSTRAINT agents_org_id_zendesk_user_id_unique;
ALTER TABLE agents DROP COLUMN zendesk_user_id;

ALTER TABLE agents DROP CONSTRAINT agents_org_id_email_unique;
ALTER TABLE agents ADD CONSTRAINT agents_email_key UNIQUE (email);
