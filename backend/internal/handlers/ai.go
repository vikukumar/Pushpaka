package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/internal/services"
)

type AIHandler struct {
	aiSvc        *services.AIService
	logRepo      *repositories.LogRepository
	deployRepo   *repositories.DeploymentRepository
	aiConfigRepo *repositories.AIConfigRepository
}

func NewAIHandler(aiSvc *services.AIService, logRepo *repositories.LogRepository, deployRepo *repositories.DeploymentRepository, aiConfigRepo *repositories.AIConfigRepository) *AIHandler {
	return &AIHandler{aiSvc: aiSvc, logRepo: logRepo, deployRepo: deployRepo, aiConfigRepo: aiConfigRepo}
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

// Chat handles free-form AI assistant questions, optionally with deployment context.
// POST /api/v1/ai/chat
func (h *AIHandler) Chat(c *gin.Context) {
	if !h.aiSvc.Available() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI integration not configured on this server (set AI_API_KEY)",
		})
		return
	}

	userID := middleware.GetUserID(c)
	var req struct {
		Message      string `json:"message" binding:"required"`
		DeploymentID string `json:"deployment_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	systemPrompt := `You are Pushpaka Assistant, an expert AI for the Pushpaka self-hosted cloud deployment platform.
You help users with:
- Debugging deployment failures and build errors
- Docker container issues and resource limits
- Git repository configuration and webhook setup
- Environment variable management
- CI/CD pipeline optimization
- Container networking and domain routing
- Performance monitoring and log analysis
Be concise, technical, and actionable. Format responses with markdown when helpful.`

	userMsg := req.Message

	// If a deployment ID is given, inject its logs as context
	if req.DeploymentID != "" {
		deployment, err := h.deployRepo.FindByID(req.DeploymentID)
		if err == nil && deployment.UserID == userID {
			logEntries, err := h.logRepo.FindByDeploymentID(req.DeploymentID)
			if err == nil && len(logEntries) > 0 {
				var sb strings.Builder
				// Include last 80 lines max for context
				start := 0
				if len(logEntries) > 80 {
					start = len(logEntries) - 80
				}
				for _, entry := range logEntries[start:] {
					sb.WriteString(entry.Message)
					sb.WriteByte('\n')
				}
				userMsg = "Deployment context (status: " + string(deployment.Status) + ", branch: " + deployment.Branch + "):\n```\n" + sb.String() + "```\n\nUser question: " + req.Message
			}
		}
	}

	reply, err := h.aiSvc.Ask(systemPrompt, userMsg)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI chat failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// GetAIConfig returns the AI provider config for the authenticated user.
// GET /api/v1/ai/config
func (h *AIHandler) GetAIConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cfg, err := h.aiConfigRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load AI config"})
		return
	}
	if cfg == nil {
		cfg = &models.AIConfig{UserID: userID, Provider: "openai"}
	}
	// Never expose the raw key — show only a masked version
	if cfg.APIKey != "" {
		visible := cfg.APIKey
		if len(visible) > 8 {
			visible = visible[:4] + strings.Repeat("*", len(visible)-8) + visible[len(visible)-4:]
		} else {
			visible = strings.Repeat("*", len(visible))
		}
		cfg.APIKeyMasked = visible
	}
	c.JSON(http.StatusOK, cfg)
}

// SaveAIConfig upserts the AI provider config for the authenticated user.
// PUT /api/v1/ai/config
func (h *AIHandler) SaveAIConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Provider           string `json:"provider"`
		APIKey             string `json:"api_key"`
		Model              string `json:"model"`
		BaseURL            string `json:"base_url"`
		SystemPrompt       string `json:"system_prompt"`
		MonitoringEnabled  bool   `json:"monitoring_enabled"`
		MonitoringInterval int    `json:"monitoring_interval"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := &models.AIConfig{
		UserID:             userID,
		Provider:           req.Provider,
		APIKey:             req.APIKey,
		Model:              req.Model,
		BaseURL:            req.BaseURL,
		SystemPrompt:       req.SystemPrompt,
		MonitoringEnabled:  req.MonitoringEnabled,
		MonitoringInterval: req.MonitoringInterval,
	}
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}
	if cfg.MonitoringInterval == 0 {
		cfg.MonitoringInterval = 300
	}

	if err := h.aiConfigRepo.Upsert(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save AI config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "AI settings saved"})
}
