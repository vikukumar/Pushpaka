package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	Common      commonConfig `yaml:"common"`
	Development modeConfig   `yaml:"development"`
	Staging     modeConfig   `yaml:"staging"`
	Production  modeConfig   `yaml:"production"`
}

type modeConfig struct {
	Server    serverConfig   `yaml:"server"`
	Database  databaseConfig `yaml:"database"`
	Redis     redisConfig    `yaml:"redis"`
	Component string         `yaml:"component"`
}

type commonConfig struct {
	JWTSecret      string       `yaml:"jwt_secret"`
	JWTExpiryHours int          `yaml:"jwt_expiry_hours"`
	BaseURL        string       `yaml:"base_url"`
	CORSOrigins    string       `yaml:"cors_origins"`
	Build          buildConfig  `yaml:"build"`
	GitHub         oauthConfig  `yaml:"github"`
	GitLab         gitlabConfig `yaml:"gitlab"`
	AI             aiConfig     `yaml:"ai"`
	SMTP           smtpConfig   `yaml:"smtp"`
	TraefikNetwork string       `yaml:"traefik_network"`
	DockerHost     string       `yaml:"docker_host"`
}

type serverConfig struct {
	Port        int    `yaml:"port"`
	LogLevel    string `yaml:"log_level"`
	Environment string `yaml:"environment"`
}

