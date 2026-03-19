package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type LogRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(l *models.DeploymentLog) error {
	return basemodel.Add(r.db, l)
}

func (r *LogRepository) FindByDeploymentID(deploymentID string) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	err := r.db.Where("deployment_id = ?", deploymentID).Order("created_at asc").Find(&logs).Error
	return logs, err
}

func (r *LogRepository) DeleteByDeploymentID(deploymentID string) error {
	return r.db.Where("deployment_id = ?", deploymentID).Delete(&models.DeploymentLog{}).Error
}
