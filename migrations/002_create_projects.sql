-- Migration 002: Create projects table
-- Pushpaka v1.0.0

CREATE TABLE IF NOT EXISTS projects (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          VARCHAR(64) NOT NULL,
    repo_url      TEXT NOT NULL,
    branch        VARCHAR(128) NOT NULL DEFAULT 'main',
    build_command TEXT NOT NULL DEFAULT '',
    start_command TEXT NOT NULL DEFAULT '',
    port          INTEGER NOT NULL DEFAULT 3000 CHECK (port > 0 AND port < 65536),
    framework     VARCHAR(64) NOT NULL DEFAULT '',
    status        VARCHAR(32) NOT NULL DEFAULT 'inactive'
                    CHECK (status IN ('active', 'inactive', 'building')),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_user_id    ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_status     ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at DESC);

COMMENT ON TABLE projects IS 'Developer projects connected to Git repositories';
