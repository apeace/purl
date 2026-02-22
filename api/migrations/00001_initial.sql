-- +goose Up

CREATE TYPE ticket_status AS ENUM ('open', 'in_progress', 'resolved', 'closed');
CREATE TYPE ticket_priority AS ENUM ('low', 'medium', 'high', 'urgent');

CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tickets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      ticket_status NOT NULL DEFAULT 'open',
    priority    ticket_priority NOT NULL DEFAULT 'medium',
    reporter_id UUID NOT NULL REFERENCES users(id),
    assignee_id UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id  UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_id  UUID NOT NULL REFERENCES users(id),
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS ticket_priority;
DROP TYPE IF EXISTS ticket_status;
