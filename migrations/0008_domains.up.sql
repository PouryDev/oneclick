-- Create domains table
CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    app_id UUID NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    domain TEXT NOT NULL,
    provider TEXT NOT NULL,
    provider_config JSONB NOT NULL DEFAULT '{}',
    cert_status TEXT NOT NULL DEFAULT 'pending', -- pending, active, failed, expired
    cert_secret_name TEXT,
    challenge_type TEXT NOT NULL DEFAULT 'http-01', -- http-01, dns-01
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_domain_provider CHECK (
        provider IN (
            'cloudflare',
            'route53',
            'manual'
        )
    ),
    CONSTRAINT chk_cert_status CHECK (
        cert_status IN (
            'pending',
            'active',
            'failed',
            'expired'
        )
    ),
    CONSTRAINT chk_challenge_type CHECK (
        challenge_type IN ('http-01', 'dns-01')
    )
);

-- Create indexes
CREATE INDEX idx_domains_app_id ON domains (app_id);

CREATE INDEX idx_domains_domain ON domains (domain);

CREATE INDEX idx_domains_cert_status ON domains (cert_status);

CREATE INDEX idx_domains_provider ON domains (provider);

-- Create trigger to update updated_at
CREATE TRIGGER update_domains_updated_at 
    BEFORE UPDATE ON domains 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add unique constraint for domain per app
ALTER TABLE domains
ADD CONSTRAINT unique_domain_per_app UNIQUE (app_id, domain);