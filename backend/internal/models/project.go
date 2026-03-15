package models

import "time"

type Project struct {
	ID           string    `db:"id"           json:"id"`
	UserID       string    `db:"user_id"      json:"user_id"`
	Name         string    `db:"name"         json:"name"`
	RepoURL      string    `db:"repo_url"     json:"repo_url"`
	Branch       string    `db:"branch"       json:"branch"`
	BuildCommand string    `db:"build_command" json:"build_command"`
	StartCommand string    `db:"start_command" json:"start_command"`
	Port         int       `db:"port"         json:"port"`
	Framework    string    `db:"framework"    json:"framework"`
	Status       string    `db:"status"       json:"status"` // active | inactive | building
	CreatedAt    time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"   json:"updated_at"`
}

type CreateProjectRequest struct {
	Name         string `json:"name"          binding:"required,min=2,max=64"`
	RepoURL      string `json:"repo_url"      binding:"required,url"`
	Branch       string `json:"branch"`
	BuildCommand string `json:"build_command"`
	StartCommand string `json:"start_command"`
	Port         int    `json:"port"`
	Framework    string `json:"framework"`
}
