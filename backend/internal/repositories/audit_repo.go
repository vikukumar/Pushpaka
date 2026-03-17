package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, resource, resource_id, metadata, ip_addr, user_agent, created_at)
		VALUES (:id, :user_id, :action, :resource, :resource_id, :metadata, :ip_addr, :user_agent, :created_at)`
	_, err := r.db.NamedExec(query, log)
	return err
}

func (r *AuditRepository) FindByUserID(userID string, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Select(&logs,
		r.db.Rebind(`SELECT * FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`),
		userID, limit, offset)
	return logs, err
}

func (r *AuditRepository) FindAll(limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Select(&logs,
		r.db.Rebind(`SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`),
		limit, offset)
	return logs, err
}
