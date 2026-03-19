package models

import "time"

type UserEditorState struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index:idx_user_project" json:"user_id"`
	ProjectID string    `gorm:"type:uuid;not null;index:idx_user_project" json:"project_id"`
	OpenTabs  string    `gorm:"type:text" json:"open_tabs"` // JSON string of paths
	ActiveTab string    `gorm:"type:text" json:"active_tab"`
	Sidebar   string    `gorm:"type:text" json:"sidebar"` // JSON string for explorer state
	UpdatedAt time.Time `json:"updated_at"`
}
