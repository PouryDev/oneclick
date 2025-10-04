-- Create services table
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    app_id UUID NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    chart TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, provisioning, running, failed, stopped
    namespace TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create service_configs table
CREATE TABLE service_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    service_id UUID NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_services_app_id ON services (app_id);

CREATE INDEX idx_services_status ON services (status);

CREATE INDEX idx_services_name ON services (name);

CREATE INDEX idx_service_configs_service_id ON service_configs (service_id);

CREATE INDEX idx_service_configs_key ON service_configs (key);

CREATE INDEX idx_service_configs_is_secret ON service_configs (is_secret);

-- Create triggers to update updated_at for services and service_configs
CREATE TRIGGER update_services_updated_at 
    BEFORE UPDATE ON services 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_service_configs_updated_at 
    BEFORE UPDATE ON service_configs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraints for valid service statuses
ALTER TABLE services
ADD CONSTRAINT check_service_status CHECK (
    status IN (
        'pending',
        'provisioning',
        'running',
        'failed',
        'stopped'
    )
);

-- Add unique constraint for service name within application
ALTER TABLE services
ADD CONSTRAINT unique_service_name_in_app UNIQUE (app_id, name);

-- Add unique constraint for service config key within service
ALTER TABLE service_configs
ADD CONSTRAINT unique_config_key_in_service UNIQUE (service_id, key);