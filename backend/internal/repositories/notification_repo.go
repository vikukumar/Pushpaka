package repositories

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) FindByUserID(userID string) (*models.NotificationConfig, error) {
	return basemodel.First[models.NotificationConfig](r.db, "user_id = ?", userID)
}

func (r *NotificationRepository) Upsert(cfg *models.NotificationConfig) error {
	basemodel.EnsureSynced[models.NotificationConfig](r.db)
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}}, // Unique index should exist on user_id
		DoUpdates: clause.AssignmentColumns([]string{
			"slack_webhook_url", "discord_webhook_url", "smtp_host", "smtp_port",
			"smtp_username", "smtp_password", "smtp_from", "smtp_to",
			"notify_on_success", "notify_on_failure", "updated_at",
		}),
	}).Create(cfg).Error
}
