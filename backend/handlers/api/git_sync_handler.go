package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vikukumar/Pushpaka/internal/services"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type GitSyncHandler struct {
	gitSyncService *services.GitSyncService
}

func NewGitSyncHandler(gitSyncService *services.GitSyncService) *GitSyncHandler {
	return &GitSyncHandler{
		gitSyncService: gitSyncService,
	}
}

// CheckForUpdates checks for git updates without syncing
// GET /api/v1/deployments/:deploymentId/git/check-updates
func (h *GitSyncHandler) CheckForUpdates(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	// Get sync track (implementation depends on your route setup)
	// This would need to be retrieved from your service
	track := &models.GitSyncTrack{
		DeploymentID: deploymentID,
	}

	if err := h.gitSyncService.CheckForUpdates(track); err != nil {
		log.Error().Err(err).Str("deployment_id", deploymentID).Msg("failed to check for updates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sync_status":   track.SyncStatus,
		"latest_sha":    track.LatestCommitSHA,
		"total_changes": track.TotalChanges,
	})
}

// SyncDeployment syncs deployment to latest git code
// POST /api/v1/deployments/:deploymentId/git/sync
func (h *GitSyncHandler) SyncDeployment(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	var req models.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.gitSyncService.SyncDeployment(deploymentID, userID, req.Force); err != nil {
		log.Error().Err(err).Str("deployment_id", deploymentID).Msg("failed to sync deployment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "syncing",
		"deployment_id": deploymentID,
	})
}

// GetSyncStatus gets the current sync status
// GET /api/v1/deployments/:deploymentId/git/status
func (h *GitSyncHandler) GetSyncStatus(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	// This would need proper service integration
	c.JSON(http.StatusOK, gin.H{
		"deployment_id": deploymentID,
		"sync_status":   "synced",
	})
}

// GetSyncHistory gets the sync history
// GET /api/v1/deployments/:deploymentId/git/history
func (h *GitSyncHandler) GetSyncHistory(c *gin.Context) {
	deploymentID := c.Param("deploymentId")
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")

	c.JSON(http.StatusOK, gin.H{
		"deployment_id": deploymentID,
		"limit":         limit,
		"offset":        offset,
		"history":       []interface{}{},
	})
}

// ApproveSync approves a pending sync request
// POST /api/v1/deployments/:deploymentId/git/approve-sync
func (h *GitSyncHandler) ApproveSync(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	var req models.SyncApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.gitSyncService.ApproveSyncRequest(req.SyncTrackID, userID, req.Approved); err != nil {
		log.Error().Err(err).Str("deployment_id", deploymentID).Msg("failed to approve sync")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status := "approved"
	if !req.Approved {
		status = "rejected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        status,
		"deployment_id": deploymentID,
	})
}

// GetChanges gets the detailed changes between commits
// GET /api/v1/deployments/:deploymentId/git/changes
func (h *GitSyncHandler) GetChanges(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	c.JSON(http.StatusOK, gin.H{
		"deployment_id": deploymentID,
		"changes":       []interface{}{},
	})
}

// EnableAutoSync enables automatic synchronization
// POST /api/v1/deployments/:deploymentId/git/auto-sync
func (h *GitSyncHandler) EnableAutoSync(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	var config models.GitAutoSyncConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.gitSyncService.EnableAutoSync(deploymentID, &config); err != nil {
		log.Error().Err(err).Str("deployment_id", deploymentID).Msg("failed to enable auto-sync")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "enabled",
		"deployment_id": deploymentID,
	})
}

// GetAutoSyncConfig gets auto-sync configuration
// GET /api/v1/deployments/:deploymentId/git/auto-sync
func (h *GitSyncHandler) GetAutoSyncConfig(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	config, err := h.gitSyncService.GetAutoSyncConfig(deploymentID)
	if err != nil {
		log.Error().Err(err).Str("deployment_id", deploymentID).Msg("failed to get auto-sync config")
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateAutoSyncConfig updates auto-sync configuration
// PATCH /api/v1/deployments/:deploymentId/git/auto-sync
func (h *GitSyncHandler) UpdateAutoSyncConfig(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	var config models.GitAutoSyncConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	config.DeploymentID = deploymentID
	if err := h.gitSyncService.UpdateAutoSyncConfig(&config); err != nil {
		log.Error().Err(err).Str("deployment_id", deploymentID).Msg("failed to update auto-sync config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DisableAutoSync disables automatic synchronization
// DELETE /api/v1/deployments/:deploymentId/git/auto-sync
func (h *GitSyncHandler) DisableAutoSync(c *gin.Context) {
	deploymentID := c.Param("deploymentId")

	config, err := h.gitSyncService.GetAutoSyncConfig(deploymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	config.Enabled = false
	if err := h.gitSyncService.UpdateAutoSyncConfig(config); err != nil {
		log.Error().Err(err).Str("deployment_id", deploymentID).Msg("failed to disable auto-sync")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "disabled",
		"deployment_id": deploymentID,
	})
}
