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
`

// NewSQLite opens (or creates) a SQLite database at path and applies the schema.
func NewSQLite(path string) (*sqlx.DB, error) {
	if path == "" {
		path = "pushpaka-dev.db"
	}
	db, err := sqlx.Open("sqlite", "file:"+path)
	if err != nil {
		return nil, err
	}
	// SQLite allows only one concurrent writer.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	// Enable foreign keys and WAL mode, then apply schema (all idempotent).
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
