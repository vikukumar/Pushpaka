package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type GitSyncRepository struct {
	db *sqlx.DB
}

func NewGitSyncRepository(db *sqlx.DB) *GitSyncRepository {
	return &GitSyncRepository{db: db}
}

// CreateGitSyncTrack creates a new git sync tracking record
func (r *GitSyncRepository) CreateGitSyncTrack(track *models.GitSyncTrack) error {
	query := `
		INSERT INTO git_sync_tracks (
			id, deployment_id, project_id, repository, branch,
			current_commit_sha, latest_commit_sha, latest_commit_message, latest_commit_author,
			sync_status, sync_approval_required, total_changes, total_additions, total_deletions,
			changes_summary, created_at, updated_at
		) VALUES (
			:id, :deployment_id, :project_id, :repository, :branch,
			:current_commit_sha, :latest_commit_sha, :latest_commit_message, :latest_commit_author,
			:sync_status, :sync_approval_required, :total_changes, :total_additions, :total_deletions,
			:changes_summary, :created_at, :updated_at
		)`
	_, err := r.db.NamedExec(query, track)
	return err
}

// GetGitSyncTrackByDeploymentID retrieves sync track by deployment ID
func (r *GitSyncRepository) GetGitSyncTrackByDeploymentID(deploymentID string) (*models.GitSyncTrack, error) {
	var track models.GitSyncTrack
	err := r.db.Get(&track, r.db.Rebind(`SELECT * FROM git_sync_tracks WHERE deployment_id = ?`), deploymentID)
	if err != nil {
		return nil, err
	}
	return &track, nil
}

// GetGitSyncTrackByID retrieves sync track by ID
func (r *GitSyncRepository) GetGitSyncTrackByID(id string) (*models.GitSyncTrack, error) {
	var track models.GitSyncTrack
	err := r.db.Get(&track, r.db.Rebind(`SELECT * FROM git_sync_tracks WHERE id = ?`), id)
	if err != nil {
		return nil, err
	}
	return &track, nil
}

// GetGitSyncTracksByProjectID retrieves all sync tracks for a project
func (r *GitSyncRepository) GetGitSyncTracksByProjectID(projectID string, limit, offset int) ([]models.GitSyncTrack, error) {
	var tracks []models.GitSyncTrack
	err := r.db.Select(&tracks,
		r.db.Rebind(`SELECT * FROM git_sync_tracks WHERE project_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`),
		projectID, limit, offset)
	return tracks, err
}

// GetOutOfSyncDeployments retrieves all out-of-sync deployments
func (r *GitSyncRepository) GetOutOfSyncDeployments(limit int) ([]models.GitSyncTrack, error) {
	var tracks []models.GitSyncTrack
	err := r.db.Select(&tracks,
		r.db.Rebind(`SELECT * FROM git_sync_tracks WHERE sync_status = ? ORDER BY updated_at ASC LIMIT ?`),
		models.GitSyncOutOfSync, limit)
	return tracks, err
}

// GetPendingApprovalSyncs retrieves syncs pending approval
func (r *GitSyncRepository) GetPendingApprovalSyncs(projectID string) ([]models.GitSyncTrack, error) {
	var tracks []models.GitSyncTrack
	err := r.db.Select(&tracks,
		r.db.Rebind(`SELECT * FROM git_sync_tracks WHERE project_id = ? AND sync_status = ? ORDER BY updated_at DESC`),
		projectID, models.GitSyncPending)
	return tracks, err
}

