-- +goose Up

CREATE TYPE zendesk_status_category AS ENUM ('new', 'open', 'pending', 'solved', 'closed');

CREATE TABLE boards (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    org_id     UUID        NOT NULL REFERENCES organizations(id),
    name       TEXT        NOT NULL,
    is_default BOOLEAN     NOT NULL DEFAULT false
);

-- Partial index enforces at most one default board per org
CREATE UNIQUE INDEX boards_one_default_per_org ON boards (org_id) WHERE (is_default = true);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON boards
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE board_columns (
    id             UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ             NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ             NOT NULL DEFAULT now(),
    board_id       UUID                    NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name           TEXT                    NOT NULL,
    position       INTEGER                 NOT NULL CHECK (position >= 0),
    zendesk_status zendesk_status_category NOT NULL,
    color          TEXT                    NOT NULL,
    UNIQUE (board_id, position),
    UNIQUE (board_id, zendesk_status)
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON board_columns
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE kanban_board_tickets (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    board_id   UUID        NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    column_id  UUID        NOT NULL REFERENCES board_columns(id) ON DELETE CASCADE,
    ticket_id  UUID        NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    position   INTEGER     NOT NULL CHECK (position >= 0),
    UNIQUE (board_id, ticket_id),
    UNIQUE (column_id, position)
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON kanban_board_tickets
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down

DROP TRIGGER IF EXISTS set_updated_at ON kanban_board_tickets;
DROP TRIGGER IF EXISTS set_updated_at ON board_columns;
DROP TRIGGER IF EXISTS set_updated_at ON boards;

DROP TABLE IF EXISTS kanban_board_tickets;
DROP TABLE IF EXISTS board_columns;
DROP TABLE IF EXISTS boards;

DROP TYPE IF EXISTS zendesk_status_category;
