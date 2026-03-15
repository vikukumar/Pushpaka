-- Migration 004: Create domains table
-- Pushpaka v1.0.0

CREATE TABLE IF NOT EXISTS domains (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain      VARCHAR(253) NOT NULL UNIQUE,
    verified    BOOLEAN NOT NULL DEFAULT FALSE,
    ssl_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_domains_project_id ON domains(project_id);
CREATE INDEX IF NOT EXISTS idx_domains_user_id    ON domains(user_id);
CREATE INDEX IF NOT EXISTS idx_domains_domain     ON domains(domain);

COMMENT ON TABLE domains IS 'Custom domain configurations per project';
