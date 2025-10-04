-- Drop triggers
DROP TRIGGER IF EXISTS update_services_updated_at ON services;

DROP TRIGGER IF EXISTS update_service_configs_updated_at ON service_configs;

-- Drop indexes
DROP INDEX IF EXISTS idx_service_configs_is_secret;

DROP INDEX IF EXISTS idx_service_configs_key;

DROP INDEX IF EXISTS idx_service_configs_service_id;

DROP INDEX IF EXISTS idx_services_name;

DROP INDEX IF EXISTS idx_services_status;

DROP INDEX IF EXISTS idx_services_app_id;

-- Drop tables
DROP TABLE IF EXISTS service_configs;

DROP TABLE IF EXISTS services;