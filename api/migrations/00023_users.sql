-- +goose Up

CREATE TABLE users (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name       TEXT        NOT NULL,
    email      TEXT        NOT NULL,
    google_id  TEXT,
    org_id     UUID        NOT NULL REFERENCES organizations(id)
);

CREATE UNIQUE INDEX users_org_id_email_unique ON users (org_id, email);
CREATE UNIQUE INDEX users_google_id_unique ON users (google_id) WHERE google_id IS NOT NULL;

CREATE TRIGGER set_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down

DROP TRIGGER IF EXISTS set_updated_at ON users;
DROP TABLE IF EXISTS users;
