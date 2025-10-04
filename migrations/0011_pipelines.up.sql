-- Migration: 0011_pipelines.up.sql
-- Description: Create pipelines and pipeline_steps tables for CI/CD pipeline management

-- Create pipelines table
CREATE TABLE pipelines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    commit_sha TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled')),
    triggered_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    logs_url TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    meta JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create pipeline_steps table
CREATE TABLE pipeline_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'success', 'failed', 'skipped', 'cancelled')),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    logs TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_pipelines_app_id ON pipelines(app_id);
CREATE INDEX idx_pipelines_repo_id ON pipelines(repo_id);
CREATE INDEX idx_pipelines_status ON pipelines(status);
CREATE INDEX idx_pipelines_triggered_by ON pipelines(triggered_by);
CREATE INDEX idx_pipelines_created_at ON pipelines(created_at DESC);
CREATE INDEX idx_pipelines_commit_sha ON pipelines(commit_sha);

CREATE INDEX idx_pipeline_steps_pipeline_id ON pipeline_steps(pipeline_id);
CREATE INDEX idx_pipeline_steps_status ON pipeline_steps(status);
CREATE INDEX idx_pipeline_steps_name ON pipeline_steps(name);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_pipelines_updated_at BEFORE UPDATE ON pipelines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pipeline_steps_updated_at BEFORE UPDATE ON pipeline_steps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add constraint to ensure finished_at is after started_at
ALTER TABLE pipelines ADD CONSTRAINT check_pipeline_times 
    CHECK (finished_at IS NULL OR started_at IS NULL OR finished_at >= started_at);

ALTER TABLE pipeline_steps ADD CONSTRAINT check_pipeline_step_times 
    CHECK (finished_at IS NULL OR started_at IS NULL OR finished_at >= started_at);
