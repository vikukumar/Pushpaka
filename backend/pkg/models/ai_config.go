package models

import "github.com/vikukumar/pushpaka/pkg/basemodel"

// AIConfig stores per-user LLM integration settings.
// The APIKey field is never included in JSON responses.
type AIConfig struct {
	basemodel.BaseModel
	UserID             string `gorm:"uniqueIndex;type:varchar(255);not null" json:"user_id"`
	Provider           string `gorm:"type:varchar(50)" json:"provider"`
	APIKey             string `gorm:"type:text" json:"-"`
	APIKeyMasked       string `gorm:"-" json:"api_key_masked"` // set by handler before response
	Model              string `gorm:"type:varchar(100)" json:"model"`
	BaseURL            string `gorm:"type:varchar(255)" json:"base_url"`
	SystemPrompt       string `gorm:"type:text" json:"system_prompt"`
	MonitoringEnabled  bool   `gorm:"default:false" json:"monitoring_enabled"`
	MonitoringInterval int    `gorm:"default:0" json:"monitoring_interval"`
}

// RAGDocument is a piece of custom knowledge injected into AI context.
type RAGDocument struct {
	basemodel.BaseModel
	UserID  string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Title   string `gorm:"type:varchar(255)" json:"title"`
	Content string `gorm:"type:text" json:"content"`
}

// AIMonitorAlert is a stored finding from the background AI monitoring service.
type AIMonitorAlert struct {
	basemodel.BaseModel
	UserID       string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	DeploymentID string `gorm:"index;type:varchar(255)" json:"deployment_id"`
	Severity     string `gorm:"type:varchar(50)" json:"severity"` // info | warn | error | critical
	Title        string `gorm:"type:varchar(255)" json:"title"`
	Message      string `gorm:"type:text" json:"message"`
	Resolved     bool   `gorm:"default:false" json:"resolved"`
}

// AITokenUsage tracks per-user daily usage of the platform's global AI key.
// Users with their own API key are exempt from platform rate limits.
type AITokenUsage struct {
	basemodel.BaseModel
	UserID string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Date   string `gorm:"index;type:varchar(50)" json:"date"` // YYYY-MM-DD UTC
	Calls  int    `gorm:"default:0" json:"calls"`
	Tokens int    `gorm:"default:0" json:"tokens"`
}

type K8sConfig struct {
	basemodel.BaseModel
	UserID      string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Name        string `gorm:"type:varchar(255)" json:"name"`
	ServerURL   string `gorm:"type:varchar(255)" json:"server_url"`
	Token       string `gorm:"type:text" json:"-"`
	TokenMasked string `gorm:"-" json:"token_masked"` // set by handler
	Namespace   string `gorm:"type:varchar(255)" json:"namespace"`
	Kubeconfig  string `gorm:"type:text" json:"-"` // raw kubeconfig YAML
	RegistryURL string `gorm:"type:varchar(255)" json:"registry_url"`
	Enabled     bool   `gorm:"default:false" json:"enabled"`
}
