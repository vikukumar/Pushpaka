package models

import (
	"time"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
)

type CommitStatus string

const (
	CommitStatusSynced    CommitStatus = "synced"
	CommitStatusBuilding  CommitStatus = "building"
	CommitStatusBuilt     CommitStatus = "built"
	CommitStatusTesting   CommitStatus = "testing"
	CommitStatusTested    CommitStatus = "tested"
	CommitStatusFailed    CommitStatus = "failed"
	CommitStatusApproved  CommitStatus = "approved"
	CommitStatusDeployed  CommitStatus = "deployed"
)

type ProjectCommit struct {
	basemodel.BaseModel
	ProjectID string `gorm:"index;type:varchar(255);not null" json:"project_id"`
	
	SHA       string `gorm:"type:varchar(100);not null" json:"sha"`
	Message   string `gorm:"type:text" json:"message"`
	Author    string `gorm:"type:varchar(255)" json:"author"`
	Branch    string `gorm:"type:varchar(100)" json:"branch"`
	
	SourcePath string `gorm:"type:varchar(512)" json:"source_path"`
	BuildPath  string `gorm:"type:varchar(512)" json:"build_path"`
	
	Status        CommitStatus `gorm:"type:varchar(50);default:'synced'" json:"status"`
	TestSummary   string       `gorm:"type:text" json:"test_summary"`
	BuildLogs     string       `gorm:"type:text" json:"build_logs"`
	TestLogs      string       `gorm:"type:text" json:"test_logs"`
	
	IsRecommended bool `gorm:"default:false" json:"is_recommended"`
	
	CommittedAt time.Time `json:"committed_at"`
}
