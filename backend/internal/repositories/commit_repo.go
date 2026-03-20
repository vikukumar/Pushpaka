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
	var commit models.ProjectCommit
	err := r.db.Where("project_id = ? AND sha = ?", projectID, sha).First(&commit).Error
	if err != nil {
		return nil, err
	}
	return &commit, nil
}

func (r *CommitRepository) FindByProject(projectID string, limit int) ([]models.ProjectCommit, error) {
	var commits []models.ProjectCommit
	err := r.db.Where("project_id = ?", projectID).Order("committed_at desc").Limit(limit).Find(&commits).Error
	return commits, err
}

func (r *CommitRepository) DeleteOldCommits(projectID string, keep int) error {
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
	var commits []models.ProjectCommit
	err := r.db.Where("project_id = ?", projectID).
		Order("committed_at desc").
		Offset(keep).
		Find(&commits).Error
	return commits, err
}

func (r *CommitRepository) FindPendingTests(limit int) ([]models.ProjectCommit, error) {
	var commits []models.ProjectCommit
	err := r.db.Where("status = ?", models.CommitStatusBuilt).Order("created_at asc").Limit(limit).Find(&commits).Error
	return commits, err
}
