package models

// DeploymentRole represents the role of a deployment in the project
type DeploymentRole string

const (
	DeploymentRoleMain    DeploymentRole = "main"    // Production deployment
	DeploymentRoleTesting DeploymentRole = "testing" // Testing/staging deployment
	DeploymentRoleBackup  DeploymentRole = "backup"  // Backup for rollback
)

// DeploymentCodeSignature captures code state at deployment time
type DeploymentCodeSignature struct {
	ID            string `db:"id"             json:"id"`
	DeploymentID  string `db:"deployment_id"  json:"deployment_id"`
	ProjectID     string `db:"project_id"     json:"project_id"`
	CommitSHA     string `db:"commit_sha"     json:"commit_sha"`
	CommitMessage string `db:"commit_message" json:"commit_message"`
	CommitAuthor  string `db:"commit_author"  json:"commit_author"`
	Branch        string `db:"branch"         json:"branch"`
	CodeHash      string `db:"code_hash"      json:"code_hash"`      // SHA256 of entire codebase
	FileCount     int    `db:"file_count"     json:"file_count"`     // Number of files
	DirectoryPath string `db:"directory_path" json:"directory_path"` // Where code is stored
	CreatedAt     Time   `db:"created_at"     json:"created_at"`
}

// DeploymentInstance represents a single deployment instance
type DeploymentInstance struct {
	ID              string         `db:"id"                json:"id"`
	DeploymentID    string         `db:"deployment_id"     json:"deployment_id"` // Parent deployment
	ProjectID       string         `db:"project_id"        json:"project_id"`
	Role            DeploymentRole `db:"role"              json:"role"`   // main/testing/backup
	Status          string         `db:"status"            json:"status"` // running/stopped/failed/testing
	ContainerID     string         `db:"container_id"      json:"container_id"`
	ProcessID       int            `db:"process_id"        json:"process_id"`
	Port            int            `db:"port"              json:"port"`
	CodeSignatureID string         `db:"code_signature_id" json:"code_signature_id"`
	InstanceDir     string         `db:"instance_dir"      json:"instance_dir"` // Working directory for this instance
	StartedAt       *Time          `db:"started_at"        json:"started_at"`
	StoppedAt       *Time          `db:"stopped_at"        json:"stopped_at"`
	LastHealthCheck *Time          `db:"last_health_check" json:"last_health_check"`
	HealthStatus    string         `db:"health_status"     json:"health_status"` // healthy/unhealthy/unknown
	RestartCount    int            `db:"restart_count"     json:"restart_count"`
	ErrorLog        string         `db:"error_log"         json:"error_log"` // Latest error
	CreatedAt       Time           `db:"created_at"        json:"created_at"`
	UpdatedAt       Time           `db:"updated_at"        json:"updated_at"`
}

// DeploymentBackup represents a backup of a deployment
type DeploymentBackup struct {
	ID              string `db:"id"               json:"id"`
	DeploymentID    string `db:"deployment_id"    json:"deployment_id"`
	ProjectID       string `db:"project_id"       json:"project_id"`
	CodeSignatureID string `db:"code_signature_id" json:"code_signature_id"`
	InstanceID      string `db:"instance_id"      json:"instance_id"`
	BackupPath      string `db:"backup_path"      json:"backup_path"` // Directory where backup is stored
	Size            int64  `db:"size"             json:"size"`        // Backup size in bytes
	Reason          string `db:"reason"           json:"reason"`      // Why backup was created (rollback/upgrade)
	IsRestored      bool   `db:"is_restored"      json:"is_restored"` // Whether this backup was restored
	RestoredAt      *Time  `db:"restored_at"      json:"restored_at"`
	CreatedAt       Time   `db:"created_at"       json:"created_at"`
}

// ProjectDeploymentStats tracks deployment statistics for a project
type ProjectDeploymentStats struct {
	ID                  string `db:"id"                   json:"id"`
	ProjectID           string `db:"project_id"           json:"project_id"`
	MainDeploymentID    string `db:"main_deployment_id"   json:"main_deployment_id"`
	TestingDeploymentID string `db:"testing_deployment_id" json:"testing_deployment_id"`
	TotalDeployments    int    `db:"total_deployments"    json:"total_deployments"`
	SuccessfulDeploys   int    `db:"successful_deploys"   json:"successful_deploys"`
	FailedDeploys       int    `db:"failed_deploys"       json:"failed_deploys"`
	TotalBackups        int    `db:"total_backups"        json:"total_backups"`
	LastDeployAt        *Time  `db:"last_deploy_at"       json:"last_deploy_at"`
	AvgDeployTime       int    `db:"avg_deploy_time"      json:"avg_deploy_time"` // in seconds
	UpdatedAt           Time   `db:"updated_at"           json:"updated_at"`
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
	ID           string               `db:"id"            json:"id"`
	DeploymentID string               `db:"deployment_id" json:"deployment_id"`
	InstanceID   string               `db:"instance_id"   json:"instance_id"`
	ProjectID    string               `db:"project_id"    json:"project_id"`
	UserID       string               `db:"user_id"       json:"user_id"`
	Action       DeploymentActionType `db:"action"        json:"action"`
	Status       string               `db:"status"        json:"status"` // pending/executing/success/failed
	Result       string               `db:"result"        json:"result"` // Result message
	CreatedAt    Time                 `db:"created_at"    json:"created_at"`
	UpdatedAt    Time                 `db:"updated_at"    json:"updated_at"`
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
	CreatedAt     Time                     `json:"created_at"`
	UpdatedAt     Time                     `json:"updated_at"`
}
