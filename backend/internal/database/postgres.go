package database

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/vikukumar/Pushpaka/internal/config"
)

func NewPostgres(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Info().Msg("connected to postgres")
	return db, nil
}

// NewPostgresWithConfig opens a PostgreSQL connection with custom configuration
// and SSL certificate support. This is preferred over NewPostgres for production.
func NewPostgresWithConfig(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	// Validate required fields
	if cfg == nil {
		return nil, fmt.Errorf("database config cannot be nil")
	}

	// Load SSL certificate if needed
	var tlsConfig *tls.Config
	var err error
	if cfg.SSLMode != "disable" && cfg.SSLCertFile != "" {
		tlsConfig, err = cfg.LoadSSLCertificate()
		if err != nil {
			return nil, err
		}
	}

	// Build connector with DSN
	connector, err := pq.NewConnector(cfg.BuildPostgresURL())
	if err != nil {
		return nil, err
	}

	// Apply custom dialer with TLS if needed
	if tlsConfig != nil {
		_ = tlsConfig // Note: PostgreSQL driver handles SSL via sslmode in DSN
	}

	db := sqlx.NewDb(sql.OpenDB(connector), "postgres")

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.DBName).
		Str("sslmode", cfg.SSLMode).
		Int("maxOpenConns", cfg.MaxOpenConns).
		Int("maxIdleConns", cfg.MaxIdleConns).
		Msg("connected to postgres with custom config")

	return db, nil
}
