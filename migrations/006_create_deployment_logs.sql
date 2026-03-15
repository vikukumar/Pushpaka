-- Migration 006: Create deployment_logs table
-- Pushpaka v1.0.0

CREATE TABLE IF NOT EXISTS deployment_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    level         VARCHAR(16) NOT NULL DEFAULT 'info'
                    CHECK (level IN ('info', 'warn', 'error', 'debug')),
    stream        VARCHAR(16) NOT NULL DEFAULT 'stdout'
                    CHECK (stream IN ('stdout', 'stderr', 'system')),
    message       TEXT NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_deployment_id ON deployment_logs(deployment_id);
CREATE INDEX IF NOT EXISTS idx_logs_created_at    ON deployment_logs(deployment_id, created_at ASC);

COMMENT ON TABLE deployment_logs IS 'Streaming logs from build and deploy processes';
