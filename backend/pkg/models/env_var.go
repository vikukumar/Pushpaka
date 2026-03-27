package models

import (
	"time"

	"github.com/vikukumar/pushpaka/pkg/basemodel"
)

type EnvVar struct {
	basemodel.BaseModel
	ProjectID string `gorm:"index;type:varchar(255);not null" json:"project_id"`
	UserID    string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Key       string `gorm:"type:varchar(255);not null" json:"key"`
	Value     string `gorm:"type:text" json:"-"`
}

type EnvVarResponse struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Key       string    `json:"key"`
	HasValue  bool      `json:"has_value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
