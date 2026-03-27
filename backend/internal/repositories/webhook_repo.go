package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/pushpaka/pkg/basemodel"
	"github.com/vikukumar/pushpaka/pkg/models"
)

type WebhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(w *models.WebhookConfig) error {
	return basemodel.Add(r.db, w)
}

func (r *WebhookRepository) FindByProjectID(projectID string) (*models.WebhookConfig, error) {
	return basemodel.First[models.WebhookConfig](r.db, "project_id = ?", projectID)
}

func (r *WebhookRepository) FindByID(id string) (*models.WebhookConfig, error) {
	return basemodel.Get[models.WebhookConfig](r.db, id)
}

func (r *WebhookRepository) ListByUserID(userID string) ([]models.WebhookConfig, error) {
	return basemodel.Query[models.WebhookConfig](r.db, "user_id = ?", userID)
}

func (r *WebhookRepository) Delete(id, userID string) error {
	basemodel.EnsureSynced[models.WebhookConfig](r.db)
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.WebhookConfig{}).Error
}

// PurgeExpiredOAuthStates removes expired OAuth CSRF state tokens.
func (r *WebhookRepository) PurgeExpiredOAuthStates(db *gorm.DB) error {
	// Note: oauth_states is likely a raw table not in models.
	// If it is a model, we should call EnsureSynced.
	return db.Exec("DELETE FROM oauth_states WHERE expires_at < ?", models.NowUTC().String()).Error
}
