-- +goose Up

CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    api_key    TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add nullable first so the ALTER succeeds on databases with existing rows.
ALTER TABLE users ADD COLUMN org_id UUID REFERENCES organizations(id);
ALTER TABLE tickets ADD COLUMN org_id UUID REFERENCES organizations(id);

-- Existing rows have no org — clear them out so we can enforce NOT NULL.
-- The seed tool will repopulate everything with correct org associations.
TRUNCATE tickets CASCADE;
TRUNCATE users CASCADE;

ALTER TABLE users ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE tickets ALTER COLUMN org_id SET NOT NULL;

-- +goose Down

ALTER TABLE tickets DROP COLUMN org_id;
ALTER TABLE users DROP COLUMN org_id;
DROP TABLE IF EXISTS organizations;
