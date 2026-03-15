package models

import "time"

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
	ID          string           `db:"id"           json:"id"`
	ProjectID   string           `db:"project_id"   json:"project_id"`
	UserID      string           `db:"user_id"      json:"user_id"`
	CommitSHA   string           `db:"commit_sha"   json:"commit_sha"`
	CommitMsg   string           `db:"commit_msg"   json:"commit_msg"`
	Branch      string           `db:"branch"       json:"branch"`
	Status      DeploymentStatus `db:"status"       json:"status"`
	ImageTag    string           `db:"image_tag"    json:"image_tag"`
	ContainerID string           `db:"container_id" json:"container_id"`
	URL         string           `db:"url"          json:"url"`
	ErrorMsg    string           `db:"error_msg"    json:"error_msg"`
	StartedAt   *time.Time       `db:"started_at"   json:"started_at"`
	FinishedAt  *time.Time       `db:"finished_at"  json:"finished_at"`
	CreatedAt   time.Time        `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"   json:"updated_at"`
}

type DeployRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Branch    string `json:"branch"`
	CommitSHA string `json:"commit_sha"`
}

type DeploymentJob struct {
	DeploymentID string            `json:"deployment_id"`
	ProjectID    string            `json:"project_id"`
	UserID       string            `json:"user_id"`
	RepoURL      string            `json:"repo_url"`
	Branch       string            `json:"branch"`
	CommitSHA    string            `json:"commit_sha"`
	BuildCommand string            `json:"build_command"`
	StartCommand string            `json:"start_command"`
	Port         int               `json:"port"`
	EnvVars      map[string]string `json:"env_vars"`
	ImageTag     string            `json:"image_tag"`
}
