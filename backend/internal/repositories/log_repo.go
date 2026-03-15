package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type LogRepository struct {
	db *sqlx.DB
}

func NewLogRepository(db *sqlx.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(l *models.DeploymentLog) error {
	query := `
		INSERT INTO deployment_logs (id, deployment_id, level, message, stream, created_at)
		VALUES (:id, :deployment_id, :level, :message, :stream, :created_at)`
	_, err := r.db.NamedExec(query, l)
	return err
}

func (r *LogRepository) FindByDeploymentID(deploymentID string) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	err := r.db.Select(&logs,
		r.db.Rebind(`SELECT * FROM deployment_logs WHERE deployment_id = ? ORDER BY created_at ASC`),
		deploymentID)
	return logs, err
}

func (r *LogRepository) DeleteByDeploymentID(deploymentID string) error {
	_, err := r.db.Exec(r.db.Rebind(`DELETE FROM deployment_logs WHERE deployment_id = ?`), deploymentID)
	return err
}
