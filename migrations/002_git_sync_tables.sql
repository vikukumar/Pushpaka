-- Migration: Create git sync tracking tables
-- Description: Add tables for tracking git changes and synchronization status

-- git_sync_tracks table: main synchronization tracking
CREATE TABLE IF NOT EXISTS git_sync_tracks (
    id TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL UNIQUE,
    project_id TEXT NOT NULL,
    repository TEXT NOT NULL,
    branch TEXT NOT NULL,
    
    -- Current state
    current_commit_sha TEXT NOT NULL,
    latest_commit_sha TEXT NOT NULL,
    latest_commit_message TEXT,
    latest_commit_author TEXT,
    
    -- Synchronization status
    sync_status TEXT NOT NULL DEFAULT 'synced',
    sync_approval_required BOOLEAN DEFAULT FALSE,
    sync_approved_by TEXT,
    sync_approved_at TIMESTAMP,
    
    -- Change metadata
    total_changes INTEGER DEFAULT 0,
    total_additions INTEGER DEFAULT 0,
    total_deletions INTEGER DEFAULT 0,
    changes_summary TEXT,
    
    -- Sync tracking
    last_sync_attempt_at TIMESTAMP,
    last_sync_attempt_error TEXT,
    last_successful_sync_at TIMESTAMP,
    notification_sent_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (deployment_id) REFERENCES deployments(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- git_changes table: individual file changes
CREATE TABLE IF NOT EXISTS git_changes (
    id TEXT PRIMARY KEY,
    sync_track_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    change_type TEXT NOT NULL,
    
    additions INTEGER DEFAULT 0,
    deletions INTEGER DEFAULT 0,
    old_content TEXT,
    new_content TEXT,
    
    created_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (sync_track_id) REFERENCES git_sync_tracks(id)
);

-- git_auto_sync_config table: auto-sync configuration
CREATE TABLE IF NOT EXISTS git_auto_sync_config (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    deployment_id TEXT NOT NULL UNIQUE,
    
    enabled BOOLEAN DEFAULT FALSE,
    require_approval BOOLEAN DEFAULT FALSE,
    polling_interval INTEGER DEFAULT 3600,
    max_concurrent INTEGER DEFAULT 1,
    only_prod_ready BOOLEAN DEFAULT FALSE,
    
    allowed_branches TEXT,
    ignore_paths TEXT,
    required_approvers TEXT,
    
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (deployment_id) REFERENCES deployments(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- deployment_sync_history table: historical sync records
CREATE TABLE IF NOT EXISTS deployment_sync_history (
    id TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    
    from_commit_sha TEXT NOT NULL,
    to_commit_sha TEXT NOT NULL,
    
    sync_type TEXT NOT NULL,
    status TEXT NOT NULL,
    total_changes INTEGER DEFAULT 0,
    duration INTEGER,
    
    triggered_by TEXT,
    sync_error TEXT,
    
    rollback_triggered BOOLEAN DEFAULT FALSE,
    rollback_reason TEXT,
    
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    FOREIGN KEY (deployment_id) REFERENCES deployments(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_git_sync_tracks_project_id ON git_sync_tracks(project_id);
CREATE INDEX IF NOT EXISTS idx_git_sync_tracks_sync_status ON git_sync_tracks(sync_status);
CREATE INDEX IF NOT EXISTS idx_git_changes_sync_track_id ON git_changes(sync_track_id);
CREATE INDEX IF NOT EXISTS idx_auto_sync_config_deployment_id ON git_auto_sync_config(deployment_id);
CREATE INDEX IF NOT EXISTS idx_sync_history_deployment_id ON deployment_sync_history(deployment_id);
CREATE INDEX IF NOT EXISTS idx_sync_history_created_at ON deployment_sync_history(created_at);

-- Update projects table to include git_repo field if not exists
ALTER TABLE projects ADD COLUMN git_repo TEXT DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_projects_git_repo ON projects(git_repo);