type databaseConfig struct {
	Driver          string `yaml:"driver"`
	Path            string `yaml:"path"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	SSLCertFile     string `yaml:"ssl_cert_file"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

type redisConfig struct {
	Enabled      *bool  `yaml:"enabled"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	MaxRetries   int    `yaml:"max_retries"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
	MaxConnAge   string `yaml:"max_conn_age"`
	PoolTimeout  string `yaml:"pool_timeout"`
	IdleTimeout  string `yaml:"idle_timeout"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
}

type buildConfig struct {
	CloneDir  string `yaml:"clone_dir"`
	DeployDir string `yaml:"deploy_dir"`
	Workers   int    `yaml:"workers"`
}

type oauthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type gitlabConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	BaseURL      string `yaml:"base_url"`
}

type aiConfig struct {
	Provider               string `yaml:"provider"`
	APIKey                 string `yaml:"api_key"`
	Model                  string `yaml:"model"`
	BaseURL                string `yaml:"base_url"`
	RateLimitPerUserPerDay int    `yaml:"rate_limit_per_user_per_day"`
}

type smtpConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

func applyConfigFile(path, mode string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}

	var cfg fileConfig
	if err := yaml.Unmarshal([]byte(os.ExpandEnv(string(raw))), &cfg); err != nil {
		return fmt.Errorf("parse config file %q: %w", path, err)
	}

	selected, err := cfg.mode(mode)
	if err != nil {
		return err
	}

	applyCommonConfig(cfg.Common)
	applyModeConfig(selected)
	os.Setenv("PUSHPAKA_CONFIG_FILE", path)
	return nil
}

func (c fileConfig) mode(mode string) (modeConfig, error) {
	switch normalizeConfigMode(mode) {
	case "production":
		return c.Production, nil
	case "staging":
		return c.Staging, nil
	case "development":
		return c.Development, nil
	default:
		return modeConfig{}, fmt.Errorf("unsupported config mode %q", mode)
	}
}

func normalizeConfigMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "prod", "production":
		return "production"
	case "stage", "staging":
		return "staging"
	default:
		return "development"
	}
}

func applyCommonConfig(cfg commonConfig) {
	setIfNotEmpty("JWT_SECRET", cfg.JWTSecret)
	setIfPositive("JWT_EXPIRY_HOURS", cfg.JWTExpiryHours)
	setIfNotEmpty("BASE_URL", cfg.BaseURL)
	setIfNotEmpty("CORS_ORIGINS", cfg.CORSOrigins)
	setIfNotEmpty("BUILD_CLONE_DIR", cfg.Build.CloneDir)
	setIfNotEmpty("BUILD_DEPLOY_DIR", cfg.Build.DeployDir)
	setIfPositive("BUILD_WORKERS", cfg.Build.Workers)
	setIfNotEmpty("DOCKER_HOST", cfg.DockerHost)
	setIfNotEmpty("TRAEFIK_NETWORK", cfg.TraefikNetwork)

	setIfNotEmpty("GITHUB_CLIENT_ID", cfg.GitHub.ClientID)
	setIfNotEmpty("GITHUB_CLIENT_SECRET", cfg.GitHub.ClientSecret)
	setIfNotEmpty("GITLAB_CLIENT_ID", cfg.GitLab.ClientID)
	setIfNotEmpty("GITLAB_CLIENT_SECRET", cfg.GitLab.ClientSecret)
	setIfNotEmpty("GITLAB_BASE_URL", cfg.GitLab.BaseURL)

	setIfNotEmpty("AI_PROVIDER", cfg.AI.Provider)
	setIfNotEmpty("AI_API_KEY", cfg.AI.APIKey)
	setIfNotEmpty("AI_MODEL", cfg.AI.Model)
	setIfNotEmpty("AI_BASE_URL", cfg.AI.BaseURL)
	setIfPositive("AI_RATE_LIMIT_PER_USER_PER_DAY", cfg.AI.RateLimitPerUserPerDay)

	setIfNotEmpty("SMTP_HOST", cfg.SMTP.Host)
	setIfPositive("SMTP_PORT", cfg.SMTP.Port)
	setIfNotEmpty("SMTP_USERNAME", cfg.SMTP.Username)
	setIfNotEmpty("SMTP_PASSWORD", cfg.SMTP.Password)
	setIfNotEmpty("SMTP_FROM", cfg.SMTP.From)
}

func applyModeConfig(cfg modeConfig) {
	if cfg.Server.Port > 0 {
		os.Setenv("PORT", fmt.Sprintf("%d", cfg.Server.Port))
	}
	setIfNotEmpty("LOG_LEVEL", cfg.Server.LogLevel)
	setIfNotEmpty("APP_ENV", cfg.Server.Environment)
	setIfNotEmpty("PUSHPAKA_COMPONENT", cfg.Component)

	applyDatabaseConfig(cfg.Database)
	applyRedisConfig(cfg.Redis)
}

func applyDatabaseConfig(cfg databaseConfig) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		return
	}

	os.Setenv("DATABASE_DRIVER", driver)

	switch driver {
	case "sqlite":
		if cfg.Path != "" {
			os.Setenv("DATABASE_URL", cfg.Path)
		}
	case "postgres":
		dsn := buildPostgresDSN(cfg)
		if dsn != "" {
			os.Setenv("DATABASE_URL", dsn)
		}
	}

	setIfNotEmpty("DB_HOST", cfg.Host)
	setIfPositive("DB_PORT", cfg.Port)
	setIfNotEmpty("DB_USER", cfg.User)
	setIfNotEmpty("DB_PASSWORD", cfg.Password)
	setIfNotEmpty("DB_NAME", cfg.Name)
	setIfNotEmpty("DB_SSL_MODE", cfg.SSLMode)
	setIfNotEmpty("DB_SSL_CERT_FILE", cfg.SSLCertFile)
	setIfPositive("DB_MAX_OPEN_CONNS", cfg.MaxOpenConns)
	setIfPositive("DB_MAX_IDLE_CONNS", cfg.MaxIdleConns)
	setIfNotEmpty("DB_CONN_MAX_LIFETIME", cfg.ConnMaxLifetime)
}

func applyRedisConfig(cfg redisConfig) {
	if cfg.Enabled != nil && !*cfg.Enabled {
		os.Setenv("REDIS_ENABLED", "false")
		os.Setenv("REDIS_URL", "")
		os.Setenv("REDIS_HOST", "")
		return
	}

	if cfg.Enabled != nil && *cfg.Enabled {
		os.Setenv("REDIS_ENABLED", "true")
	}

	if cfg.Host == "" || cfg.Port == 0 {
		return
	}

	os.Setenv("REDIS_URL", buildRedisURL(cfg))
	setIfNotEmpty("REDIS_HOST", cfg.Host)
	setIfPositive("REDIS_PORT", cfg.Port)
	setIfNotEmpty("REDIS_PASSWORD", cfg.Password)
	setIfPositive("REDIS_DB", cfg.DB)
	setIfPositive("REDIS_MAX_RETRIES", cfg.MaxRetries)
	setIfPositive("REDIS_POOL_SIZE", cfg.PoolSize)
	setIfPositive("REDIS_MIN_IDLE_CONNS", cfg.MinIdleConns)
	setIfNotEmpty("REDIS_MAX_CONN_AGE", cfg.MaxConnAge)
	setIfNotEmpty("REDIS_POOL_TIMEOUT", cfg.PoolTimeout)
	setIfNotEmpty("REDIS_IDLE_TIMEOUT", cfg.IdleTimeout)
	setIfNotEmpty("REDIS_READ_TIMEOUT", cfg.ReadTimeout)
	setIfNotEmpty("REDIS_WRITE_TIMEOUT", cfg.WriteTimeout)
}

func buildPostgresDSN(cfg databaseConfig) string {
	if cfg.Host == "" || cfg.Port == 0 || cfg.User == "" || cfg.Name == "" {
		return ""
	}

	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		sslMode,
	)
	if cfg.SSLCertFile != "" {
		dsn += "&sslrootcert=" + cfg.SSLCertFile
	}
	return dsn
}

func buildRedisURL(cfg redisConfig) string {
	if cfg.Password != "" {
		return fmt.Sprintf("redis://default:%s@%s:%d/%d", cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	}
	return fmt.Sprintf("redis://%s:%d/%d", cfg.Host, cfg.Port, cfg.DB)
}

func setIfNotEmpty(key, value string) {
	if strings.TrimSpace(value) != "" {
		os.Setenv(key, value)
	}
}

func setIfPositive(key string, value int) {
	if value > 0 {
		os.Setenv(key, fmt.Sprintf("%d", value))
	}
}
