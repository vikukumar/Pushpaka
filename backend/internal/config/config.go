package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	DatabaseDriver     string // "postgres" (default) or "sqlite"
	DatabaseURL        string
	DatabaseConfig     *DatabaseConfig // Structured database config (alternative to DSN)
	RedisURL           string
	RedisConfig        *RedisConfig // Structured Redis config (alternative to URL)
	JWTSecret          string
	JWTExpiry          int // hours
	AppEnv             string
	AppMode            string // development, staging, production
	BaseURL            string // public base URL, e.g. https://push.example.com
	AllowedOrigins     []string
	DockerHost         string
	TraefikNetwork     string
	GithubClientID     string
	GithubClientSecret string
	// GitLab OAuth (also supports self-hosted instances via GitlabBaseURL)
	GitlabClientID     string
	GitlabClientSecret string
	GitlabBaseURL      string // default: https://gitlab.com
	// CloneDir is the temporary workspace where repositories are git-cloned.
	// Set via BUILD_CLONE_DIR (or legacy BUILD_DIR).
	CloneDir string
	// DeployDir is the permanent directory for in-place (no-Docker) deployments.
	// Set via BUILD_DEPLOY_DIR.
	DeployDir string
	LogLevel  string

	// AI integration (OpenAI-compatible API)
	AIProvider string // "openai", "openrouter", "gemini", "anthropic", "ollama"
	AIAPIKey   string
	AIModel    string // e.g. "gpt-4o-mini", "claude-3-haiku", "gemini-pro"
	AIBaseURL  string // override endpoint for OpenRouter / Ollama / self-hosted
	// AIRateLimitPerUserPerDay caps daily AI calls per user when they are using the
	// global platform key (not their own). 0 = unlimited.
	AIRateLimitPerUserPerDay int

	// Default notification channels (can be overridden per-user in DB)
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func Load() *Config {
	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	cloneDir := getEnv("BUILD_CLONE_DIR", "")
	if cloneDir == "" {
		cloneDir = getEnv("BUILD_DIR", defaultCloneDir())
	}
	deployDir := getEnv("BUILD_DEPLOY_DIR", defaultDeployDir())

	appMode := normalizeMode(getEnv("APP_ENV", "development"))

	// Build DatabaseURL from structured config if DATABASE_URL not explicitly set
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" && getEnv("DATABASE_DRIVER", "postgres") == "postgres" {
		dbCfg := LoadDatabaseConfig(appMode)
		databaseURL = dbCfg.BuildPostgresURL()
	}

	// Build RedisURL from structured config if REDIS_URL not explicitly set
	redisURL := getEnv("REDIS_URL", "")
	if redisURL == "" && getEnv("REDIS_HOST", "") != "" {
		redisCfg := LoadRedisConfig(appMode)
		redisURL = redisCfg.BuildRedisURL()
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseDriver:     getEnv("DATABASE_DRIVER", "postgres"),
		DatabaseURL:        databaseURL,
		DatabaseConfig:     LoadDatabaseConfig(appMode),
		RedisURL:           redisURL,
		RedisConfig:        LoadRedisConfig(appMode),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:          jwtExpiry,
		AppEnv:             getEnv("APP_ENV", "development"),
		AppMode:            appMode,
		BaseURL:            getEnv("BASE_URL", "http://localhost:"+getEnv("PORT", "8080")),
		AllowedOrigins:     strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
		DockerHost:         getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		TraefikNetwork:     getEnv("TRAEFIK_NETWORK", "pushpaka-network"),
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitlabClientID:     getEnv("GITLAB_CLIENT_ID", ""),
		GitlabClientSecret: getEnv("GITLAB_CLIENT_SECRET", ""),
		GitlabBaseURL:      getEnv("GITLAB_BASE_URL", "https://gitlab.com"),
		CloneDir:           cloneDir,
		DeployDir:          deployDir,
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		AIProvider:         getEnv("AI_PROVIDER", "openai"),
		AIAPIKey:           getEnv("AI_API_KEY", ""),
		AIModel:            getEnv("AI_MODEL", "gpt-4o-mini"),
		AIBaseURL:          getEnv("AI_BASE_URL", ""),
		AIRateLimitPerUserPerDay: func() int {
			v, _ := strconv.Atoi(getEnv("AI_RATE_LIMIT_PER_USER_PER_DAY", "0"))
			return v
		}(),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}
}

// EnsureDirs creates CloneDir and DeployDir if they do not already exist.
func (c *Config) EnsureDirs() error {
	for _, dir := range []string{c.CloneDir, c.DeployDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("pushpaka: cannot create directory %q: %w", dir, err)
		}
	}
	return nil
}

func defaultCloneDir() string {
	return filepath.Join(os.TempDir(), "pushpaka-builds")
}

func defaultDeployDir() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = os.Getenv("USERPROFILE")
		}
		if base == "" {
			base = `C:\pushpaka`
		}
		return filepath.Join(base, "pushpaka", "deploy")
	}
	return "/deploy/pushpaka"
}

// normalizeMode converts app environment names to standard modes
func normalizeMode(env string) string {
	switch strings.ToLower(env) {
	case "prod", "production":
		return "production"
	case "stage", "staging":
		return "staging"
	default:
		return "development"
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
