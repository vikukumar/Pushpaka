package models

import (
	"time"

	"github.com/vikukumar/pushpaka/pkg/basemodel"
)

// WebhookConfig is a per-project incoming webhook endpoint.
// Pushpaka validates HMAC-SHA256 signatures from GitHub/GitLab and
// triggers a new deployment when the push matches the configured branch.
type WebhookConfig struct {
	basemodel.BaseModel
	ProjectID string `gorm:"index;type:varchar(255);not null" json:"project_id"`
	UserID    string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Secret    string `gorm:"type:text;not null" json:"-"`      // never exposed in API responses
	Provider  string `gorm:"type:varchar(50)" json:"provider"` // "github" | "gitlab"
	Branch    string `gorm:"type:varchar(100)" json:"branch"`  // "" = any branch
}

type CreateWebhookRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Provider  string `json:"provider"` // default "github"
	Branch    string `json:"branch"`
}

type WebhookConfigResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Provider  string `json:"provider"`
	Branch    string `json:"branch"`
	// WebhookURL is the full URL the user must register on GitHub/GitLab.
	WebhookURL string    `json:"webhook_url"`
	CreatedAt  time.Time `json:"created_at"`
}
