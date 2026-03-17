package database

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite" // registers "sqlite" driver
)

// NewSQLite reuses the same embedded schema as the backend module.
// This is intentionally minimal: the worker needs the same tables as the API.
const sqliteSchema = `
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, name TEXT NOT NULL,
    password_hash TEXT NOT NULL, api_key TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL DEFAULT 'user', created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL,
    repo_url TEXT NOT NULL, branch TEXT NOT NULL DEFAULT 'main',
    build_command TEXT NOT NULL DEFAULT '', start_command TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 3000, framework TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'inactive',
    is_private INTEGER NOT NULL DEFAULT 0, git_token TEXT NOT NULL DEFAULT '',
    cpu_limit TEXT NOT NULL DEFAULT '', memory_limit TEXT NOT NULL DEFAULT '',
    restart_policy TEXT NOT NULL DEFAULT 'unless-stopped',
    deploy_target TEXT NOT NULL DEFAULT 'docker', k8s_namespace TEXT NOT NULL DEFAULT '',
    install_command TEXT NOT NULL DEFAULT '', run_dir TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, user_id TEXT NOT NULL,
    commit_sha TEXT NOT NULL DEFAULT '', commit_msg TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT 'main',
    status TEXT NOT NULL DEFAULT 'queued', image_tag TEXT NOT NULL DEFAULT '',
    container_id TEXT NOT NULL DEFAULT '', url TEXT NOT NULL DEFAULT '',
    error_msg TEXT NOT NULL DEFAULT '', started_at TEXT, finished_at TEXT,
    external_port INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS deployment_logs (
    id TEXT PRIMARY KEY, deployment_id TEXT NOT NULL,
    level TEXT NOT NULL DEFAULT 'info', stream TEXT NOT NULL DEFAULT 'stdout',
    message TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, user_id TEXT NOT NULL,
    domain TEXT NOT NULL UNIQUE, verified INTEGER NOT NULL DEFAULT 0,
    ssl_enabled INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS environment_variables (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, user_id TEXT NOT NULL,
    key TEXT NOT NULL, value TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
    CONSTRAINT uq_env_project_key UNIQUE (project_id, key)
);
`

func NewSQLite(path string) (*sqlx.DB, error) {
	if path == "" {
		path = "pushpaka-dev.db"
	}
	// _busy_timeout: retry for up to 5 s on SQLITE_BUSY instead of failing
	// immediately. Required when the API and embedded worker share the same file
	// in all-in-one / dev mode and both run schema setup concurrently.
	dsn := "file:" + path + "?_busy_timeout=5000&_journal_mode=WAL"
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, err
	}
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
		`ALTER TABLE projects ADD COLUMN is_private INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE projects ADD COLUMN git_token TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN cpu_limit TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN memory_limit TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN restart_policy TEXT NOT NULL DEFAULT 'unless-stopped'`,
		`ALTER TABLE projects ADD COLUMN deploy_target TEXT NOT NULL DEFAULT 'docker'`,
		`ALTER TABLE projects ADD COLUMN k8s_namespace TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN install_command TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE projects ADD COLUMN run_dir TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE deployments ADD COLUMN external_port INTEGER NOT NULL DEFAULT 0`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			if !strings.Contains(err.Error(), "duplicate column name") {
				return nil, err
			}
		}
	}
	log.Info().Str("path", path).Msg("worker connected to sqlite")
	return db, nil
}
