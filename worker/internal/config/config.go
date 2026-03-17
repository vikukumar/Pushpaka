package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Config struct {
	DatabaseDriver string // "postgres" (default) or "sqlite"
	DatabaseURL    string
	RedisURL       string
	DockerHost     string
	TraefikNetwork string
	// CloneDir is the temporary workspace where repositories are git-cloned
	// and built. Set via BUILD_CLONE_DIR (or legacy BUILD_DIR).
	// Default: <os temp dir>/pushpaka-builds
	CloneDir string
	// DeployDir is the permanent directory used for in-place (no-Docker)
	// deployments -- each project gets its own subdirectory here.
	// Set via BUILD_DEPLOY_DIR.
	// Default: /deploy/pushpaka  (Linux/Mac) | %LOCALAPPDATA%\pushpaka\deploy  (Windows)
	DeployDir    string
	BuildWorkers int
	AppEnv       string
}

func Load() *Config {
	workers, _ := strconv.Atoi(getEnv("BUILD_WORKERS", "3"))

	// CloneDir: prefer BUILD_CLONE_DIR, fall back to legacy BUILD_DIR, then platform default.
	cloneDir := getEnv("BUILD_CLONE_DIR", "")
	if cloneDir == "" {
		cloneDir = getEnv("BUILD_DIR", defaultCloneDir())
	}

	deployDir := getEnv("BUILD_DEPLOY_DIR", defaultDeployDir())

	return &Config{
		DatabaseDriver: getEnv("DATABASE_DRIVER", "postgres"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://pushpaka:pushpaka@localhost:5432/pushpaka?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		DockerHost:     getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		TraefikNetwork: getEnv("TRAEFIK_NETWORK", "pushpaka-network"),
		CloneDir:       cloneDir,
		DeployDir:      deployDir,
		BuildWorkers:   workers,
		AppEnv:         getEnv("APP_ENV", "development"),
	}
}

// EnsureDirs creates CloneDir and DeployDir (and their parents) if they do
// not already exist. Call this once at worker startup before processing jobs.
func (c *Config) EnsureDirs() error {
	for _, dir := range []string{c.CloneDir, c.DeployDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("pushpaka: cannot create directory %q: %w", dir, err)
		}
	}
	return nil
}

// defaultCloneDir returns a platform-appropriate temp directory for git clones.
func defaultCloneDir() string {
	return filepath.Join(os.TempDir(), "pushpaka-builds")
}

// defaultDeployDir returns a platform-appropriate permanent directory for
// running in-place (no-Docker) deployments.
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
