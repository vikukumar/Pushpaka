package repositories

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type DeploymentRepository struct {
	db *sqlx.DB
}

func NewDeploymentRepository(db *sqlx.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) Create(d *models.Deployment) error {
	query := `
		INSERT INTO deployments (id, project_id, user_id, commit_sha, commit_msg, branch, status, image_tag, container_id, url, error_msg, started_at, finished_at, created_at, updated_at)
		VALUES (:id, :project_id, :user_id, :commit_sha, :commit_msg, :branch, :status, :image_tag, :container_id, :url, :error_msg, :started_at, :finished_at, :created_at, :updated_at)`
	_, err := r.db.NamedExec(query, d)
	return err
}

func (r *DeploymentRepository) FindByProjectID(projectID string, limit, offset int) ([]models.Deployment, error) {
	var deployments []models.Deployment
	err := r.db.Select(&deployments,
		r.db.Rebind(`SELECT * FROM deployments WHERE project_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`),
		projectID, limit, offset)
	return deployments, err
}

func (r *DeploymentRepository) FindByUserID(userID string, limit, offset int) ([]models.Deployment, error) {
	var deployments []models.Deployment
	err := r.db.Select(&deployments,
		r.db.Rebind(`SELECT * FROM deployments WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`),
		userID, limit, offset)
	return deployments, err
}

func (r *DeploymentRepository) FindByID(id string) (*models.Deployment, error) {
	var d models.Deployment
	err := r.db.Get(&d, r.db.Rebind(`SELECT * FROM deployments WHERE id = ?`), id)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DeploymentRepository) UpdateStatus(id string, status models.DeploymentStatus, errorMsg string) error {
	_, err := r.db.Exec(
		r.db.Rebind(`UPDATE deployments SET status = ?, error_msg = ?, updated_at = ? WHERE id = ?`),
		status, errorMsg, time.Now().UTC(), id)
	return err
}

func (r *DeploymentRepository) Update(d *models.Deployment) error {
	query := `
		UPDATE deployments
		SET status = :status, image_tag = :image_tag, container_id = :container_id,
		    url = :url, error_msg = :error_msg, started_at = :started_at,
		    finished_at = :finished_at, updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExec(query, d)
	return err
}

func (r *DeploymentRepository) FindLatestByProjectID(projectID string) (*models.Deployment, error) {
	var d models.Deployment
	err := r.db.Get(&d,
		r.db.Rebind(`SELECT * FROM deployments WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`),
		projectID)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
