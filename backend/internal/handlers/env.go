package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yourusername/pushpaka/internal/middleware"
	"github.com/yourusername/pushpaka/internal/models"
	"github.com/yourusername/pushpaka/internal/services"
)

type EnvHandler struct {
	envSvc *services.EnvService
}

func NewEnvHandler(envSvc *services.EnvService) *EnvHandler {
	return &EnvHandler{envSvc: envSvc}
}

func (h *EnvHandler) Set(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.SetEnvVarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ev, err := h.envSvc.Set(userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set environment variable"})
		return
	}
	c.JSON(http.StatusOK, ev)
}

func (h *EnvHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id query param required"})
		return
	}

	vars, err := h.envSvc.List(projectID, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch env vars"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vars})
}

func (h *EnvHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.DeleteEnvVarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.envSvc.Delete(req.ProjectID, req.Key, userID); err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete env var"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "env var deleted"})
}
