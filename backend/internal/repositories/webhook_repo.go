package repositories

import (
	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
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
	var w models.WebhookConfig
	err := r.db.Where("project_id = ?", projectID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WebhookRepository) FindByID(id string) (*models.WebhookConfig, error) {
	return basemodel.Get[models.WebhookConfig](r.db, id)
}

func (r *WebhookRepository) ListByUserID(userID string) ([]models.WebhookConfig, error) {
	var wh []models.WebhookConfig
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&wh).Error
	return wh, err
}

func (r *WebhookRepository) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.WebhookConfig{}).Error
}

// PurgeExpiredOAuthStates removes expired OAuth CSRF state tokens.
func (r *WebhookRepository) PurgeExpiredOAuthStates(db *gorm.DB) error {
	return db.Exec("DELETE FROM oauth_states WHERE expires_at < ?", models.NowUTC().String()).Error
}
