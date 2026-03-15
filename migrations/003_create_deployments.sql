-- Migration 003: Create deployments table
-- Pushpaka v1.0.0

CREATE TABLE IF NOT EXISTS deployments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    commit_sha   VARCHAR(40) NOT NULL DEFAULT '',
    commit_msg   TEXT NOT NULL DEFAULT '',
    branch       VARCHAR(128) NOT NULL DEFAULT 'main',
    status       VARCHAR(32) NOT NULL DEFAULT 'queued'
                   CHECK (status IN ('queued', 'building', 'running', 'failed', 'stopped')),
    image_tag    TEXT NOT NULL DEFAULT '',
    container_id TEXT NOT NULL DEFAULT '',
    url          TEXT NOT NULL DEFAULT '',
    error_msg    TEXT NOT NULL DEFAULT '',
    started_at   TIMESTAMP WITH TIME ZONE,
    finished_at  TIMESTAMP WITH TIME ZONE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deployments_project_id  ON deployments(project_id);
CREATE INDEX IF NOT EXISTS idx_deployments_user_id     ON deployments(user_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status      ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_created_at  ON deployments(created_at DESC);

COMMENT ON TABLE deployments IS 'Deployment records tracking each build and container run';
