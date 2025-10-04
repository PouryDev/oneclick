-- Drop triggers
DROP TRIGGER IF EXISTS update_user_organizations_updated_at ON user_organizations;

DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;

-- Drop indexes
DROP INDEX IF EXISTS idx_user_organizations_org_id;

DROP INDEX IF EXISTS idx_user_organizations_user_id;

DROP INDEX IF EXISTS idx_organizations_name;

-- Drop tables
DROP TABLE IF EXISTS user_organizations;

DROP TABLE IF EXISTS organizations;