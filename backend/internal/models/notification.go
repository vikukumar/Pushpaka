package models

// NotificationConfig holds per-user notification channel settings.
type NotificationConfig struct {
	ID                string `db:"id"                   json:"id"`
	UserID            string `db:"user_id"              json:"user_id"`
	SlackWebhookURL   string `db:"slack_webhook_url"    json:"slack_webhook_url"`
	DiscordWebhookURL string `db:"discord_webhook_url"  json:"discord_webhook_url"`
	SMTPHost          string `db:"smtp_host"            json:"smtp_host"`
	SMTPPort          int    `db:"smtp_port"            json:"smtp_port"`
	SMTPUsername      string `db:"smtp_username"        json:"smtp_username"`
	SMTPPassword      string `db:"smtp_password"        json:"-"`
	SMTPFrom          string `db:"smtp_from"            json:"smtp_from"`
	SMTPTo            string `db:"smtp_to"              json:"smtp_to"`
	NotifyOnSuccess   bool   `db:"notify_on_success"    json:"notify_on_success"`
	NotifyOnFailure   bool   `db:"notify_on_failure"    json:"notify_on_failure"`
	CreatedAt         Time   `db:"created_at"           json:"created_at"`
	UpdatedAt         Time   `db:"updated_at"           json:"updated_at"`
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
