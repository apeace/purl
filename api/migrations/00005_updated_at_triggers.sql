-- +goose Up

-- Shared trigger function used by all tables
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add missing updated_at columns
ALTER TABLE comments       ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE customer_emails ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE customer_phones ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Wire up triggers on every table
CREATE TRIGGER set_updated_at BEFORE UPDATE ON customers        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON tickets          FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON comments         FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON organizations    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON customer_emails  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON customer_phones  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON agents           FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down

DROP TRIGGER IF EXISTS set_updated_at ON customers;
DROP TRIGGER IF EXISTS set_updated_at ON tickets;
DROP TRIGGER IF EXISTS set_updated_at ON comments;
DROP TRIGGER IF EXISTS set_updated_at ON organizations;
DROP TRIGGER IF EXISTS set_updated_at ON customer_emails;
DROP TRIGGER IF EXISTS set_updated_at ON customer_phones;
DROP TRIGGER IF EXISTS set_updated_at ON agents;

DROP FUNCTION IF EXISTS set_updated_at();

ALTER TABLE customer_phones  DROP COLUMN updated_at;
ALTER TABLE customer_emails  DROP COLUMN updated_at;
ALTER TABLE comments         DROP COLUMN updated_at;
