package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/internal/services"
)

// GitSyncWorker handles background git synchronization tasks
type GitSyncWorker struct {
	gitSyncService  *services.GitSyncService
	gitSyncRepo     *repositories.GitSyncRepository
	deploymentRepo  *repositories.DeploymentRepository
	projectRepo     *repositories.ProjectRepository
	notificationSvc *services.NotificationService
	rdb             *redis.Client
	name            string
	pollInterval    time.Duration
	logger          *zerolog.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewGitSyncWorker creates a new git sync worker
func NewGitSyncWorker(
	gitSyncService *services.GitSyncService,
	gitSyncRepo *repositories.GitSyncRepository,
	deploymentRepo *repositories.DeploymentRepository,
	projectRepo *repositories.ProjectRepository,
	notificationSvc *services.NotificationService,
	rdb *redis.Client,
	logger *zerolog.Logger,
) *GitSyncWorker {
	return &GitSyncWorker{
		gitSyncService:  gitSyncService,
		gitSyncRepo:     gitSyncRepo,
		deploymentRepo:  deploymentRepo,
		projectRepo:     projectRepo,
		notificationSvc: notificationSvc,
		rdb:             rdb,
		name:            "git-sync-worker",
		pollInterval:    10 * time.Second, // Check every 10 seconds
		logger:          logger,
	}
}

// Start begins the git sync worker
func (w *GitSyncWorker) Start(ctx context.Context) error {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.logger.Info().Msg("git sync worker started")

	go w.runPolling()
	go w.runApprovalCleanup()

	return nil
}

// Stop stops the git sync worker
func (w *GitSyncWorker) Stop() error {
	w.logger.Info().Msg("stopping git sync worker")
	if w.cancel != nil {
		w.cancel()
	}
	return nil
}

// runPolling runs the main polling loop for auto-sync
func (w *GitSyncWorker) runPolling() {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info().Msg("polling loop stopped")
			return

		case <-ticker.C:
			w.checkForUpdates()
			w.processAutoSync()
		}
	}
}

// runApprovalCleanup runs cleanup of stale approval requests
func (w *GitSyncWorker) runApprovalCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info().Msg("approval cleanup stopped")
			return

		case <-ticker.C:
			w.cleanupStalePendingApprovals()
		}
	}
}

// checkForUpdates checks all tracked deployments for git updates
func (w *GitSyncWorker) checkForUpdates() {
	// Get all tracked sync records
	tracks, err := w.gitSyncRepo.GetOutOfSyncDeployments(100)
	if err != nil {
		w.logger.Error().Err(err).Msg("failed to get out-of-sync deployments")
		return
	}

	for _, track := range tracks {
		if err := w.gitSyncService.CheckForUpdates(&track); err != nil {
			w.logger.Error().Err(err).
				Str("deployment_id", track.DeploymentID).
				Msg("failed to check for updates")
			continue
		}

		// Send notifications if status changed
		if track.SyncStatus == "out_of_sync" && track.NotificationSentAt == nil {
			w.notifyOutOfSync(&track)
		}
	}
}

// processAutoSync processes auto-sync for eligible deployments
func (w *GitSyncWorker) processAutoSync() {
	// Get all auto-sync configs
	configs, err := w.gitSyncRepo.GetAllAutoSyncConfigs()
	if err != nil {
		w.logger.Error().Err(err).Msg("failed to get auto-sync configs")
		return
	}

	for _, config := range configs {
		if !config.Enabled {
			continue
		}

		// Check if polling interval has elapsed
		if !w.shouldPolling(&config) {
			continue
		}

		// Get sync track
		track, err := w.gitSyncRepo.GetGitSyncTrackByDeploymentID(config.DeploymentID)
		if err != nil {
			w.logger.Error().Err(err).
				Str("deployment_id", config.DeploymentID).
				Msg("failed to get sync track")
			continue
		}

		// Check if should auto-sync
		if !w.gitSyncService.ShouldAutoSync(track, &config) {
			continue
		}

		// Check if requires approval
		if config.RequireApproval && track.SyncApprovedBy == "" {
			w.logger.Info().
				Str("deployment_id", config.DeploymentID).
				Msg("auto-sync requires approval, waiting...")
			w.notifyPendingApproval(track)
			continue
		}

		// Perform sync
		userID := "auto-sync-worker"
		if err := w.gitSyncService.SyncDeployment(track.ID, userID, false); err != nil {
			w.logger.Error().Err(err).
				Str("deployment_id", config.DeploymentID).
				Msg("failed to auto-sync deployment")
			w.notifyAutoSyncFailed(track, err)
			continue
		}

		w.logger.Info().
			Str("deployment_id", config.DeploymentID).
			Msg("auto-sync completed successfully")
		w.notifyAutoSyncSuccess(track)
	}
}

