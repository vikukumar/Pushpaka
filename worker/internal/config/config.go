package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseDriver string // "postgres" (default) or "sqlite"
	DatabaseURL    string
	RedisURL       string
	DockerHost     string
	TraefikNetwork string
	BuildDir       string
	BuildWorkers   int
	AppEnv         string
}

func Load() *Config {
	workers, _ := strconv.Atoi(getEnv("BUILD_WORKERS", "3"))
	return &Config{
		DatabaseDriver: getEnv("DATABASE_DRIVER", "postgres"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://pushpaka:pushpaka@localhost:5432/pushpaka?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		DockerHost:     getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		TraefikNetwork: getEnv("TRAEFIK_NETWORK", "pushpaka-network"),
		BuildDir:       getEnv("BUILD_DIR", "/tmp/pushpaka-builds"),
		BuildWorkers:   workers,
		AppEnv:         getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
