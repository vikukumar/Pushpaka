package models

import (
	"time"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
)

// DeploymentRole represents the role of a deployment in the project
type DeploymentRole string

const (
	DeploymentRoleMain    DeploymentRole = "main"    // Production deployment
	DeploymentRoleTesting DeploymentRole = "testing" // Testing/staging deployment
	DeploymentRoleBackup  DeploymentRole = "backup"  // Backup for rollback
)

// DeploymentCodeSignature captures code state at deployment time
type DeploymentCodeSignature struct {
	basemodel.BaseModel
	DeploymentID  string `gorm:"index;type:varchar(255);not null" json:"deployment_id"`
	ProjectID     string `gorm:"index;type:varchar(255);not null" json:"project_id"`
	CommitSHA     string `gorm:"type:varchar(255)" json:"commit_sha"`
	CommitMessage string `gorm:"type:text" json:"commit_message"`
	CommitAuthor  string `gorm:"type:varchar(255)" json:"commit_author"`
	Branch        string `gorm:"type:varchar(100)" json:"branch"`
	CodeHash      string `gorm:"type:varchar(255)" json:"code_hash"`      // SHA256 of entire codebase
	FileCount     int    `gorm:"default:0" json:"file_count"`             // Number of files
	DirectoryPath string `gorm:"type:varchar(255)" json:"directory_path"` // Where code is stored
}

// DeploymentInstance represents a single deployment instance
type DeploymentInstance struct {
	basemodel.BaseModel
	DeploymentID    string         `gorm:"index;type:varchar(255);not null" json:"deployment_id"` // Parent deployment
	ProjectID       string         `gorm:"index;type:varchar(255);not null" json:"project_id"`
	Role            DeploymentRole `gorm:"type:varchar(50)" json:"role"`   // main/testing/backup
	Status          string         `gorm:"type:varchar(50)" json:"status"` // running/stopped/failed/testing
	ContainerID     string         `gorm:"type:varchar(255)" json:"container_id"`
	ProcessID       int            `gorm:"default:0" json:"process_id"`
	Port            int            `gorm:"default:0" json:"port"`
	CodeSignatureID string         `gorm:"type:varchar(255)" json:"code_signature_id"`
	InstanceDir     string         `gorm:"type:varchar(255)" json:"instance_dir"` // Working directory for this instance
	StartedAt       *time.Time     `json:"started_at"`
	StoppedAt       *time.Time     `json:"stopped_at"`
	LastHealthCheck *time.Time     `json:"last_health_check"`
	HealthStatus    string         `gorm:"type:varchar(50)" json:"health_status"` // healthy/unhealthy/unknown
	RestartCount    int            `gorm:"default:0" json:"restart_count"`
	ErrorLog        string         `gorm:"type:text" json:"error_log"` // Latest error
}

// DeploymentBackup represents a backup of a deployment
type DeploymentBackup struct {
	basemodel.BaseModel
	DeploymentID    string     `gorm:"index;type:varchar(255);not null" json:"deployment_id"`
	ProjectID       string     `gorm:"index;type:varchar(255);not null" json:"project_id"`
	CodeSignatureID string     `gorm:"type:varchar(255)" json:"code_signature_id"`
	InstanceID      string     `gorm:"type:varchar(255)" json:"instance_id"`
	BackupPath      string     `gorm:"type:varchar(255)" json:"backup_path"` // Directory where backup is stored
	Size            int64      `gorm:"default:0" json:"size"`                // Backup size in bytes
	Reason          string     `gorm:"type:text" json:"reason"`              // Why backup was created (rollback/upgrade)
	IsRestored      bool       `gorm:"default:false" json:"is_restored"`     // Whether this backup was restored
	RestoredAt      *time.Time `json:"restored_at"`
}

// ProjectDeploymentStats tracks deployment statistics for a project
type ProjectDeploymentStats struct {
	basemodel.BaseModel
	ProjectID           string     `gorm:"uniqueIndex;type:varchar(255);not null" json:"project_id"`
	MainDeploymentID    string     `gorm:"type:varchar(255)" json:"main_deployment_id"`
	TestingDeploymentID string     `gorm:"type:varchar(255)" json:"testing_deployment_id"`
	TotalDeployments    int        `gorm:"default:0" json:"total_deployments"`
	SuccessfulDeploys   int        `gorm:"default:0" json:"successful_deploys"`
	FailedDeploys       int        `gorm:"default:0" json:"failed_deploys"`
	TotalBackups        int        `gorm:"default:0" json:"total_backups"`
	LastDeployAt        *time.Time `json:"last_deploy_at"`
	AvgDeployTime       int        `gorm:"default:0" json:"avg_deploy_time"` // in seconds
}

// DeploymentAction represents user actions on deployments
type DeploymentActionType string

const (
	DeploymentActionStart    DeploymentActionType = "start"
	DeploymentActionStop     DeploymentActionType = "stop"
	DeploymentActionRestart  DeploymentActionType = "restart"
	DeploymentActionRetry    DeploymentActionType = "retry"
	DeploymentActionRollback DeploymentActionType = "rollback"
	DeploymentActionSync     DeploymentActionType = "sync"
)

type DeploymentAction struct {
	basemodel.BaseModel
	DeploymentID string               `gorm:"index;type:varchar(255);not null" json:"deployment_id"`
	InstanceID   string               `gorm:"type:varchar(255)" json:"instance_id"`
	ProjectID    string               `gorm:"index;type:varchar(255);not null" json:"project_id"`
	UserID       string               `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Action       DeploymentActionType `gorm:"type:varchar(50)" json:"action"`
	Status       string               `gorm:"type:varchar(50)" json:"status"` // pending/executing/success/failed
	Result       string               `gorm:"type:text" json:"result"`        // Result message
}

// API Request/Response types

// DeploymentConfig represents deployment configuration
type DeploymentConfig struct {
	MaxDeployments int `json:"max_deployments"` // Number of concurrent deployments
	MaxBackups     int `json:"max_backups"`     // Number of backups to keep
}

// CreateDeploymentRequest for creating new deployment
type CreateDeploymentRequest struct {
	ProjectID    string `json:"project_id" binding:"required"`
	DeploymentID string `json:"deployment_id" binding:"required"` // Will be created if not exists
}

// DeploymentActionRequest for performing actions
type DeploymentActionRequest struct {
	Action   string `json:"action" binding:"required"` // start/stop/restart/retry/rollback
	Reason   string `json:"reason"`
	BackupID string `json:"backup_id"` // For rollback action
}

// SwitchMainDeploymentRequest for promoting testing to main
type SwitchMainDeploymentRequest struct {
	NewMainDeploymentID  string `json:"new_main_deployment_id" binding:"required"`
	CreateBackup         bool   `json:"create_backup"`
	BackupCurrentMain    bool   `json:"backup_current_main"`
	GracefulShutdownTime int    `json:"graceful_shutdown_time"` // Seconds to wait before forcing shutdown
}

// DeploymentResponse for API responses
type DeploymentResponse struct {
	ID            string                   `json:"id"`
	ProjectID     string                   `json:"project_id"`
	Role          DeploymentRole           `json:"role"`
	Status        string                   `json:"status"`
	Port          int                      `json:"port"`
	HealthStatus  string                   `json:"health_status"`
	CodeSignature *DeploymentCodeSignature `json:"code_signature"`
	Instances     []DeploymentInstance     `json:"instances"`
	Backups       []DeploymentBackup       `json:"backups"`
	RestartCount  int                      `json:"restart_count"`
	LastAction    *DeploymentAction        `json:"last_action"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}
