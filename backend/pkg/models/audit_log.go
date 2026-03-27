package models

import "github.com/vikukumar/pushpaka/pkg/basemodel"

type AuditLog struct {
	basemodel.BaseModel
	UserID     string `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Action     string `gorm:"type:varchar(100)" json:"action"`
	Resource   string `gorm:"type:varchar(100)" json:"resource"`
	ResourceID string `gorm:"type:varchar(255)" json:"resource_id"`
	Metadata   string `gorm:"type:text" json:"metadata"` // raw JSON string
	IPAddr     string `gorm:"type:varchar(100)" json:"ip_addr"`
	UserAgent  string `gorm:"type:text" json:"user_agent"`
}
