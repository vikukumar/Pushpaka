package models

import "time"

type EnvVar struct {
	ID        string    `db:"id"         json:"id"`
	ProjectID string    `db:"project_id" json:"project_id"`
	UserID    string    `db:"user_id"    json:"user_id"`
	Key       string    `db:"key"        json:"key"`
	Value     string    `db:"value"      json:"-"` // never returned in JSON
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// EnvVarResponse is the safe response that masks the value
type EnvVarResponse struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Key       string    `json:"key"`
	HasValue  bool      `json:"has_value"`
	CreatedAt time.Time `json:"created_at"`
}

type SetEnvVarRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Key       string `json:"key"        binding:"required"`
	Value     string `json:"value"      binding:"required"`
}

type DeleteEnvVarRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Key       string `json:"key"        binding:"required"`
}
