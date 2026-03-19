package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type DeploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) Create(d *models.Deployment) error {
	return basemodel.Add(r.db, d)
}

func (r *DeploymentRepository) FindByProjectID(projectID string, limit, offset int) ([]models.Deployment, error) {
	var deployments []models.Deployment
	err := r.db.Where("project_id = ?", projectID).Order("created_at desc").Limit(limit).Offset(offset).Find(&deployments).Error
	return deployments, err
}

func (r *DeploymentRepository) FindByUserID(userID string, limit, offset int) ([]models.Deployment, error) {
	var deployments []models.Deployment
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Offset(offset).Find(&deployments).Error
	return deployments, err
}

func (r *DeploymentRepository) FindByID(id string) (*models.Deployment, error) {
	return basemodel.Get[models.Deployment](r.db, id)
}

func (r *DeploymentRepository) UpdateStatus(id string, status models.DeploymentStatus, errorMsg string) error {
	return r.db.Model(&models.Deployment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"error_msg":  errorMsg,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *DeploymentRepository) Update(d *models.Deployment) error {
	return basemodel.Modify(r.db, d)
}

func (r *DeploymentRepository) FindRunningByProjectID(projectID string) (*models.Deployment, error) {
	var d models.Deployment
	err := r.db.Where("project_id = ? AND status = ?", projectID, "running").Order("created_at desc").First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DeploymentRepository) FindLatestByProjectID(projectID string) (*models.Deployment, error) {
	var d models.Deployment
	err := r.db.Where("project_id = ?", projectID).Order("created_at desc").First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// FailStaleQueued marks all deployments that are still "queued" as "failed".
// Called at startup when no Redis/worker is available, to clear stuck records.
func (r *DeploymentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Deployment{}).Error
}

func (r *DeploymentRepository) FailStaleQueued(errorMsg string) error {
	// Use models.NowUTC() so Value() emits RFC3339Nano text, which scans correctly
	// back into *models.Time via Scan() without format ambiguity.
	now := models.NowUTC()
	nowStr, _ := now.Value()
	return r.db.Model(&models.Deployment{}).Where("status = ?", "queued").Updates(map[string]interface{}{
		"status":      "failed",
		"error_msg":   errorMsg,
		"finished_at": nowStr,
		"updated_at":  nowStr,
	}).Error
}