// shouldPolling checks if polling interval has elapsed for a config
func (w *GitSyncWorker) shouldPolling(config *repositories.GitAutoSyncConfig) bool {
	// Use redis for distributed lock/timing
	key := fmt.Sprintf("pushpaka:autosync:poll:%s", config.ID)

	// Check if we've already polled recently
	result, err := w.rdb.Get(w.ctx, key).Result()
	if err == nil && result == "1" {
		return false // Already polled recently
	}

	// Set lock with TTL equal to polling interval
	w.rdb.Set(w.ctx, key, "1", time.Duration(config.PollingInterval)*time.Second)
	return true
}

// cleanupStalePendingApprovals removes approval requests that have timed out
func (w *GitSyncWorker) cleanupStalePendingApprovals() {
	// Implementation would depend on your database schema
	// This example assumes a method exists on the repo
	cleaned, err := w.gitSyncRepo.CleanupStalePendingApprovals(24 * time.Hour)
	if err != nil {
		w.logger.Error().Err(err).Msg("failed to cleanup stale approvals")
		return
	}

	if cleaned > 0 {
		w.logger.Info().Int("count", cleaned).Msg("cleaned up stale approval requests")
	}
}

// notifyOutOfSync sends notification when deployment is out of sync
func (w *GitSyncWorker) notifyOutOfSync(track *repositories.GitSyncTrack) {
	deployment, err := w.deploymentRepo.FindByID(track.DeploymentID)
	if err != nil {
		w.logger.Error().Err(err).Msg("failed to get deployment for notification")
		return
	}

	project, err := w.projectRepo.FindByID(track.ProjectID)
	if err != nil {
		w.logger.Error().Err(err).Msg("failed to get project for notification")
		return
	}

	message := fmt.Sprintf(
		"Deployment '%s' is out of sync.\nCurrent: %s\nLatest: %s\nChanges: %d files",
		deployment.ID,
		track.CurrentCommitSHA[:8],
		track.LatestCommitSHA[:8],
		track.TotalChanges,
	)

	notification := &repositories.Notification{
		ProjectID: project.ID,
		Type:      "git_out_of_sync",
		Title:     fmt.Sprintf("Out of Sync: %s", deployment.ID),
		Message:   message,
		Severity:  "warning",
		CreatedAt: time.Now(),
	}

	w.notificationSvc.SendNotification(notification)

	// Mark notification as sent
	track.NotificationSentAt = &repositories.Time{Time: time.Now()}
	if err := w.gitSyncRepo.UpdateGitSyncTrack(track); err != nil {
		w.logger.Error().Err(err).Msg("failed to update notification timestamp")
	}
}

// notifyPendingApproval sends notification for pending approval
func (w *GitSyncWorker) notifyPendingApproval(track *repositories.GitSyncTrack) {
	if track.NotificationSentAt != nil {
		// Already notified, don't spam
		return
	}

	message := fmt.Sprintf(
		"Deployment '%s' has pending auto-sync approval.\nChanges: %d files\nApproval required from: %s",
		track.DeploymentID,
		track.TotalChanges,
		"tech-leads", // This should come from config
	)

	notification := &repositories.Notification{
		ProjectID: track.ProjectID,
		Type:      "git_sync_pending_approval",
		Title:     fmt.Sprintf("Approval Pending: %s", track.DeploymentID),
		Message:   message,
		Severity:  "info",
		CreatedAt: time.Now(),
	}

	w.notificationSvc.SendNotification(notification)
}

// notifyAutoSyncSuccess sends notification on successful auto-sync
func (w *GitSyncWorker) notifyAutoSyncSuccess(track *repositories.GitSyncTrack) {
	message := fmt.Sprintf(
		"Deployment '%s' auto-synced successfully.\nNewCommit: %s\nChanges: %d files",
		track.DeploymentID,
		track.LatestCommitSHA[:8],
		track.TotalChanges,
	)

	notification := &repositories.Notification{
		ProjectID: track.ProjectID,
		Type:      "git_sync_success",
		Title:     fmt.Sprintf("Auto-Sync Successful: %s", track.DeploymentID),
		Message:   message,
		Severity:  "info",
		CreatedAt: time.Now(),
	}

	w.notificationSvc.SendNotification(notification)
}

// notifyAutoSyncFailed sends notification on auto-sync failure
func (w *GitSyncWorker) notifyAutoSyncFailed(track *repositories.GitSyncTrack, err error) {
	message := fmt.Sprintf(
		"Deployment '%s' auto-sync failed.\nError: %v\nPlease investigate and retry manually.",
		track.DeploymentID,
		err,
	)

	notification := &repositories.Notification{
		ProjectID: track.ProjectID,
		Type:      "git_sync_failed",
		Title:     fmt.Sprintf("Auto-Sync Failed: %s", track.DeploymentID),
		Message:   message,
		Severity:  "error",
		CreatedAt: time.Now(),
	}

	w.notificationSvc.SendNotification(notification)
}

// HealthCheck returns worker health status
func (w *GitSyncWorker) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"name":    w.name,
		"status":  "healthy",
		"running": w.ctx != nil && w.ctx.Err() == nil,
	}
}
