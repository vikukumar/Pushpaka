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
	// Resource limits -- passed through to `docker run` flags.
	// Empty string means "no limit" (Docker default).
	// Examples: CPULimit="0.5" (half a core), MemoryLimit="512m".
	CPULimit      string `db:"cpu_limit"      json:"cpu_limit"`
	MemoryLimit   string `db:"memory_limit"   json:"memory_limit"`
	RestartPolicy string `db:"restart_policy" json:"restart_policy"`
	CreatedAt Time `db:"created_at" json:"created_at"`
	UpdatedAt Time `db:"updated_at" json:"updated_at"`
}

type CreateProjectRequest struct {
	Name          string `json:"name"          binding:"required,min=2,max=64"`
	RepoURL       string `json:"repo_url"      binding:"required,url"`
	Branch        string `json:"branch"`
	BuildCommand  string `json:"build_command"`
	StartCommand  string `json:"start_command"`
	Port          int    `json:"port"`
	Framework     string `json:"framework"`
	IsPrivate     bool   `json:"is_private"`
	GitToken      string `json:"git_token"`
	CPULimit      string `json:"cpu_limit"`
	MemoryLimit   string `json:"memory_limit"`
	RestartPolicy string `json:"restart_policy"`
}

// UpdateProjectRequest allows updating mutable project fields.
type UpdateProjectRequest struct {
	Name         string `json:"name"`
	Branch       string `json:"branch"`
	BuildCommand string `json:"build_command"`
	StartCommand string `json:"start_command"`
	Port         int    `json:"port"`
	Framework    string `json:"framework"`
	IsPrivate    bool   `json:"is_private"`
	// GitToken -- if empty the existing stored token is preserved.
	GitToken      string `json:"git_token"`
	CPULimit      string `json:"cpu_limit"`
	MemoryLimit   string `json:"memory_limit"`
	RestartPolicy string `json:"restart_policy"`
}

