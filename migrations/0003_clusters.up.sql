-- Create clusters table
CREATE TABLE clusters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    provider TEXT NOT NULL,
    region TEXT NOT NULL,
    kubeconfig_encrypted BYTEA,
    node_count INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'provisioning',
    kube_version TEXT,
    last_health_check TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_clusters_org_id ON clusters(org_id);
CREATE INDEX idx_clusters_status ON clusters(status);
CREATE INDEX idx_clusters_provider ON clusters(provider);

-- Create trigger to update updated_at for clusters
CREATE TRIGGER update_clusters_updated_at 
    BEFORE UPDATE ON clusters 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraint for valid status values
ALTER TABLE clusters ADD CONSTRAINT check_cluster_status 
    CHECK (status IN ('provisioning', 'active', 'error', 'deleting'));
