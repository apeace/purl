-- +goose Up

CREATE TABLE zendesk_webhook_events (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    org_id      UUID        NOT NULL REFERENCES organizations(id),
    event_id    TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    payload     JSONB       NOT NULL
);

-- +goose Down

DROP TABLE zendesk_webhook_events;
