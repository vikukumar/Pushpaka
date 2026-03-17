package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/internal/services"
)

type AIHandler struct {
	aiSvc      *services.AIService
	logRepo    *repositories.LogRepository
	deployRepo *repositories.DeploymentRepository
}

func NewAIHandler(aiSvc *services.AIService, logRepo *repositories.LogRepository, deployRepo *repositories.DeploymentRepository) *AIHandler {
	return &AIHandler{aiSvc: aiSvc, logRepo: logRepo, deployRepo: deployRepo}
}

// AnalyzeLogs retrieves deployment logs and sends them to the AI for analysis.
// POST /api/v1/deployments/:id/analyze
func (h *AIHandler) AnalyzeLogs(c *gin.Context) {
	if !h.aiSvc.Available() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI integration not configured on this server (set AI_API_KEY)",
		})
		return
	}

	userID := middleware.GetUserID(c)
	deploymentID := c.Param("id")

	deployment, err := h.deployRepo.FindByID(deploymentID)
	if err != nil || deployment.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	logEntries, err := h.logRepo.FindByDeploymentID(deploymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve logs"})
		return
	}

	if len(logEntries) == 0 {
		c.JSON(http.StatusOK, gin.H{"analysis": "No logs found for this deployment."})
		return
	}

	// Assemble plain-text log dump
	var sb strings.Builder
	for _, entry := range logEntries {
		sb.WriteString(entry.Message)
		sb.WriteByte('\n')
	}

	analysis, err := h.aiSvc.AnalyzeLogs(sb.String())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}
