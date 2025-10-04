-- Create repositories table
CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    url TEXT NOT NULL,
    default_branch TEXT NOT NULL DEFAULT 'main',
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_repositories_org_id ON repositories (org_id);

CREATE INDEX idx_repositories_type ON repositories(type);

CREATE INDEX idx_repositories_url ON repositories (url);

-- Create trigger to update updated_at for repositories
CREATE TRIGGER update_repositories_updated_at 
    BEFORE UPDATE ON repositories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraint for valid repository types
ALTER TABLE repositories
ADD CONSTRAINT check_repository_type CHECK (
    type IN ('github', 'gitlab', 'gitea')
);

-- Add unique constraint for org_id + url combination
ALTER TABLE repositories
ADD CONSTRAINT unique_org_repo_url UNIQUE (org_id, url);