package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewHealthHandler(db *sqlx.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
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
