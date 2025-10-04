-- Create organizations table
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create user_organizations junction table
CREATE TABLE user_organizations (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, org_id)
);

-- Create indexes
CREATE INDEX idx_organizations_name ON organizations (name);

CREATE INDEX idx_user_organizations_user_id ON user_organizations (user_id);

CREATE INDEX idx_user_organizations_org_id ON user_organizations (org_id);

-- Create trigger to update updated_at for organizations
CREATE TRIGGER update_organizations_updated_at 
    BEFORE UPDATE ON organizations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger to update updated_at for user_organizations
CREATE TRIGGER update_user_organizations_updated_at 
    BEFORE UPDATE ON user_organizations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();