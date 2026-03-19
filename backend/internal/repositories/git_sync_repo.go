package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type GitSyncRepository struct {
	db *gorm.DB
}

func NewGitSyncRepository(db *gorm.DB) *GitSyncRepository {
	return &GitSyncRepository{db: db}
}

// CreateGitSyncTrack creates a new git sync tracking record
func (r *GitSyncRepository) CreateGitSyncTrack(track *models.GitSyncTrack) error {
	return basemodel.Add(r.db, track)
}

// GetGitSyncTrackByDeploymentID retrieves sync track by deployment ID
func (r *GitSyncRepository) GetGitSyncTrackByDeploymentID(deploymentID string) (*models.GitSyncTrack, error) {
	return basemodel.First[models.GitSyncTrack](r.db, "deployment_id = ?", deploymentID)
}

// GetGitSyncTrackByID retrieves sync track by ID
func (r *GitSyncRepository) GetGitSyncTrackByID(id string) (*models.GitSyncTrack, error) {
	return basemodel.Get[models.GitSyncTrack](r.db, id)
}

// GetGitSyncTracksByProjectID retrieves all sync tracks for a project
func (r *GitSyncRepository) GetGitSyncTracksByProjectID(projectID string, limit, offset int) ([]models.GitSyncTrack, error) {
	var tracks []models.GitSyncTrack
	err := r.db.Where("project_id = ?", projectID).Order("updated_at desc").Limit(limit).Offset(offset).Find(&tracks).Error
	return tracks, err
}

// GetOutOfSyncDeployments retrieves all out-of-sync deployments
func (r *GitSyncRepository) GetOutOfSyncDeployments(limit int) ([]models.GitSyncTrack, error) {
	var tracks []models.GitSyncTrack
	err := r.db.Where("sync_status = ?", models.GitSyncOutOfSync).Order("updated_at asc").Limit(limit).Find(&tracks).Error
	return tracks, err
}

// GetPendingApprovalSyncs retrieves syncs pending approval
func (r *GitSyncRepository) GetPendingApprovalSyncs(projectID string) ([]models.GitSyncTrack, error) {
	var tracks []models.GitSyncTrack
	err := r.db.Where("project_id = ? AND sync_status = ?", projectID, models.GitSyncPending).Order("updated_at desc").Find(&tracks).Error
	return tracks, err
}

// UpdateGitSyncTrack updates an existing git sync track
func (r *GitSyncRepository) UpdateGitSyncTrack(track *models.GitSyncTrack) error {
	return basemodel.Modify(r.db, track)
}

// CreateGitChange creates a new git change record
func (r *GitSyncRepository) CreateGitChange(change *models.GitChange) error {
	return basemodel.Add(r.db, change)
}

// GetGitChangesBySyncTrackID retrieves all changes for a sync track
func (r *GitSyncRepository) GetGitChangesBySyncTrackID(syncTrackID string) ([]models.GitChange, error) {
	var changes []models.GitChange
	err := r.db.Where("sync_track_id = ?", syncTrackID).Order("created_at desc").Find(&changes).Error
	return changes, err
}

// CreateAutoSyncConfig creates auto-sync configuration
func (r *GitSyncRepository) CreateAutoSyncConfig(config *models.GitAutoSyncConfig) error {
	return basemodel.Add(r.db, config)
}

// GetAutoSyncConfig retrieves auto-sync config by deployment ID
func (r *GitSyncRepository) GetAutoSyncConfig(deploymentID string) (*models.GitAutoSyncConfig, error) {
	return basemodel.First[models.GitAutoSyncConfig](r.db, "deployment_id = ?", deploymentID)
}

// UpdateAutoSyncConfig updates auto-sync configuration
func (r *GitSyncRepository) UpdateAutoSyncConfig(config *models.GitAutoSyncConfig) error {
	return basemodel.Modify(r.db, config)
}

// CreateSyncHistory creates a sync history record
func (r *GitSyncRepository) CreateSyncHistory(history *models.DeploymentSyncHistory) error {
	return basemodel.Add(r.db, history)
}

// GetSyncHistory retrieves sync history for a deployment
func (r *GitSyncRepository) GetSyncHistory(deploymentID string, limit, offset int) ([]models.DeploymentSyncHistory, error) {
	var history []models.DeploymentSyncHistory
	err := r.db.Where("deployment_id = ?", deploymentID).Order("created_at desc").Limit(limit).Offset(offset).Find(&history).Error
	return history, err
}

// GetSyncHistoryCount returns the total number of sync history records
func (r *GitSyncRepository) GetSyncHistoryCount(deploymentID string) (int, error) {
	var count int64
	err := r.db.Model(&models.DeploymentSyncHistory{}).Where("deployment_id = ?", deploymentID).Count(&count).Error
	return int(count), err
}
