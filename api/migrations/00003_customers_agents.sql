-- +goose Up

-- Rename users to customers
ALTER TABLE users RENAME TO customers;

-- Drop email from customers; contact details move to dedicated tables
ALTER TABLE customers DROP COLUMN email;

CREATE TABLE customer_emails (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    email       TEXT NOT NULL,
    verified    BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE customer_phones (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    phone       TEXT NOT NULL,
    verified    BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Agents are internal staff who handle and respond to tickets
CREATE TABLE agents (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    org_id     UUID NOT NULL REFERENCES organizations(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tickets are filed by customers and assigned to agents
ALTER TABLE tickets DROP CONSTRAINT tickets_assignee_id_fkey;
ALTER TABLE tickets ADD CONSTRAINT tickets_assignee_id_fkey
    FOREIGN KEY (assignee_id) REFERENCES agents(id);

-- Comments carry a role to distinguish customer messages from agent responses
CREATE TYPE comment_role AS ENUM ('customer', 'agent');

ALTER TABLE comments RENAME COLUMN author_id TO customer_author_id;
ALTER TABLE comments ALTER COLUMN customer_author_id DROP NOT NULL;
ALTER TABLE comments ADD COLUMN agent_author_id UUID REFERENCES agents(id);
ALTER TABLE comments ADD COLUMN role comment_role NOT NULL DEFAULT 'customer';
ALTER TABLE comments ADD CONSTRAINT comments_author_check
    CHECK ((customer_author_id IS NULL) != (agent_author_id IS NULL));

-- +goose Down

ALTER TABLE comments DROP CONSTRAINT comments_author_check;
ALTER TABLE comments DROP COLUMN role;
ALTER TABLE comments DROP COLUMN agent_author_id;
ALTER TABLE comments ALTER COLUMN customer_author_id SET NOT NULL;
ALTER TABLE comments RENAME COLUMN customer_author_id TO author_id;
DROP TYPE comment_role;

ALTER TABLE tickets DROP CONSTRAINT tickets_assignee_id_fkey;
ALTER TABLE tickets ADD CONSTRAINT tickets_assignee_id_fkey
    FOREIGN KEY (assignee_id) REFERENCES customers(id);

DROP TABLE agents;
DROP TABLE customer_phones;
DROP TABLE customer_emails;

-- Restore email column before renaming back
ALTER TABLE customers ADD COLUMN email TEXT NOT NULL DEFAULT '';
ALTER TABLE customers ALTER COLUMN email DROP DEFAULT;
ALTER TABLE customers ADD CONSTRAINT customers_email_key UNIQUE (email);
ALTER TABLE customers RENAME TO users;
