package repositories

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/vikukumar/Pushpaka/pkg/models"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) FindByUserID(userID string) (*models.NotificationConfig, error) {
	var cfg models.NotificationConfig
	err := r.db.Where("user_id = ?", userID).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *NotificationRepository) Upsert(cfg *models.NotificationConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}}, // Unique index should exist on user_id
		DoUpdates: clause.AssignmentColumns([]string{
			"slack_webhook_url", "discord_webhook_url", "smtp_host", "smtp_port",
			"smtp_username", "smtp_password", "smtp_from", "smtp_to",
			"notify_on_success", "notify_on_failure", "updated_at",
		}),
	}).Create(cfg).Error
}