// UpdateGitSyncTrack updates an existing git sync track
func (r *GitSyncRepository) UpdateGitSyncTrack(track *models.GitSyncTrack) error {
	query := `
		UPDATE git_sync_tracks SET
			latest_commit_sha = :latest_commit_sha,
			latest_commit_message = :latest_commit_message,
			latest_commit_author = :latest_commit_author,
			sync_status = :sync_status,
			sync_approval_required = :sync_approval_required,
			sync_approved_by = :sync_approved_by,
			sync_approved_at = :sync_approved_at,
			total_changes = :total_changes,
			total_additions = :total_additions,
			total_deletions = :total_deletions,
			changes_summary = :changes_summary,
			last_sync_attempt_at = :last_sync_attempt_at,
			last_sync_attempt_error = :last_sync_attempt_error,
			last_successful_sync_at = :last_successful_sync_at,
			notification_sent_at = :notification_sent_at,
			updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExec(query, track)
	return err
}

// CreateGitChange creates a new git change record
func (r *GitSyncRepository) CreateGitChange(change *models.GitChange) error {
	query := `
		INSERT INTO git_changes (id, sync_track_id, file_path, change_type, additions, deletions, old_content, new_content, created_at)
		VALUES (:id, :sync_track_id, :file_path, :change_type, :additions, :deletions, :old_content, :new_content, :created_at)`
	_, err := r.db.NamedExec(query, change)
	return err
}

// GetGitChangesBySyncTrackID retrieves all changes for a sync track
func (r *GitSyncRepository) GetGitChangesBySyncTrackID(syncTrackID string) ([]models.GitChange, error) {
	var changes []models.GitChange
	err := r.db.Select(&changes,
		r.db.Rebind(`SELECT * FROM git_changes WHERE sync_track_id = ? ORDER BY created_at DESC`),
		syncTrackID)
	return changes, err
}

// CreateAutoSyncConfig creates auto-sync configuration
func (r *GitSyncRepository) CreateAutoSyncConfig(config *models.GitAutoSyncConfig) error {
	query := `
		INSERT INTO git_auto_sync_config (
			id, project_id, deployment_id, enabled, require_approval,
			polling_interval, max_concurrent, only_prod_ready,
			allowed_branches, ignore_paths, required_approvers, created_at, updated_at
		) VALUES (
			:id, :project_id, :deployment_id, :enabled, :require_approval,
			:polling_interval, :max_concurrent, :only_prod_ready,
			:allowed_branches, :ignore_paths, :required_approvers, :created_at, :updated_at
		)`
	_, err := r.db.NamedExec(query, config)
	return err
}

// GetAutoSyncConfig retrieves auto-sync config by deployment ID
func (r *GitSyncRepository) GetAutoSyncConfig(deploymentID string) (*models.GitAutoSyncConfig, error) {
	var config models.GitAutoSyncConfig
	err := r.db.Get(&config,
		r.db.Rebind(`SELECT * FROM git_auto_sync_config WHERE deployment_id = ?`),
		deploymentID)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateAutoSyncConfig updates auto-sync configuration
func (r *GitSyncRepository) UpdateAutoSyncConfig(config *models.GitAutoSyncConfig) error {
	query := `
		UPDATE git_auto_sync_config SET
			enabled = :enabled,
			require_approval = :require_approval,
			polling_interval = :polling_interval,
			max_concurrent = :max_concurrent,
			only_prod_ready = :only_prod_ready,
			allowed_branches = :allowed_branches,
			ignore_paths = :ignore_paths,
			required_approvers = :required_approvers,
			updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExec(query, config)
	return err
}

// CreateSyncHistory creates a sync history record
func (r *GitSyncRepository) CreateSyncHistory(history *models.DeploymentSyncHistory) error {
	query := `
		INSERT INTO deployment_sync_history (
			id, deployment_id, project_id, from_commit_sha, to_commit_sha,
			sync_type, status, total_changes, duration, triggered_by,
			sync_error, rollback_triggered, rollback_reason, created_at, updated_at
		) VALUES (
			:id, :deployment_id, :project_id, :from_commit_sha, :to_commit_sha,
			:sync_type, :status, :total_changes, :duration, :triggered_by,
			:sync_error, :rollback_triggered, :rollback_reason, :created_at, :updated_at
		)`
	_, err := r.db.NamedExec(query, history)
	return err
}

// GetSyncHistory retrieves sync history for a deployment
func (r *GitSyncRepository) GetSyncHistory(deploymentID string, limit, offset int) ([]models.DeploymentSyncHistory, error) {
	var history []models.DeploymentSyncHistory
	err := r.db.Select(&history,
		r.db.Rebind(`SELECT * FROM deployment_sync_history WHERE deployment_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`),
		deploymentID, limit, offset)
	return history, err
}

// GetSyncHistoryCount returns the total number of sync history records
func (r *GitSyncRepository) GetSyncHistoryCount(deploymentID string) (int, error) {
	var count int
	err := r.db.Get(&count,
		r.db.Rebind(`SELECT COUNT(*) FROM deployment_sync_history WHERE deployment_id = ?`),
		deploymentID)
	return count, err
}
