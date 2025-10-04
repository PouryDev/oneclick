-- Create git_servers table
CREATE TABLE git_servers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    type TEXT NOT NULL DEFAULT 'gitea',
    domain TEXT NOT NULL,
    storage TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, provisioning, running, failed, stopped
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_git_server_type CHECK (type IN ('gitea')),
    CONSTRAINT chk_git_server_status CHECK (
        status IN (
            'pending',
            'provisioning',
            'running',
            'failed',
            'stopped'
        )
    )
);

-- Create runners table
CREATE TABLE runners (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending', -- pending, provisioning, running, failed, stopped
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_runner_type CHECK (
        type IN ('github', 'gitlab', 'custom')
    ),
    CONSTRAINT chk_runner_status CHECK (
        status IN (
            'pending',
            'provisioning',
            'running',
            'failed',
            'stopped'
        )
    )
);

-- Create job_queue table for background processing
CREATE TABLE job_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    type TEXT NOT NULL, -- 'git_server_install', 'runner_deploy', etc.
    status TEXT NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    payload JSONB NOT NULL DEFAULT '{}',
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT chk_job_status CHECK (
        status IN (
            'pending',
            'processing',
            'completed',
            'failed'
        )
    )
);

-- Create indexes
CREATE INDEX idx_git_servers_org_id ON git_servers (org_id);

CREATE INDEX idx_git_servers_status ON git_servers (status);

CREATE INDEX idx_git_servers_type ON git_servers(type);

CREATE INDEX idx_runners_org_id ON runners (org_id);

CREATE INDEX idx_runners_status ON runners (status);

CREATE INDEX idx_runners_type ON runners(type);

CREATE INDEX idx_job_queue_org_id ON job_queue (org_id);

CREATE INDEX idx_job_queue_status ON job_queue (status);

CREATE INDEX idx_job_queue_type ON job_queue(type);

CREATE INDEX idx_job_queue_created_at ON job_queue (created_at);

-- Create triggers to update updated_at for git_servers and runners
CREATE TRIGGER update_git_servers_updated_at 
    BEFORE UPDATE ON git_servers 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_runners_updated_at 
    BEFORE UPDATE ON runners 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add unique constraints
ALTER TABLE git_servers
ADD CONSTRAINT unique_git_server_domain_per_org UNIQUE (org_id, domain);

ALTER TABLE runners
ADD CONSTRAINT unique_runner_name_per_org UNIQUE (org_id, name);