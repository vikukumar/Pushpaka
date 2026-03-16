package handlers

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// WorkerStatsProvider is implemented by the in-process queue to expose live worker counts.
type WorkerStatsProvider interface {
	TotalWorkers() int
	ActiveJobs() int
}

type HealthHandler struct {
	db          *sqlx.DB
	rdb         *redis.Client
	workerStats WorkerStatsProvider
}

func NewHealthHandler(db *sqlx.DB, rdb *redis.Client, ws WorkerStatsProvider) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb, workerStats: ws}
}

func (h *HealthHandler) Health(c *gin.Context) {
	status := "ok"
	dbOK := true
	redisOK := true

	if err := h.db.Ping(); err != nil {
		dbOK = false
		status = "degraded"
	}

	if h.rdb == nil {
		redisOK = false
	} else if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
		redisOK = false
		status = "degraded"
	}

	code := http.StatusOK
	if status != "ok" {
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"status":  status,
		"version": "v1.0.0",
		"checks": gin.H{
			"database": dbOK,
			"redis":    redisOK,
		},
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// Metrics returns the Prometheus handler wrapped for Gin
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// System returns a live snapshot of system capabilities and worker status.
func (h *HealthHandler) System(c *gin.Context) {
	// Docker availability
	dockerAvailable, dockerHost := checkDockerAvailable()

	// Git availability
	gitAvailable := false
	gitVersion := ""
	if out, err := runWithTimeout("git", "--version"); err == nil {
		gitAvailable = true
		gitVersion = strings.TrimSpace(out)
	}

	// Running inside a container?
	inContainer := isRunningInContainer()

	// Queue mode
	queueMode := "redis"
	if h.rdb == nil {
		queueMode = "in-process"
	}

	// Worker stats
	totalWorkers, activeJobs := 0, 0
	if h.workerStats != nil {
		totalWorkers = h.workerStats.TotalWorkers()
		activeJobs = h.workerStats.ActiveJobs()
	}

	c.JSON(http.StatusOK, gin.H{
		"docker": gin.H{
			"available": dockerAvailable,
			"host":      dockerHost,
		},
		"git": gin.H{
			"available": gitAvailable,
			"version":   gitVersion,
		},
		"workers": gin.H{
			"total":       totalWorkers,
			"active_jobs": activeJobs,
			"idle":        max(0, totalWorkers-activeJobs),
			"queue_mode":  queueMode,
		},
		"runtime": gin.H{
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			"in_container": inContainer,
		},
	})
}

// checkDockerAvailable probes the Docker socket / named pipe and returns
// (available, detectedHost).
func checkDockerAvailable() (bool, string) {
	// Common socket paths
	candidates := []string{
		"/var/run/docker.sock",
		"/run/docker.sock",
	}
	if runtime.GOOS == "windows" {
		candidates = []string{`\\.\pipe\docker_engine`}
	}
	// Also respect DOCKER_HOST env
	if dh := os.Getenv("DOCKER_HOST"); dh != "" {
		h := strings.TrimPrefix(dh, "unix://")
		h = strings.TrimPrefix(h, "npipe://")
		candidates = append([]string{h}, candidates...)
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			// Socket exists — also verify we can connect
			if _, err2 := runWithTimeout("docker", "info"); err2 == nil {
				return true, path
			}
		}
	}

	// Last resort: just try the CLI
	if _, err := runWithTimeout("docker", "info"); err == nil {
		return true, os.Getenv("DOCKER_HOST")
	}
	return false, ""
}

// isRunningInContainer checks common signals that we're inside a container.
func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		return strings.Contains(string(data), "docker") || strings.Contains(string(data), "containerd")
	}
	return false
}

// runWithTimeout runs a command with a 3-second timeout and returns combined output.
func runWithTimeout(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return string(out), err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
