-- Migration: 0012_events_readmodel.up.sql
-- Description: Create event logging and read model tables for audit and dashboard functionality

-- Event logs table for audit trail
CREATE TABLE event_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    action TEXT NOT NULL, -- e.g., "app_created", "pipeline_started", "cluster_imported"
    resource_type TEXT NOT NULL, -- e.g., "app", "cluster", "pipeline", "release"
    resource_id UUID NOT NULL,
    details JSONB NOT NULL DEFAULT '{}', -- Additional context and metadata
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for event_logs
CREATE INDEX idx_event_logs_org_id ON event_logs (org_id);

CREATE INDEX idx_event_logs_user_id ON event_logs (user_id);

CREATE INDEX idx_event_logs_action ON event_logs (action);

CREATE INDEX idx_event_logs_resource_type ON event_logs (resource_type);

CREATE INDEX idx_event_logs_resource_id ON event_logs (resource_id);

CREATE INDEX idx_event_logs_created_at ON event_logs (created_at DESC);

-- Read model projects table for denormalized summaries
CREATE TABLE read_model_projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    key TEXT NOT NULL, -- e.g., "recent_failed_pipelines", "top_apps_by_deployments"
    value JSONB NOT NULL DEFAULT '{}', -- Denormalized data
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (org_id, key)
);

-- Indexes for read_model_projects
CREATE INDEX idx_read_model_projects_org_id ON read_model_projects (org_id);

CREATE INDEX idx_read_model_projects_key ON read_model_projects (key);

-- Dashboard counts table for aggregated metrics
CREATE TABLE dashboard_counts (
    org_id UUID PRIMARY KEY REFERENCES organizations (id) ON DELETE CASCADE,
    apps_count INTEGER NOT NULL DEFAULT 0,
    clusters_count INTEGER NOT NULL DEFAULT 0,
    running_pipelines INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for dashboard_counts
CREATE INDEX idx_dashboard_counts_updated_at ON dashboard_counts (updated_at);

-- Function to update dashboard counts
CREATE OR REPLACE FUNCTION update_dashboard_counts(p_org_id UUID)
RETURNS VOID AS $$
BEGIN
    INSERT INTO dashboard_counts (org_id, apps_count, clusters_count, running_pipelines, updated_at)
    VALUES (
        p_org_id,
        (SELECT COUNT(*) FROM applications WHERE org_id = p_org_id),
        (SELECT COUNT(*) FROM clusters WHERE org_id = p_org_id),
        (SELECT COUNT(*) FROM pipelines WHERE org_id = p_org_id AND status IN ('pending', 'running')),
        NOW()
    )
    ON CONFLICT (org_id) DO UPDATE SET
        apps_count = EXCLUDED.apps_count,
        clusters_count = EXCLUDED.clusters_count,
        running_pipelines = EXCLUDED.running_pipelines,
        updated_at = EXCLUDED.updated_at;
END;
$$ LANGUAGE plpgsql;

-- Trigger function to automatically log events
CREATE OR REPLACE FUNCTION log_event()
RETURNS TRIGGER AS $$
DECLARE
    event_action TEXT;
    resource_type TEXT;
    org_id UUID;
    user_id UUID;
BEGIN
    -- Determine action and resource type based on operation
    IF TG_OP = 'INSERT' THEN
        event_action := TG_TABLE_NAME || '_created';
    ELSIF TG_OP = 'UPDATE' THEN
        event_action := TG_TABLE_NAME || '_updated';
    ELSIF TG_OP = 'DELETE' THEN
        event_action := TG_TABLE_NAME || '_deleted';
    END IF;
    
    -- Map table names to resource types
    resource_type := TG_TABLE_NAME;
    
    -- Get org_id and user_id based on table
    IF TG_TABLE_NAME = 'applications' THEN
        IF TG_OP = 'DELETE' THEN
            org_id := OLD.org_id;
            user_id := OLD.created_by;
        ELSE
            org_id := NEW.org_id;
            user_id := NEW.created_by;
        END IF;
    ELSIF TG_TABLE_NAME = 'clusters' THEN
        IF TG_OP = 'DELETE' THEN
            org_id := OLD.org_id;
            user_id := OLD.created_by;
        ELSE
            org_id := NEW.org_id;
            user_id := NEW.created_by;
        END IF;
    ELSIF TG_TABLE_NAME = 'pipelines' THEN
        IF TG_OP = 'DELETE' THEN
            org_id := (SELECT org_id FROM applications WHERE id = OLD.app_id);
            user_id := OLD.triggered_by;
        ELSE
            org_id := (SELECT org_id FROM applications WHERE id = NEW.app_id);
            user_id := NEW.triggered_by;
        END IF;
    ELSE
        -- For other tables, try to get org_id and user_id from common patterns
        IF TG_OP = 'DELETE' THEN
            org_id := OLD.org_id;
            user_id := OLD.created_by;
        ELSE
            org_id := NEW.org_id;
            user_id := NEW.created_by;
        END IF;
    END IF;
    
    -- Insert event log
    INSERT INTO event_logs (org_id, user_id, action, resource_type, resource_id, details)
    VALUES (
        org_id,
        user_id,
        event_action,
        resource_type,
        COALESCE(NEW.id, OLD.id),
        CASE 
            WHEN TG_OP = 'INSERT' THEN to_jsonb(NEW)
            WHEN TG_OP = 'UPDATE' THEN to_jsonb(NEW)
            WHEN TG_OP = 'DELETE' THEN to_jsonb(OLD)
        END
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic event logging (optional - can be disabled if manual logging is preferred)
-- Uncomment these if you want automatic event logging via triggers

-- CREATE TRIGGER applications_event_trigger
--     AFTER INSERT OR UPDATE OR DELETE ON applications
--     FOR EACH ROW EXECUTE FUNCTION log_event();

-- CREATE TRIGGER clusters_event_trigger
--     AFTER INSERT OR UPDATE OR DELETE ON clusters
--     FOR EACH ROW EXECUTE FUNCTION log_event();

-- CREATE TRIGGER pipelines_event_trigger
--     AFTER INSERT OR UPDATE OR DELETE ON pipelines
--     FOR EACH ROW EXECUTE FUNCTION log_event();