-- Migration 005: Create environment_variables table
-- Pushpaka v1.0.0

CREATE TABLE IF NOT EXISTS environment_variables (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        VARCHAR(255) NOT NULL,
    value      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_env_project_key UNIQUE (project_id, key)
);

CREATE INDEX IF NOT EXISTS idx_env_project_id ON environment_variables(project_id);

COMMENT ON TABLE environment_variables IS 'Encrypted environment variables injected at container runtime';
