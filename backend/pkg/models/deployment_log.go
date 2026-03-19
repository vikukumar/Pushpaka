package models

import "github.com/vikukumar/Pushpaka/pkg/basemodel"

type DeploymentLog struct {
	basemodel.BaseModel
	DeploymentID string `gorm:"index;type:varchar(255);not null" json:"deployment_id"`
	Level        string `gorm:"type:varchar(50)" json:"level"`
	Message      string `gorm:"type:text" json:"message"`
	Stream       string `gorm:"type:varchar(50)" json:"stream"`
}
