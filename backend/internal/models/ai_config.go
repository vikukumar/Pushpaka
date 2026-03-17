package models

// AIConfig stores per-user LLM integration settings.
// The APIKey field is never included in JSON responses.
type AIConfig struct {
	ID                 string `db:"id"                  json:"id"`
	UserID             string `db:"user_id"             json:"user_id"`
	Provider           string `db:"provider"            json:"provider"`
	APIKey             string `db:"api_key"             json:"-"`
	APIKeyMasked       string `db:"-"                   json:"api_key_masked"` // set by handler before response
	Model              string `db:"model"               json:"model"`
	BaseURL            string `db:"base_url"            json:"base_url"`
	SystemPrompt       string `db:"system_prompt"       json:"system_prompt"`
	MonitoringEnabled  bool   `db:"monitoring_enabled"  json:"monitoring_enabled"`
	MonitoringInterval int    `db:"monitoring_interval" json:"monitoring_interval"`
	CreatedAt          Time   `db:"created_at"          json:"created_at"`
	UpdatedAt          Time   `db:"updated_at"          json:"updated_at"`
}

// RAGDocument is a piece of custom knowledge injected into AI context.
type RAGDocument struct {
	ID        string `db:"id"         json:"id"`
	UserID    string `db:"user_id"    json:"user_id"`
	Title     string `db:"title"      json:"title"`
	Content   string `db:"content"    json:"content"`
	CreatedAt Time   `db:"created_at" json:"created_at"`
	UpdatedAt Time   `db:"updated_at" json:"updated_at"`
}

// AIMonitorAlert is a stored finding from the background AI monitoring service.
type AIMonitorAlert struct {
	ID           string `db:"id"            json:"id"`
	UserID       string `db:"user_id"       json:"user_id"`
	DeploymentID string `db:"deployment_id" json:"deployment_id"`
	Severity     string `db:"severity"      json:"severity"` // info | warn | error | critical
	Title        string `db:"title"         json:"title"`
	Message      string `db:"message"       json:"message"`
	Resolved     bool   `db:"resolved"      json:"resolved"`
	CreatedAt    Time   `db:"created_at"    json:"created_at"`
}

// K8sConfig stores Kubernetes cluster connection details for a user.
type K8sConfig struct {
	ID          string `db:"id"           json:"id"`
	UserID      string `db:"user_id"      json:"user_id"`
	Name        string `db:"name"         json:"name"`
	ServerURL   string `db:"server_url"   json:"server_url"`
	Token       string `db:"token"        json:"-"`
	TokenMasked string `db:"-"            json:"token_masked"` // set by handler
	Namespace   string `db:"namespace"    json:"namespace"`
	Kubeconfig  string `db:"kubeconfig"   json:"-"` // raw kubeconfig YAML
	RegistryURL string `db:"registry_url" json:"registry_url"`
	Enabled     bool   `db:"enabled"      json:"enabled"`
	CreatedAt   Time   `db:"created_at"   json:"created_at"`
	UpdatedAt   Time   `db:"updated_at"   json:"updated_at"`
}
