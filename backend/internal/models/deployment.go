package models

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
	ID           string           `db:"id"            json:"id"`
	ProjectID    string           `db:"project_id"    json:"project_id"`
	UserID       string           `db:"user_id"       json:"user_id"`
	CommitSHA    string           `db:"commit_sha"    json:"commit_sha"`
	CommitMsg    string           `db:"commit_msg"    json:"commit_msg"`
	Branch       string           `db:"branch"        json:"branch"`
	Status       DeploymentStatus `db:"status"        json:"status"`
	ImageTag     string           `db:"image_tag"     json:"image_tag"`
	ContainerID  string           `db:"container_id"  json:"container_id"`
	URL          string           `db:"url"           json:"url"`
	ExternalPort int              `db:"external_port" json:"external_port"`
	ErrorMsg     string           `db:"error_msg"     json:"error_msg"`
	StartedAt    *Time            `db:"started_at"    json:"started_at"`
	FinishedAt   *Time            `db:"finished_at"   json:"finished_at"`
	CreatedAt    Time             `db:"created_at"    json:"created_at"`
	UpdatedAt    Time             `db:"updated_at"    json:"updated_at"`
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
	// GitToken is the PAT for private-repo cloning. Never logged or stored in deployment records.
	GitToken string `json:"git_token,omitempty"`
	// Resource limits for the Docker container (empty = Docker defaults)
	CPULimit      string `json:"cpu_limit,omitempty"`
	MemoryLimit   string `json:"memory_limit,omitempty"`
	RestartPolicy string `json:"restart_policy,omitempty"`
	// NotificationURL is an internal callback URL that the worker POSTs to
	// when the deployment finishes (success or failure). This triggers all
	// enabled notification channels (Slack, Discord, SMTP) without the worker
	// needing direct SMTP / webhook credentials.
	NotificationURL string `json:"notification_url,omitempty"`
}
