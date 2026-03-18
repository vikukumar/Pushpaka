package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"time"
)

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string // disable, require, verify-ca, verify-full
	SSLCertFile     string // Path to CA certificate file
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	// Connection pool settings
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	MaxConnAge   time.Duration
	PoolTimeout  time.Duration
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// LoadDatabaseConfig builds DatabaseConfig from environment variables
// with sensible defaults for the given mode (development, staging, production)
func LoadDatabaseConfig(mode string) *DatabaseConfig {
	cfg := &DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            parseIntEnv("DB_PORT", 5432),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "pushpaka"),
		SSLMode:         getEnv("DB_SSL_MODE", sslModeDefault(mode)),
		SSLCertFile:     getEnv("DB_SSL_CERT_FILE", ""),
		MaxOpenConns:    parseIntEnv("DB_MAX_OPEN_CONNS", connPoolSize(mode, "open")),
		MaxIdleConns:    parseIntEnv("DB_MAX_IDLE_CONNS", connPoolSize(mode, "idle")),
		ConnMaxLifetime: parseDurationEnv("DB_CONN_MAX_LIFETIME", connLifetime(mode)),
	}
	return cfg
}

// LoadRedisConfig builds RedisConfig from environment variables
// with sensible defaults for the given mode
func LoadRedisConfig(mode string) *RedisConfig {
	cfg := &RedisConfig{
		Host:         getEnv("REDIS_HOST", "localhost"),
		Port:         parseIntEnv("REDIS_PORT", 6379),
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           parseIntEnv("REDIS_DB", 0),
		MaxRetries:   parseIntEnv("REDIS_MAX_RETRIES", 3),
		PoolSize:     parseIntEnv("REDIS_POOL_SIZE", redisPoolSize(mode)),
		MinIdleConns: parseIntEnv("REDIS_MIN_IDLE_CONNS", minIdleConns(mode)),
		MaxConnAge:   parseDurationEnv("REDIS_MAX_CONN_AGE", 10*time.Minute),
		PoolTimeout:  parseDurationEnv("REDIS_POOL_TIMEOUT", 4*time.Second),
		IdleTimeout:  parseDurationEnv("REDIS_IDLE_TIMEOUT", 5*time.Minute),
		ReadTimeout:  parseDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: parseDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
	}
	return cfg
}

// BuildPostgresURL constructs a PostgreSQL DSN from the config
func (c *DatabaseConfig) BuildPostgresURL() string {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.SSLMode,
	)

	// Add certificate path if SSL is enabled and cert file is specified
	if c.SSLMode != "disable" && c.SSLCertFile != "" {
		dsn += "&sslrootcert=" + c.SSLCertFile
	}

	return dsn
}

// LoadSSLCertificate loads the SSL certificate for database connections
// Returns nil if SSL is disabled or cert file is not specified
func (c *DatabaseConfig) LoadSSLCertificate() (*tls.Config, error) {
	if c.SSLMode == "disable" || c.SSLCertFile == "" {
		return nil, nil
	}

	// Read certificate file
	certPEM, err := os.ReadFile(c.SSLCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSL certificate file %q: %w", c.SSLCertFile, err)
	}

	// Create certificate pool - start with system certs if available
	var caCertPool *x509.CertPool
	if cp, err := x509.SystemCertPool(); err == nil {
		caCertPool = cp
	} else {
		// Fallback: create new cert pool
		caCertPool = x509.NewCertPool()
	}

	if !caCertPool.AppendCertsFromPEM(certPEM) {
		return nil, fmt.Errorf("failed to append certificate to pool")
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.SSLMode == "require", // require mode doesn't validate hostname
	}

	return tlsConfig, nil
}

// BuildRedisURL constructs a Redis URL from the config
// Format: redis://[user:password@]host:port/db
func (c *RedisConfig) BuildRedisURL() string {
	if c.Password != "" {
		return fmt.Sprintf(
			"redis://default:%s@%s:%d/%d",
			c.Password,
			c.Host,
			c.Port,
			c.DB,
		)
	}
	return fmt.Sprintf(
		"redis://%s:%d/%d",
		c.Host,
		c.Port,
		c.DB,
	)
}

// Helper functions for defaults based on mode
func sslModeDefault(mode string) string {
	switch mode {
	case "production":
		return "require"
	case "staging":
		return "prefer"
	default: // development
		return "disable"
	}
}

func connPoolSize(mode, connType string) int {
	switch mode {
	case "production":
		if connType == "open" {
			return 30
		}
		return 10
	case "staging":
		if connType == "open" {
			return 20
		}
		return 5
	default: // development
		if connType == "open" {
			return 10
		}
		return 3
	}
}

func connLifetime(mode string) time.Duration {
	switch mode {
	case "production":
		return 5 * time.Minute
	case "staging":
		return 10 * time.Minute
	default: // development
		return 15 * time.Minute
	}
}

func redisPoolSize(mode string) int {
	switch mode {
	case "production":
		return 20
	case "staging":
		return 10
	default: // development
		return 5
	}
}

func minIdleConns(mode string) int {
	switch mode {
	case "production":
		return 5
	case "staging":
		return 2
	default: // development
		return 1
	}
}

// Parser helper functions
func parseIntEnv(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	v, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return v
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	v, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return v
}
