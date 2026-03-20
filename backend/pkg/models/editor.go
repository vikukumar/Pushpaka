package models

import (
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
)

type UserEditorState struct {
	basemodel.BaseModel
	UserID    string `gorm:"index:idx_user_project;type:varchar(255);not null" json:"user_id"`
	ProjectID string `gorm:"index:idx_user_project;type:varchar(255);not null" json:"project_id"`
	OpenTabs  string `gorm:"type:text" json:"open_tabs"` // JSON string of paths
	ActiveTab string `gorm:"type:text" json:"active_tab"`
	Sidebar   string `gorm:"type:text" json:"sidebar"` // JSON string for explorer state
}
