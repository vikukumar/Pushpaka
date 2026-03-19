package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/internal/services"
)

type WebhookHandler struct {
	webhookSvc *services.WebhookService
}

func NewWebhookHandler(svc *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookSvc: svc}
}

// Create creates a new incoming webhook endpoint for a project.
// POST /api/v1/webhooks
func (h *WebhookHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.webhookSvc.Create(userID, &req)
	if err != nil {
		if err == services.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook"})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// List returns all webhooks owned by the authenticated user.
// GET /api/v1/webhooks
func (h *WebhookHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	webhooks, err := h.webhookSvc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list webhooks"})
		return
	}
	c.JSON(http.StatusOK, webhooks)
}

// Delete removes a webhook endpoint.
// DELETE /api/v1/webhooks/:id
func (h *WebhookHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.webhookSvc.Delete(c.Param("id"), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete webhook"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Receive is the public endpoint that GitHub/GitLab POSTs push events to.
// POST /api/v1/webhooks/:id/receive
// This endpoint is intentionally public (no JWT required).
func (h *WebhookHandler) Receive(c *gin.Context) {
	webhookID := c.Param("id")

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20)) // 1 MiB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// GitHub uses X-Hub-Signature-256; GitLab uses X-Gitlab-Token
	sig := c.GetHeader("X-Hub-Signature-256")
	if sig == "" {
		sig = c.GetHeader("X-Gitlab-Token")
	}

	if err := h.webhookSvc.Receive(webhookID, body, sig); err != nil {
		switch err {
		case services.ErrWebhookNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		case services.ErrWebhookSignatureInvalid:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
