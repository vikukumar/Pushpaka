package models

import "time"

type DeploymentLog struct {
	ID           string    `db:"id"            json:"id"`
	DeploymentID string    `db:"deployment_id" json:"deployment_id"`
	Level        string    `db:"level"         json:"level"` // info | error | debug
	Message      string    `db:"message"       json:"message"`
	Stream       string    `db:"stream"        json:"stream"` // stdout | stderr | system
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}
