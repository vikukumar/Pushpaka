package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(p *models.Project) error {
	return basemodel.Add(r.db, p)
}

func (r *ProjectRepository) FindByUserID(userID string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) FindByID(id, userID string) (*models.Project, error) {
	return basemodel.First[models.Project](r.db, "id = ? AND user_id = ?", id, userID)
}

func (r *ProjectRepository) Update(p *models.Project) error {
	return basemodel.Modify(r.db, p)
}

// FindByIDInternal returns the project including the git_token (for worker jobs).
// Only use this server-side -- never marshal the result directly to an API response.
func (r *ProjectRepository) FindByIDInternal(id string) (*models.Project, error) {
	return basemodel.Get[models.Project](r.db, id)
}

func (r *ProjectRepository) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Project{}).Error
}
