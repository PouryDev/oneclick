-- Drop triggers
DROP TRIGGER IF EXISTS update_git_servers_updated_at ON git_servers;

DROP TRIGGER IF EXISTS update_runners_updated_at ON runners;

-- Drop indexes
DROP INDEX IF EXISTS idx_job_queue_created_at;

DROP INDEX IF EXISTS idx_job_queue_type;

DROP INDEX IF EXISTS idx_job_queue_status;

DROP INDEX IF EXISTS idx_job_queue_org_id;

DROP INDEX IF EXISTS idx_runners_type;

DROP INDEX IF EXISTS idx_runners_status;

DROP INDEX IF EXISTS idx_runners_org_id;

DROP INDEX IF EXISTS idx_git_servers_type;

DROP INDEX IF EXISTS idx_git_servers_status;

DROP INDEX IF EXISTS idx_git_servers_org_id;

-- Drop tables
DROP TABLE IF EXISTS job_queue;

DROP TABLE IF EXISTS runners;

DROP TABLE IF EXISTS git_servers;