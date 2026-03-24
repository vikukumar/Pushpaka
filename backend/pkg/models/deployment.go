package models

import (
	"time"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
)

// DeploymentStatus represents the state of a deployment
type DeploymentStatus string

const (
	DeploymentQueued   DeploymentStatus = "queued"
	DeploymentBuilding DeploymentStatus = "building"
	DeploymentRunning  DeploymentStatus = "running"
	DeploymentFailed   DeploymentStatus = "failed"
	DeploymentStopped  DeploymentStatus = "stopped"
)

type Deployment struct {
	basemodel.BaseModel
	ProjectID    string           `gorm:"index;type:varchar(255);not null" json:"project_id"`
	UserID       string           `gorm:"index;type:varchar(255);not null" json:"user_id"`
	WorkerID     string           `gorm:"index;type:varchar(255)" json:"worker_id"`
	CommitSHA    string           `gorm:"type:varchar(255)" json:"commit_sha"`
	CommitMsg    string           `gorm:"type:text" json:"commit_msg"`
	Branch       string           `gorm:"type:varchar(100)" json:"branch"`
	Status       DeploymentStatus `gorm:"type:varchar(50)" json:"status"`
	ImageTag     string           `gorm:"type:varchar(255)" json:"image_tag"`
	ContainerID  string           `gorm:"type:varchar(255)" json:"container_id"`
	URL          string           `gorm:"type:varchar(255)" json:"url"`
	ExternalPort int              `gorm:"default:0" json:"external_port"`
	IsDefault    bool             `gorm:"default:false;index" json:"is_default"` // This is the live/production deployment
	ErrorMsg     string           `gorm:"type:text" json:"error_msg"`
	Resolution   string           `gorm:"type:text" json:"resolution"`
	StartedAt    *time.Time       `json:"started_at"`
	FinishedAt   *time.Time       `json:"finished_at"`
}

type DeployRequest struct {
	ProjectID     string `json:"project_id" binding:"required"`
	Branch        string `json:"branch"`
	CommitSHA     string `json:"commit_sha"`
	CommitMsg     string `json:"commit_msg"`
	ShouldPromote bool   `json:"should_promote"` // Auto-promote to live/default on success
	IsBuildOnly   bool   `json:"is_build_only"`  // Sync and build only, do not deploy
}

type DeploymentJob struct {
	DeploymentID   string            `json:"deployment_id"`
	ProjectID      string            `json:"project_id"`
	UserID         string            `json:"user_id"`
	WorkerID       string            `json:"worker_id,omitempty"`
	RepoURL        string            `json:"repo_url"`
	Branch         string            `json:"branch"`
	CommitSHA      string            `json:"commit_sha"`
	CommitMsg      string            `json:"commit_msg"`
	InstallCommand string            `json:"install_command,omitempty"`
	BuildCommand   string            `json:"build_command"`
	StartCommand   string            `json:"start_command"`
	TestCommand    string            `json:"test_command"`
	RunDir         string            `json:"run_dir,omitempty"`
	Port           int               `json:"port"`
	ExternalPort   int               `json:"external_port"` // The host port to bind to
	EnvVars        map[string]string `json:"env_vars"`
	ImageTag       string            `json:"image_tag"`
	// GitToken is the PAT for private-repo cloning. Never logged or stored in deployment records.
	GitToken  string `json:"git_token,omitempty"`
	IsPrivate bool   `json:"is_private,omitempty"`
	// Resource limits for the Docker container (empty = Docker defaults)
	CPULimit      string `json:"cpu_limit,omitempty"`
	MemoryLimit   string `json:"memory_limit,omitempty"`
	RestartPolicy string `json:"restart_policy,omitempty"`
	IsRecovery    bool   `json:"is_recovery,omitempty"`
	IsBuildOnly   bool   `json:"is_build_only,omitempty"`
	ShouldPromote bool   `json:"should_promote,omitempty"` // Promote on success (auto-set for sync/restart)
	// NotificationURL is an internal callback URL that the worker POSTs to
	// when the deployment finishes (success or failure). This triggers all
	// enabled notification channels (Slack, Discord, SMTP) without the worker
	// needing direct SMTP / webhook credentials.
	NotificationURL string `json:"notification_url,omitempty"`
}
