package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/vikukumar/Pushpaka/internal/services"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate the origin against allowed origins
		return true
	},
}

type LogHandler struct {
	logSvc *services.LogService
}

func NewLogHandler(logSvc *services.LogService) *LogHandler {
	return &LogHandler{logSvc: logSvc}
}

// GetLogs returns all logs for a deployment (REST)
func (h *LogHandler) GetLogs(c *gin.Context) {
	deploymentID := c.Param("id")
	logs, err := h.logSvc.GetByDeployment(deploymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// StreamLogs streams logs for a deployment over WebSocket
func (h *LogHandler) StreamLogs(c *gin.Context) {
	deploymentID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	// Send existing logs first
	existingLogs, err := h.logSvc.GetByDeployment(deploymentID)
	if err == nil {
		for _, l := range existingLogs {
			if err := conn.WriteJSON(l); err != nil {
				return
			}
		}
	}

	// Poll for new logs every 500ms (simple implementation)
	// In production, use pub/sub (Redis) to push logs in real-time
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	lastCount := len(existingLogs)

	for range ticker.C {
		logs, err := h.logSvc.GetByDeployment(deploymentID)
		if err != nil {
			continue
		}
		if len(logs) > lastCount {
			for _, l := range logs[lastCount:] {
				if err := conn.WriteJSON(l); err != nil {
					return
				}
			}
			lastCount = len(logs)
		}
	}
}
