-- Drop trigger
DROP TRIGGER IF EXISTS update_repositories_updated_at ON repositories;

-- Drop indexes
CREATE INDEX IF EXISTS idx_repositories_url;

DROP INDEX IF EXISTS idx_repositories_type;

DROP INDEX IF EXISTS idx_repositories_org_id;

-- Drop table
DROP TABLE IF EXISTS repositories;