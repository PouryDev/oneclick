-- Drop trigger
DROP TRIGGER IF EXISTS update_clusters_updated_at ON clusters;

-- Drop indexes
DROP INDEX IF EXISTS idx_clusters_provider;
DROP INDEX IF EXISTS idx_clusters_status;
DROP INDEX IF EXISTS idx_clusters_org_id;

-- Drop table
DROP TABLE IF EXISTS clusters;
