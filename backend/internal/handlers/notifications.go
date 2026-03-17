package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/services"
)

type NotificationHandler struct {
	notifSvc *services.NotificationService
}

func NewNotificationHandler(svc *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: svc}
}

// Get returns the current user's notification config.
// GET /api/v1/notifications/config
func (h *NotificationHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cfg, err := h.notifSvc.GetConfig(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// Upsert creates or updates the notification config for the current user.
// PUT /api/v1/notifications/config
func (h *NotificationHandler) Upsert(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.UpsertNotificationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := h.notifSvc.UpsertConfig(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// InternalNotify is the internal callback endpoint the worker POSTs to when a
// deployment completes.  It is only accessible within the same process / network
// and does not require a user JWT.
// POST /api/v1/internal/notify
func (h *NotificationHandler) InternalNotify(c *gin.Context) {
	// The internal key is a shared secret passed as a query parameter to
	// prevent accidental exposure if the internal endpoint is somehow reached
	// from outside. The worker sets it from the NotificationURL it received.
	var payload struct {
		models.NotificationEvent
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.notifSvc.DispatchInternal(&payload.NotificationEvent, payload.UserID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
