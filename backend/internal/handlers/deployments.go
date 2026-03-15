package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/services"
)

type DeploymentHandler struct {
	deploymentSvc *services.DeploymentService
}

func NewDeploymentHandler(deploymentSvc *services.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{deploymentSvc: deploymentSvc}
}

func (h *DeploymentHandler) Deploy(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deployment, err := h.deploymentSvc.Trigger(userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to trigger deployment"})
		return
	}
	c.JSON(http.StatusCreated, deployment)
}

func (h *DeploymentHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	deployments, err := h.deploymentSvc.List(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch deployments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": deployments})
}

func (h *DeploymentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	deployment, err := h.deploymentSvc.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrDeploymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch deployment"})
		return
	}
	c.JSON(http.StatusOK, deployment)
}

func (h *DeploymentHandler) Rollback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	deployment, err := h.deploymentSvc.Rollback(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rollback failed"})
		return
	}
	c.JSON(http.StatusCreated, deployment)
}
