-- +goose Up

ALTER TABLE organizations ADD COLUMN slug TEXT UNIQUE;

-- Backfill slugs for existing rows
UPDATE organizations
SET slug = lower(regexp_replace(regexp_replace(trim(name), '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g'));

ALTER TABLE organizations ALTER COLUMN slug SET NOT NULL;

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

ALTER TABLE organizations ADD COLUMN zendesk_subdomain TEXT;
ALTER TABLE organizations ADD COLUMN zendesk_api_key TEXT;

-- +goose Down

DROP TRIGGER IF EXISTS organizations_before_insert_set_slug ON organizations;
DROP FUNCTION IF EXISTS organizations_set_slug();
ALTER TABLE organizations DROP COLUMN zendesk_api_key;
ALTER TABLE organizations DROP COLUMN zendesk_subdomain;
ALTER TABLE organizations DROP COLUMN slug;
