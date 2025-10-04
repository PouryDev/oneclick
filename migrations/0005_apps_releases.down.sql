-- Drop triggers
DROP TRIGGER IF EXISTS update_applications_updated_at ON applications;

DROP TRIGGER IF EXISTS update_releases_updated_at ON releases;

-- Drop indexes
DROP INDEX IF EXISTS idx_releases_created_at;

DROP INDEX IF EXISTS idx_releases_created_by;

DROP INDEX IF EXISTS idx_releases_status;

DROP INDEX IF EXISTS idx_releases_app_id;

DROP INDEX IF EXISTS idx_applications_name;

DROP INDEX IF EXISTS idx_applications_repo_id;

DROP INDEX IF EXISTS idx_applications_cluster_id;

DROP INDEX IF EXISTS idx_applications_org_id;

-- Drop tables
DROP TABLE IF EXISTS releases;

DROP TABLE IF EXISTS applications;