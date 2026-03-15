package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               string
	DatabaseDriver     string // "postgres" (default) or "sqlite"
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	JWTExpiry          int // hours
	AppEnv             string
	AllowedOrigins     []string
	DockerHost         string
	TraefikNetwork     string
	GithubClientID     string
	GithubClientSecret string
	BuildDir           string
	LogLevel           string
}

func Load() *Config {
	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseDriver:     getEnv("DATABASE_DRIVER", "postgres"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://pushpaka:pushpaka@localhost:5432/pushpaka?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", ""), // empty = Redis disabled (deployment triggers unavailable)
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:          jwtExpiry,
		AppEnv:             getEnv("APP_ENV", "development"),
		AllowedOrigins:     []string{getEnv("CORS_ORIGINS", "http://localhost:3000")},
		DockerHost:         getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		TraefikNetwork:     getEnv("TRAEFIK_NETWORK", "pushpaka-network"),
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		BuildDir:           getEnv("BUILD_DIR", "/tmp/pushpaka-builds"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
