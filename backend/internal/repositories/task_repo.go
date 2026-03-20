package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *models.ProjectTask) error {
	return basemodel.Add(r.db, task)
}

func (r *TaskRepository) Update(task *models.ProjectTask) error {
	return basemodel.Modify(r.db, task)
}

func (r *TaskRepository) Get(id string) (*models.ProjectTask, error) {
	return basemodel.Get[models.ProjectTask](r.db, id)
}

func (r *TaskRepository) FindByProject(projectID string, limit int) ([]models.ProjectTask, error) {
	var tasks []models.ProjectTask
	err := r.db.Where("project_id = ?", projectID).Order("created_at desc").Limit(limit).Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) FindByProjectID(projectID string) ([]models.ProjectTask, error) {
	return r.FindByProject(projectID, 50)
}

func (r *TaskRepository) FindPending(taskType models.TaskType, limit int) ([]models.ProjectTask, error) {
	var tasks []models.ProjectTask
	err := r.db.Where("type = ? AND status = ?", taskType, models.TaskStatusPending).
		Order("created_at asc").Limit(limit).Find(&tasks).Error
	return tasks, err
}
