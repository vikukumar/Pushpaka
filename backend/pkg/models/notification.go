package models

import "github.com/vikukumar/Pushpaka/pkg/basemodel"

// NotificationConfig holds per-user notification channel settings.
type NotificationConfig struct {
	basemodel.BaseModel
	UserID            string `gorm:"uniqueIndex;type:varchar(255);not null" json:"user_id"`
	SlackWebhookURL   string `gorm:"type:text" json:"slack_webhook_url"`
	DiscordWebhookURL string `gorm:"type:text" json:"discord_webhook_url"`
	SMTPHost          string `gorm:"type:varchar(255)" json:"smtp_host"`
	SMTPPort          int    `gorm:"default:0" json:"smtp_port"`
	SMTPUsername      string `gorm:"type:varchar(255)" json:"smtp_username"`
	SMTPPassword      string `gorm:"type:text" json:"-"`
	SMTPFrom          string `gorm:"type:varchar(255)" json:"smtp_from"`
	SMTPTo            string `gorm:"type:varchar(255)" json:"smtp_to"`
	NotifyOnSuccess   bool   `gorm:"default:false" json:"notify_on_success"`
	NotifyOnFailure   bool   `gorm:"default:false" json:"notify_on_failure"`
}

type UpsertNotificationConfigRequest struct {
	SlackWebhookURL   string `json:"slack_webhook_url"`
	DiscordWebhookURL string `json:"discord_webhook_url"`
	SMTPHost          string `json:"smtp_host"`
	SMTPPort          int    `json:"smtp_port"`
	SMTPUsername      string `json:"smtp_username"`
	SMTPPassword      string `json:"smtp_password"`
	SMTPFrom          string `json:"smtp_from"`
	SMTPTo            string `json:"smtp_to"`
	NotifyOnSuccess   *bool  `json:"notify_on_success"`
	NotifyOnFailure   *bool  `json:"notify_on_failure"`
}

// NotificationEvent is the payload sent to all enabled channels after a
// deployment completes (success or failure).
type NotificationEvent struct {
	DeploymentID string `json:"deployment_id"`
	ProjectName  string `json:"project_name"`
	Status       string `json:"status"` // "running" or "failed"
	Branch       string `json:"branch"`
	CommitSHA    string `json:"commit_sha"`
	URL          string `json:"url"`
	ErrorMsg     string `json:"error_msg,omitempty"`
}
