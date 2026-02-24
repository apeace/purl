-- +goose Up

CREATE TABLE zendesk_webhook_events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    org_id       UUID        NOT NULL REFERENCES organizations(id),
    event_id     TEXT        NOT NULL,
    event_type   TEXT        NOT NULL,
    payload      JSONB       NOT NULL,
    processed_at TIMESTAMPTZ
);

-- Partial index for efficiently scanning unprocessed events in arrival order.
CREATE INDEX zendesk_webhook_events_unprocessed
    ON zendesk_webhook_events (created_at)
    WHERE processed_at IS NULL;

-- +goose Down

DROP INDEX zendesk_webhook_events_unprocessed;
DROP TABLE zendesk_webhook_events;
