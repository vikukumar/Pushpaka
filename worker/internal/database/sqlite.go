package database

import (
	_ "github.com/glebarez/sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
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
    status TEXT NOT NULL DEFAULT 'inactive', created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, user_id TEXT NOT NULL,
    commit_sha TEXT NOT NULL DEFAULT '', commit_msg TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT 'main',
    status TEXT NOT NULL DEFAULT 'queued', image_tag TEXT NOT NULL DEFAULT '',
    container_id TEXT NOT NULL DEFAULT '', url TEXT NOT NULL DEFAULT '',
    error_msg TEXT NOT NULL DEFAULT '', started_at TEXT, finished_at TEXT,
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
	db, err := sqlx.Open("sqlite", "file:"+path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec(sqliteSchema); err != nil {
		return nil, err
	}
	log.Info().Str("path", path).Msg("worker connected to sqlite")
	return db, nil
}
