-- Drop trigger
DROP TRIGGER IF EXISTS update_domains_updated_at ON domains;

-- Drop indexes
DROP INDEX IF EXISTS idx_domains_provider;

DROP INDEX IF EXISTS idx_domains_cert_status;

DROP INDEX IF EXISTS idx_domains_domain;

DROP INDEX IF EXISTS idx_domains_app_id;

-- Drop table
DROP TABLE IF EXISTS domains;