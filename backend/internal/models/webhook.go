package models

// WebhookConfig is a per-project incoming webhook endpoint.
// Pushpaka validates HMAC-SHA256 signatures from GitHub/GitLab and
// triggers a new deployment when the push matches the configured branch.
type WebhookConfig struct {
	ID        string `db:"id"         json:"id"`
	ProjectID string `db:"project_id" json:"project_id"`
	UserID    string `db:"user_id"    json:"user_id"`
	Secret    string `db:"secret"     json:"-"`        // never exposed in API responses
	Provider  string `db:"provider"   json:"provider"` // "github" | "gitlab"
	Branch    string `db:"branch"     json:"branch"`   // "" = any branch
	CreatedAt Time   `db:"created_at" json:"created_at"`
	UpdatedAt Time   `db:"updated_at" json:"updated_at"`
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
	WebhookURL string `json:"webhook_url"`
	CreatedAt  Time   `json:"created_at"`
}
