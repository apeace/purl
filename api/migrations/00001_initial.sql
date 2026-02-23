-- +goose Up

CREATE TYPE ticket_status   AS ENUM ('open', 'in_progress', 'resolved', 'closed');
CREATE TYPE ticket_priority AS ENUM ('low', 'medium', 'high', 'urgent');
CREATE TYPE comment_role    AS ENUM ('customer', 'agent');

-- Shared trigger function: keeps updated_at current on every row update
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE organizations (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    name               TEXT        NOT NULL,
    slug               TEXT        NOT NULL UNIQUE,
    api_key            TEXT        NOT NULL UNIQUE,
    zendesk_subdomain  TEXT,
    zendesk_api_key    TEXT
);

-- Auto-generate slug from name on insert when not explicitly provided
CREATE OR REPLACE FUNCTION organizations_set_slug()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.slug IS NULL OR NEW.slug = '' THEN
        NEW.slug := lower(regexp_replace(regexp_replace(trim(NEW.name), '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g'));
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER organizations_before_insert_set_slug
BEFORE INSERT ON organizations
FOR EACH ROW EXECUTE FUNCTION organizations_set_slug();

CREATE TRIGGER set_updated_at BEFORE UPDATE ON organizations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE customers (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name       TEXT        NOT NULL,
    org_id     UUID        NOT NULL REFERENCES organizations(id)
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON customers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE customer_emails (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    customer_id UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    email       TEXT        NOT NULL,
    verified    BOOLEAN     NOT NULL DEFAULT false
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON customer_emails
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE customer_phones (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    customer_id UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    phone       TEXT        NOT NULL,
    verified    BOOLEAN     NOT NULL DEFAULT false
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON customer_phones
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Agents are internal staff who handle and respond to tickets
CREATE TABLE agents (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    email      TEXT        NOT NULL UNIQUE,
    name       TEXT        NOT NULL,
    org_id     UUID        NOT NULL REFERENCES organizations(id)
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON agents
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE tickets (
    id          UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT now(),
    title       TEXT             NOT NULL,
    description TEXT             NOT NULL DEFAULT '',
    status      ticket_status    NOT NULL DEFAULT 'open',
    priority    ticket_priority  NOT NULL DEFAULT 'medium',
    reporter_id UUID             NOT NULL REFERENCES customers(id),
    assignee_id UUID             REFERENCES agents(id),
    org_id      UUID             NOT NULL REFERENCES organizations(id)
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON tickets
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE comments (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    ticket_id          UUID         NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    customer_author_id UUID         REFERENCES customers(id),
    agent_author_id    UUID         REFERENCES agents(id),
    role               comment_role NOT NULL,
    body               TEXT         NOT NULL,
    CONSTRAINT comments_author_check
        CHECK ((customer_author_id IS NULL) != (agent_author_id IS NULL))
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON comments
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down

DROP TRIGGER IF EXISTS set_updated_at ON comments;
DROP TRIGGER IF EXISTS set_updated_at ON tickets;
DROP TRIGGER IF EXISTS set_updated_at ON agents;
DROP TRIGGER IF EXISTS set_updated_at ON customer_phones;
DROP TRIGGER IF EXISTS set_updated_at ON customer_emails;
DROP TRIGGER IF EXISTS set_updated_at ON customers;
DROP TRIGGER IF EXISTS set_updated_at ON organizations;
DROP TRIGGER IF EXISTS organizations_before_insert_set_slug ON organizations;

DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS customer_phones;
DROP TABLE IF EXISTS customer_emails;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS organizations;

DROP FUNCTION IF EXISTS organizations_set_slug();
DROP FUNCTION IF EXISTS set_updated_at();

DROP TYPE IF EXISTS comment_role;
DROP TYPE IF EXISTS ticket_priority;
DROP TYPE IF EXISTS ticket_status;
