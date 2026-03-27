package models

import "github.com/vikukumar/pushpaka/pkg/basemodel"

type Domain struct {
	basemodel.BaseModel
	ProjectID  string `gorm:"index;type:varchar(255);not null" json:"project_id"`
	UserID     string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Domain     string `gorm:"uniqueIndex;type:varchar(255);not null" json:"domain"`
	Verified   bool   `gorm:"default:false" json:"verified"`
	SSLEnabled bool   `gorm:"default:false" json:"ssl_enabled"`
}

type AddDomainRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Domain    string `json:"domain"     binding:"required"`
}
