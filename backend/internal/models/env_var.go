package models

type EnvVar struct {
	ID        string `db:"id"         json:"id"`
	ProjectID string `db:"project_id" json:"project_id"`
	UserID    string `db:"user_id"    json:"user_id"`
	Key       string `db:"key"        json:"key"`
	Value     string `db:"value"      json:"-"`
	CreatedAt Time   `db:"created_at" json:"created_at"`
	UpdatedAt Time   `db:"updated_at" json:"updated_at"`
}

type EnvVarResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Key       string `json:"key"`
	HasValue  bool   `json:"has_value"`
	CreatedAt Time   `json:"created_at"`
	UpdatedAt Time   `json:"updated_at"`
}

type SetEnvRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Key       string `json:"key"        binding:"required"`
	Value     string `json:"value"      binding:"required"`
}

type DeleteEnvRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Key       string `json:"key"        binding:"required"`
}

// SetEnvVarRequest is an alias kept for service layer compatibility.
type SetEnvVarRequest = SetEnvRequest

// DeleteEnvVarRequest is an alias kept for handler layer compatibility.
type DeleteEnvVarRequest = DeleteEnvRequest
