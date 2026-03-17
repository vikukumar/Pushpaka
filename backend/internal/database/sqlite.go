package database

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite" // registers "sqlite" driver
)

// sqliteSchema is the SQLite-compatible equivalent of migrations 001-006.
// UUIDs -> TEXT, BOOLEAN -> INTEGER, TIMESTAMP WITH TIME ZONE -> TEXT (RFC3339).
// All statements are idempotent (IF NOT EXISTS / IF NOT EXISTS index).
const sqliteSchema = `
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    api_key       TEXT NOT NULL UNIQUE,
    role          TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_email   ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

CREATE TABLE IF NOT EXISTS projects (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    repo_url      TEXT NOT NULL,
    branch        TEXT NOT NULL DEFAULT 'main',
    build_command TEXT NOT NULL DEFAULT '',
    start_command TEXT NOT NULL DEFAULT '',
    port          INTEGER NOT NULL DEFAULT 3000 CHECK (port > 0 AND port < 65536),
    framework     TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'inactive'
                    CHECK (status IN ('active', 'inactive', 'building')),
    is_private    INTEGER NOT NULL DEFAULT 0,
    git_token     TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_projects_user_id    ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_status     ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at DESC);

CREATE TABLE IF NOT EXISTS deployments (
    id            TEXT PRIMARY KEY,
    project_id    TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    commit_sha    TEXT NOT NULL DEFAULT '',
    commit_msg    TEXT NOT NULL DEFAULT '',
    branch        TEXT NOT NULL DEFAULT 'main',
    status        TEXT NOT NULL DEFAULT 'queued'
                    CHECK (status IN ('queued', 'building', 'running', 'failed', 'stopped')),
    image_tag     TEXT NOT NULL DEFAULT '',
    container_id  TEXT NOT NULL DEFAULT '',
    url           TEXT NOT NULL DEFAULT '',
    external_port INTEGER NOT NULL DEFAULT 0,
    error_msg     TEXT NOT NULL DEFAULT '',
    started_at    TEXT,
    finished_at   TEXT,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_deployments_project_id ON deployments(project_id);
CREATE INDEX IF NOT EXISTS idx_deployments_user_id    ON deployments(user_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status     ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON deployments(created_at DESC);

CREATE TABLE IF NOT EXISTS domains (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain      TEXT NOT NULL UNIQUE,
    verified    INTEGER NOT NULL DEFAULT 0,
    ssl_enabled INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_domains_project_id ON domains(project_id);
CREATE INDEX IF NOT EXISTS idx_domains_user_id    ON domains(user_id);
CREATE INDEX IF NOT EXISTS idx_domains_domain     ON domains(domain);

CREATE TABLE IF NOT EXISTS environment_variables (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        TEXT NOT NULL,
    value      TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT uq_env_project_key UNIQUE (project_id, key)
);
CREATE INDEX IF NOT EXISTS idx_env_project_id ON environment_variables(project_id);

CREATE TABLE IF NOT EXISTS deployment_logs (
    id            TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    level         TEXT NOT NULL DEFAULT 'info'
                    CHECK (level IN ('info', 'warn', 'error', 'debug')),
    stream        TEXT NOT NULL DEFAULT 'stdout'
                    CHECK (stream IN ('stdout', 'stderr', 'system')),
    message       TEXT NOT NULL,
    created_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_logs_deployment_id ON deployment_logs(deployment_id);
CREATE INDEX IF NOT EXISTS idx_logs_created_at    ON deployment_logs(deployment_id, created_at ASC);

-- Audit logs: immutable record of every user action
CREATE TABLE IF NOT EXISTS audit_logs (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action      TEXT NOT NULL,       -- e.g. "project.create", "deployment.trigger"
    resource    TEXT NOT NULL,       -- "project", "deployment", etc.
    resource_id TEXT NOT NULL DEFAULT '',
    metadata    TEXT NOT NULL DEFAULT '{}',  -- arbitrary JSON blob
    ip_addr     TEXT NOT NULL DEFAULT '',
    user_agent  TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_user_id    ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action     ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at DESC);

-- Per-user notification settings (Slack / Discord / SMTP)
CREATE TABLE IF NOT EXISTS notification_configs (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    slack_webhook_url   TEXT NOT NULL DEFAULT '',
    discord_webhook_url TEXT NOT NULL DEFAULT '',
    smtp_host           TEXT NOT NULL DEFAULT '',
    smtp_port           INTEGER NOT NULL DEFAULT 587,
    smtp_username       TEXT NOT NULL DEFAULT '',
    smtp_password       TEXT NOT NULL DEFAULT '',
    smtp_from           TEXT NOT NULL DEFAULT '',
    smtp_to             TEXT NOT NULL DEFAULT '',
    notify_on_success   INTEGER NOT NULL DEFAULT 1,
    notify_on_failure   INTEGER NOT NULL DEFAULT 1,
    created_at          TEXT NOT NULL,
    updated_at          TEXT NOT NULL
);

-- Incoming webhook endpoints (one per project; provider = github | gitlab)
CREATE TABLE IF NOT EXISTS webhook_configs (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret     TEXT NOT NULL,   -- HMAC-SHA256 secret used to verify signatures
    provider   TEXT NOT NULL DEFAULT 'github' CHECK (provider IN ('github', 'gitlab')),
    branch     TEXT NOT NULL DEFAULT '',  -- empty = trigger on any branch
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_webhook_project_id ON webhook_configs(project_id);

-- OAuth state tokens (short-lived CSRF prevention)
CREATE TABLE IF NOT EXISTS oauth_states (
    state      TEXT PRIMARY KEY,
    user_id    TEXT,              -- NULL for new-user flows
    provider   TEXT NOT NULL DEFAULT 'github',
    redirect   TEXT NOT NULL DEFAULT '',
    expires_at TEXT NOT NULL
);

-- ─── AI Integration ──────────────────────────────────────────────────────────
-- Per-user LLM configuration (provider, key, model, system prompt, monitoring)
CREATE TABLE IF NOT EXISTS ai_configs (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    provider            TEXT NOT NULL DEFAULT 'openai',
    api_key             TEXT NOT NULL DEFAULT '',
    model               TEXT NOT NULL DEFAULT '',
    base_url            TEXT NOT NULL DEFAULT '',
    system_prompt       TEXT NOT NULL DEFAULT '',
    monitoring_enabled  INTEGER NOT NULL DEFAULT 0,
    monitoring_interval INTEGER NOT NULL DEFAULT 300,
    created_at          TEXT NOT NULL,
    updated_at          TEXT NOT NULL
);

-- Custom knowledge base documents for RAG (Retrieval-Augmented Generation)
CREATE TABLE IF NOT EXISTS rag_documents (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_rag_user_id ON rag_documents(user_id);

-- AI monitoring alerts generated by the background analysis service
CREATE TABLE IF NOT EXISTS ai_monitor_alerts (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    severity      TEXT NOT NULL DEFAULT 'info',
    title         TEXT NOT NULL,
    message       TEXT NOT NULL,
    resolved      INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_monitor_alerts_user_id      ON ai_monitor_alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_monitor_alerts_deployment_id ON ai_monitor_alerts(deployment_id);
CREATE INDEX IF NOT EXISTS idx_monitor_alerts_created_at   ON ai_monitor_alerts(created_at DESC);

-- ─── AI Token Usage Tracking ─────────────────────────────────────────────────
-- Daily usage counter per user, used for global key rate limiting.
CREATE TABLE IF NOT EXISTS ai_token_usage (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date       TEXT NOT NULL,  -- YYYY-MM-DD UTC
    calls      INTEGER NOT NULL DEFAULT 0,
    tokens     INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL,
    CONSTRAINT uq_ai_usage_user_date UNIQUE (user_id, date)
);
CREATE INDEX IF NOT EXISTS idx_ai_usage_user_date ON ai_token_usage(user_id, date);

-- ─── Kubernetes Integration ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS k8s_configs (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL DEFAULT 'default',
    server_url   TEXT NOT NULL DEFAULT '',
    token        TEXT NOT NULL DEFAULT '',
    namespace    TEXT NOT NULL DEFAULT 'pushpaka',
    kubeconfig   TEXT NOT NULL DEFAULT '',
    registry_url TEXT NOT NULL DEFAULT '',
    enabled      INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);
`

