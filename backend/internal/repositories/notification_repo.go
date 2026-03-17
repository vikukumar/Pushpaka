package repositories

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) FindByUserID(userID string) (*models.NotificationConfig, error) {
	var cfg models.NotificationConfig
	err := r.db.Get(&cfg,
		r.db.Rebind(`SELECT * FROM notification_configs WHERE user_id = ?`),
		userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *NotificationRepository) Upsert(cfg *models.NotificationConfig) error {
	query := r.db.Rebind(`
		INSERT INTO notification_configs
			(id, user_id, slack_webhook_url, discord_webhook_url, smtp_host, smtp_port,
			 smtp_username, smtp_password, smtp_from, smtp_to, notify_on_success, notify_on_failure,
			 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			slack_webhook_url   = excluded.slack_webhook_url,
			discord_webhook_url = excluded.discord_webhook_url,
			smtp_host           = excluded.smtp_host,
			smtp_port           = excluded.smtp_port,
			smtp_username       = excluded.smtp_username,
			smtp_password       = excluded.smtp_password,
			smtp_from           = excluded.smtp_from,
			smtp_to             = excluded.smtp_to,
			notify_on_success   = excluded.notify_on_success,
			notify_on_failure   = excluded.notify_on_failure,
			updated_at          = excluded.updated_at`)
	_, err := r.db.Exec(query,
		cfg.ID, cfg.UserID, cfg.SlackWebhookURL, cfg.DiscordWebhookURL,
		cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword,
		cfg.SMTPFrom, cfg.SMTPTo, cfg.NotifyOnSuccess, cfg.NotifyOnFailure,
		cfg.CreatedAt, cfg.UpdatedAt)
	return err
}
