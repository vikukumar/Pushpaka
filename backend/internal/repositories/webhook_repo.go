package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type WebhookRepository struct {
	db *sqlx.DB
}

func NewWebhookRepository(db *sqlx.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(w *models.WebhookConfig) error {
	query := `
		INSERT INTO webhook_configs (id, project_id, user_id, secret, provider, branch, created_at, updated_at)
		VALUES (:id, :project_id, :user_id, :secret, :provider, :branch, :created_at, :updated_at)`
	_, err := r.db.NamedExec(query, w)
	return err
}

func (r *WebhookRepository) FindByProjectID(projectID string) (*models.WebhookConfig, error) {
	var w models.WebhookConfig
	err := r.db.Get(&w,
		r.db.Rebind(`SELECT * FROM webhook_configs WHERE project_id = ? LIMIT 1`),
		projectID)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WebhookRepository) FindByID(id string) (*models.WebhookConfig, error) {
	var w models.WebhookConfig
	err := r.db.Get(&w,
		r.db.Rebind(`SELECT * FROM webhook_configs WHERE id = ?`),
		id)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WebhookRepository) ListByUserID(userID string) ([]models.WebhookConfig, error) {
	var wh []models.WebhookConfig
	err := r.db.Select(&wh,
		r.db.Rebind(`SELECT * FROM webhook_configs WHERE user_id = ? ORDER BY created_at DESC`),
		userID)
	return wh, err
}

func (r *WebhookRepository) Delete(id, userID string) error {
	_, err := r.db.Exec(
		r.db.Rebind(`DELETE FROM webhook_configs WHERE id = ? AND user_id = ?`),
		id, userID)
	return err
}

// PurgeExpiredOAuthStates removes expired OAuth CSRF state tokens.
func (r *WebhookRepository) PurgeExpiredOAuthStates(db *sqlx.DB) error {
	_, err := db.Exec(db.Rebind(`DELETE FROM oauth_states WHERE expires_at < ?`),
		models.NowUTC().String())
	return err
}
