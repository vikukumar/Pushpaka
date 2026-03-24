package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type CommitRepository struct {
	db *gorm.DB
}

func NewCommitRepository(db *gorm.DB) *CommitRepository {
	return &CommitRepository{db: db}
}

func (r *CommitRepository) Create(commit *models.ProjectCommit) error {
	return basemodel.Add(r.db, commit)
}

func (r *CommitRepository) Update(commit *models.ProjectCommit) error {
	return basemodel.Modify(r.db, commit)
}

func (r *CommitRepository) Get(id string) (*models.ProjectCommit, error) {
	return basemodel.Get[models.ProjectCommit](r.db, id)
}

func (r *CommitRepository) FindBySHA(projectID, sha string) (*models.ProjectCommit, error) {
	return basemodel.First[models.ProjectCommit](r.db, "project_id = ? AND sha = ?", projectID, sha)
}

func (r *CommitRepository) FindByProject(projectID string, limit int) ([]models.ProjectCommit, error) {
	basemodel.EnsureSynced[models.ProjectCommit](r.db)
	var dest []models.ProjectCommit
	err := r.db.Where("project_id = ?", projectID).Order("committed_at DESC").Limit(limit).Find(&dest).Error
	return dest, err
}

func (r *CommitRepository) DeleteOldCommits(projectID string, keep int) error {
	basemodel.EnsureSynced[models.ProjectCommit](r.db)
	var ids []string
	// Get IDs of commits to delete (those beyond the 'keep' limit)
	err := r.db.Model(&models.ProjectCommit{}).
		Where("project_id = ?", projectID).
		Order("committed_at desc").
		Offset(keep).
		Pluck("id", &ids).Error
	if err != nil {
		return err
	}

	if len(ids) > 0 {
		return r.db.Where("id IN ?", ids).Delete(&models.ProjectCommit{}).Error
	}
	return nil
}

func (r *CommitRepository) FindOldCommits(projectID string, keep int) ([]models.ProjectCommit, error) {
	basemodel.EnsureSynced[models.ProjectCommit](r.db)
	var dest []models.ProjectCommit
	err := r.db.Where("project_id = ?", projectID).Order("committed_at DESC").Offset(keep).Find(&dest).Error
	return dest, err
}

func (r *CommitRepository) FindPendingTests(limit int) ([]models.ProjectCommit, error) {
	basemodel.EnsureSynced[models.ProjectCommit](r.db)
	var dest []models.ProjectCommit
	err := r.db.Where("status = ?", models.CommitStatusBuilt).Order("created_at ASC").Limit(limit).Find(&dest).Error
	return dest, err
}
