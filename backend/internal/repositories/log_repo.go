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
	return basemodel.Query[models.DeploymentLog](r.db, "deployment_id = ?", deploymentID)
}

func (r *LogRepository) DeleteByDeploymentID(deploymentID string) error {
	basemodel.EnsureSynced[models.DeploymentLog](r.db)
	return r.db.Where("deployment_id = ?", deploymentID).Delete(&models.DeploymentLog{}).Error
}
