package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/internal/services"
	"github.com/vikukumar/Pushpaka/pkg/tunnel"
)

var workerWSUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for worker sync
	},
}

type WorkerHandler struct {
	svc *services.WorkerNodeService
}

func NewWorkerHandler(svc *services.WorkerNodeService) *WorkerHandler {
	return &WorkerHandler{svc: svc}
}

// Register endpoint allows new worker initialization via PAT
func (h *WorkerHandler) Register(c *gin.Context) {
	var req models.RegisterWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	res, err := h.svc.RegisterWorker(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed or invalid zone PAT"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ConnectWS upgrades to a websocket stream for continuous sync
func (h *WorkerHandler) ConnectWS(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.Query("token")
	}

	worker, err := h.svc.Authenticate(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid worker token"})
		return
	}

	ws, err := workerWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade websocket connection"})
		return
	}

	// Register this websocket as a multiplexed Yamux tunnel stream
	session, err := tunnel.GlobalManager.Register(worker.ID, ws)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize yamux tunnel session"})
		return
	}

	// This blocks until the session drops
	<-session.CloseChan()
}

// Poll is an HTTP fallback alternative for sync
func (h *WorkerHandler) Poll(c *gin.Context) {
	token := c.GetHeader("Authorization")
	worker, err := h.svc.Authenticate(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid worker token"})
		return
	}

	// Future: Fetch pending deployment jobs for this worker, and emit queued tasks
	// Currently returning heartbeat OK
	c.JSON(http.StatusOK, gin.H{
		"worker_id": worker.ID,
		"status":    "active",
		"pending":   0,
	})
}

// ListNodes returns all registered worker nodes for the admin dashboard
func (h *WorkerHandler) ListNodes(c *gin.Context) {
	workers, err := h.svc.ListWorkers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workers"})
		return
	}
	c.JSON(http.StatusOK, workers)
}

// GetZonePAT retrieves the global installation Zone PAT
func (h *WorkerHandler) GetZonePAT(c *gin.Context) {
	pat, err := h.svc.GetZonePAT()
	if err != nil || pat == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "zone PAT not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"zone_pat": pat})
}


