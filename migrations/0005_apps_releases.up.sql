-- Create applications table
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    cluster_id UUID NOT NULL REFERENCES clusters (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    repo_id UUID NOT NULL REFERENCES repositories (id) ON DELETE CASCADE,
    path TEXT, -- Optional path within repository
    default_branch TEXT NOT NULL DEFAULT 'main',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create releases table
CREATE TABLE releases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    app_id UUID NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    image TEXT NOT NULL,
    tag TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, running, succeeded, failed
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    meta JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_applications_org_id ON applications (org_id);

CREATE INDEX idx_applications_cluster_id ON applications (cluster_id);

CREATE INDEX idx_applications_repo_id ON applications (repo_id);

CREATE INDEX idx_applications_name ON applications (name);

CREATE INDEX idx_releases_app_id ON releases (app_id);

CREATE INDEX idx_releases_status ON releases (status);

CREATE INDEX idx_releases_created_by ON releases (created_by);

CREATE INDEX idx_releases_created_at ON releases (created_at);

-- Create triggers to update updated_at for applications and releases
CREATE TRIGGER update_applications_updated_at 
    BEFORE UPDATE ON applications 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_releases_updated_at 
    BEFORE UPDATE ON releases 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraints for valid release statuses
ALTER TABLE releases
ADD CONSTRAINT check_release_status CHECK (
    status IN (
        'pending',
        'running',
        'succeeded',
        'failed'
    )
);

-- Add unique constraint for application name within cluster
ALTER TABLE applications
ADD CONSTRAINT unique_app_name_in_cluster UNIQUE (cluster_id, name);