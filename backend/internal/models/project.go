package models

type Project struct {
	ID           string `db:"id"            json:"id"`
	UserID       string `db:"user_id"       json:"user_id"`
	Name         string `db:"name"          json:"name"`
	RepoURL      string `db:"repo_url"      json:"repo_url"`
	Branch       string `db:"branch"        json:"branch"`
	BuildCommand string `db:"build_command" json:"build_command"`
	StartCommand string `db:"start_command" json:"start_command"`
	Port         int    `db:"port"          json:"port"`
	Framework    string `db:"framework"     json:"framework"`
	Status       string `db:"status"        json:"status"`
	// GitToken is the personal access token for cloning private repositories.
	// It is NEVER serialised into API responses (json:"-").
	GitToken  string `db:"git_token"     json:"-"`
	IsPrivate bool   `db:"is_private"    json:"is_private"`
	// Build configuration
	InstallCommand string `db:"install_command" json:"install_command"`
	// RunDir is the subdirectory within the repo to use as the working directory.
	RunDir string `db:"run_dir" json:"run_dir"`
	// Resource limits -- passed through to `docker run` flags.
	// Empty string means "no limit" (Docker default).
	// Examples: CPULimit="0.5" (half a core), MemoryLimit="512m".
	CPULimit      string `db:"cpu_limit"      json:"cpu_limit"`
	MemoryLimit   string `db:"memory_limit"   json:"memory_limit"`
	RestartPolicy string `db:"restart_policy" json:"restart_policy"`
	// DeployTarget determines where the project runs: "docker" (default) or "kubernetes".
	DeployTarget string `db:"deploy_target"  json:"deploy_target"`
	K8sNamespace string `db:"k8s_namespace"  json:"k8s_namespace"`
	// Deployment limits and backup configuration
	MaxDeployments int    `db:"max_deployments" json:"max_deployments"` // Max simultaneous deployments (default: 2)
	MaxBackups     int    `db:"max_backups"     json:"max_backups"`     // Max backup versions kept (default: 3)
	CloneDirectory string `db:"clone_directory" json:"clone_directory"` // Base directory for git clones
	GitClonePath   string `db:"git_clone_path"  json:"git_clone_path"`  // Current git clone directory for project
	MainDeployID   string `db:"main_deploy_id"  json:"main_deploy_id"`  // Current main deployment ID
	CreatedAt      Time   `db:"created_at" json:"created_at"`
	UpdatedAt      Time   `db:"updated_at" json:"updated_at"`
}

type CreateProjectRequest struct {
	Name           string `json:"name"            binding:"required,min=2,max=64"`
	RepoURL        string `json:"repo_url"        binding:"required,url"`
	Branch         string `json:"branch"`
	InstallCommand string `json:"install_command"`
	BuildCommand   string `json:"build_command"`
	StartCommand   string `json:"start_command"`
	RunDir         string `json:"run_dir"`
	Port           int    `json:"port"`
	Framework      string `json:"framework"`
	IsPrivate      bool   `json:"is_private"`
	GitToken       string `json:"git_token"`
	CPULimit       string `json:"cpu_limit"`
	MemoryLimit    string `json:"memory_limit"`
	RestartPolicy  string `json:"restart_policy"`
	DeployTarget   string `json:"deploy_target"`
	K8sNamespace   string `json:"k8s_namespace"`
	MaxDeployments int    `json:"max_deployments"` // Default: 2 (1 main + 1 testing)
	MaxBackups     int    `json:"max_backups"`     // Default: 3
}

// UpdateProjectRequest allows updating mutable project fields.
type UpdateProjectRequest struct {
	Name           string `json:"name"`
	RepoURL        string `json:"repo_url"`
	Branch         string `json:"branch"`
	InstallCommand string `json:"install_command"`
	BuildCommand   string `json:"build_command"`
	StartCommand   string `json:"start_command"`
	RunDir         string `json:"run_dir"`
	Port           int    `json:"port"`
	Framework      string `json:"framework"`
	IsPrivate      bool   `json:"is_private"`
	// GitToken -- if empty the existing stored token is preserved.
	GitToken       string `json:"git_token"`
	CPULimit       string `json:"cpu_limit"`
	MemoryLimit    string `json:"memory_limit"`
	RestartPolicy  string `json:"restart_policy"`
	DeployTarget   string `json:"deploy_target"`
	K8sNamespace   string `json:"k8s_namespace"`
	MaxDeployments int    `json:"max_deployments"`
	MaxBackups     int    `json:"max_backups"`
}