// NewSQLite opens (or creates) a SQLite database at path and applies the schema.
func NewSQLite(path string) (*sqlx.DB, error) {
	if path == "" {
		path = "pushpaka-dev.db"
	}
	// _busy_timeout: retry for up to 5 s instead of immediately returning
	// SQLITE_BUSY when another connection holds a write lock. This is essential
	// in all-in-one mode where the API and the embedded worker share the same
	// SQLite file and both run schema setup on startup.
	// _journal_mode=WAL: allows concurrent reads alongside writes and reduces
	// contention significantly.
	dsn := "file:" + path + "?_busy_timeout=5000&_journal_mode=WAL"
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite allows only one concurrent writer.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	// Enable foreign keys; WAL is already set via DSN but we also run it via
	// PRAGMA to ensure it is active for this connection.
	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec(sqliteSchema); err != nil {
		return nil, err
	}
	// Idempotent column additions for databases created before schema updates.
	migrations := []string{
		`ALTER TABLE deployments ADD COLUMN external_port INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN is_private INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN git_token TEXT NOT NULL DEFAULT ''`,
		// Resource limits added in v2.0
		`ALTER TABLE projects ADD COLUMN cpu_limit TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN memory_limit TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN restart_policy TEXT NOT NULL DEFAULT 'unless-stopped'`,
		// Kubernetes + multi-target deployment added in v3.0
		`ALTER TABLE projects ADD COLUMN deploy_target TEXT NOT NULL DEFAULT 'docker'`,
		`ALTER TABLE projects ADD COLUMN k8s_namespace TEXT NOT NULL DEFAULT ''`,
		// Build configuration added in v3.1
		`ALTER TABLE projects ADD COLUMN install_command TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN run_dir TEXT NOT NULL DEFAULT ''`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			// "duplicate column name" is expected on an up-to-date database -- ignore it.
			if !strings.Contains(err.Error(), "duplicate column name") {
				return nil, err
			}
		}
	}
	log.Info().Str("path", path).Msg("connected to sqlite")
	return db, nil
}
