-- Migration: Add Advanced Deployment Management Tables
-- Description: Support for multi-instance deployments, code signatures, backups, and deployment actions

-- Projects table additions (handled via ALTER if needed)
-- ALTER TABLE projects ADD COLUMN max_deployments INTEGER DEFAULT 2;
-- ALTER TABLE projects ADD COLUMN max_backups INTEGER DEFAULT 3;
-- ALTER TABLE projects ADD COLUMN clone_directory TEXT DEFAULT '';
-- ALTER TABLE projects ADD COLUMN git_clone_path TEXT DEFAULT '';
-- ALTER TABLE projects ADD COLUMN main_deploy_id TEXT DEFAULT '';

-- deployment_code_signatures: Track code state at deployment time
CREATE TABLE IF NOT EXISTS deployment_code_signatures (
    id TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    
    -- Git metadata
    commit_sha TEXT NOT NULL,
    commit_message TEXT,
    commit_author TEXT,
    branch TEXT NOT NULL,
    
    -- Code state
    code_hash TEXT NOT NULL,              -- SHA256 of entire codebase
    file_count INTEGER DEFAULT 0,
    directory_path TEXT NOT NULL,         -- Where code is stored
    
    created_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (deployment_id) REFERENCES deployments(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    UNIQUE(deployment_id, commit_sha)
);

-- deployment_instances: Individual deployment instances
CREATE TABLE IF NOT EXISTS deployment_instances (
    id TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL,          -- Parent deployment
    project_id TEXT NOT NULL,
    
    role TEXT NOT NULL,                   -- main/testing/backup
    status TEXT NOT NULL DEFAULT 'stopped',  -- running/stopped/failed/testing
    
    -- Container/Process info
    container_id TEXT,
    process_id INTEGER,
    port INTEGER,
    
    -- Code tracking
    code_signature_id TEXT,
    instance_dir TEXT NOT NULL,           -- Working directory for this instance
    
    -- Lifecycle tracking
    started_at TIMESTAMP,
    stopped_at TIMESTAMP,
    last_health_check TIMESTAMP,
    health_status TEXT DEFAULT 'unknown',  -- healthy/unhealthy/unknown
    
    -- Error tracking
    restart_count INTEGER DEFAULT 0,
    error_log TEXT,
    
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (deployment_id) REFERENCES deployments(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (code_signature_id) REFERENCES deployment_code_signatures(id)
);

-- deployment_backups: Backup versions for rollback
CREATE TABLE IF NOT EXISTS deployment_backups (
    id TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    
    code_signature_id TEXT,
    instance_id TEXT,
    
    backup_path TEXT NOT NULL,            -- Directory where backup is stored
    size INTEGER DEFAULT 0,               -- Size in bytes
    
    reason TEXT,                          -- rollback/upgrade/manual
    is_restored BOOLEAN DEFAULT FALSE,    -- Was this backup restored?
    restored_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (deployment_id) REFERENCES deployments(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (code_signature_id) REFERENCES deployment_code_signatures(id),
    FOREIGN KEY (instance_id) REFERENCES deployment_instances(id)
);

-- deployment_actions: Track user actions on deployments
CREATE TABLE IF NOT EXISTS deployment_actions (
    id TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL,
    instance_id TEXT,
    project_id TEXT NOT NULL,
    
    user_id TEXT NOT NULL,
    action TEXT NOT NULL,                 -- start/stop/restart/retry/rollback/sync
    
    status TEXT NOT NULL DEFAULT 'pending', -- pending/executing/success/failed
    result TEXT,                          -- Result message
    
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (deployment_id) REFERENCES deployments(id),
    FOREIGN KEY (instance_id) REFERENCES deployment_instances(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- project_deployment_stats: Statistics for project deployments
CREATE TABLE IF NOT EXISTS project_deployment_stats (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL UNIQUE,
    
    main_deployment_id TEXT,
    testing_deployment_id TEXT,
    
    total_deployments INTEGER DEFAULT 0,
    successful_deploys INTEGER DEFAULT 0,
    failed_deploys INTEGER DEFAULT 0,
    total_backups INTEGER DEFAULT 0,
    
    last_deploy_at TIMESTAMP,
    avg_deploy_time INTEGER DEFAULT 0,    -- Average deployment time in seconds
    
    updated_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_code_signatures_deployment_id ON deployment_code_signatures(deployment_id);
CREATE INDEX IF NOT EXISTS idx_code_signatures_project_id ON deployment_code_signatures(project_id);
CREATE INDEX IF NOT EXISTS idx_code_signatures_created_at ON deployment_code_signatures(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_deployment_instances_deployment_id ON deployment_instances(deployment_id);
CREATE INDEX IF NOT EXISTS idx_deployment_instances_project_id ON deployment_instances(project_id);
CREATE INDEX IF NOT EXISTS idx_deployment_instances_role_status ON deployment_instances(project_id, role, status);
CREATE INDEX IF NOT EXISTS idx_deployment_instances_created_at ON deployment_instances(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_deployment_backups_deployment_id ON deployment_backups(deployment_id);
CREATE INDEX IF NOT EXISTS idx_deployment_backups_project_id ON deployment_backups(project_id);
CREATE INDEX IF NOT EXISTS idx_deployment_backups_created_at ON deployment_backups(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_deployment_actions_deployment_id ON deployment_actions(deployment_id);
CREATE INDEX IF NOT EXISTS idx_deployment_actions_instance_id ON deployment_actions(instance_id);
CREATE INDEX IF NOT EXISTS idx_deployment_actions_project_id ON deployment_actions(project_id);
CREATE INDEX IF NOT EXISTS idx_deployment_actions_created_at ON deployment_actions(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_project_stats_project_id ON project_deployment_stats(project_id);

-- Update projects table
ALTER TABLE projects ADD COLUMN IF NOT EXISTS max_deployments INTEGER DEFAULT 2;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS max_backups INTEGER DEFAULT 3;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS clone_directory TEXT DEFAULT '';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS git_clone_path TEXT DEFAULT '';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS main_deploy_id TEXT DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_projects_max_deployments ON projects(max_deployments);
CREATE INDEX IF NOT EXISTS idx_projects_git_clone_path ON projects(git_clone_path);
