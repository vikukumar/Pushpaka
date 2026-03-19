package models

import (
	"time"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
)

// GitSyncStatus represents the synchronization status between deployment and git
type GitSyncStatus string

const (
	GitSyncSynced    GitSyncStatus = "synced"      // Deployment matches git commit
	GitSyncOutOfSync GitSyncStatus = "out_of_sync" // New commits available
	GitSyncSyncing   GitSyncStatus = "syncing"     // Currently syncing
	GitSyncFailed    GitSyncStatus = "failed"      // Sync failed
	GitSyncPending   GitSyncStatus = "pending"     // Waiting for approval to sync
)

// GitChangeType represents the type of git change
type GitChangeType string

const (
	GitChangeTypeAdded    GitChangeType = "added"
	GitChangeTypeModified GitChangeType = "modified"
	GitChangeTypeDeleted  GitChangeType = "deleted"
	GitChangeTypeRenamed  GitChangeType = "renamed"
)

// GitChange represents a single file change in git
type GitChange struct {
	basemodel.BaseModel
	SyncTrackID string        `gorm:"index;type:varchar(255);not null" json:"sync_track_id"`
	FilePath    string        `gorm:"type:text" json:"file_path"`
	ChangeType  GitChangeType `gorm:"type:varchar(50)" json:"change_type"`
	Additions   int           `gorm:"default:0" json:"additions"`
	Deletions   int           `gorm:"default:0" json:"deletions"`
	OldContent  string        `gorm:"type:text" json:"old_content,omitempty"`
	NewContent  string        `gorm:"type:text" json:"new_content,omitempty"`
}

// GitSyncTrack represents git synchronization tracking for a deployment
type GitSyncTrack struct {
	basemodel.BaseModel
	DeploymentID         string        `gorm:"index;type:varchar(255);not null" json:"deployment_id"`
	ProjectID            string        `gorm:"index;type:varchar(255);not null" json:"project_id"`
	Repository           string        `gorm:"type:varchar(255)" json:"repository"`
	Branch               string        `gorm:"type:varchar(100)" json:"branch"`
	CurrentCommitSHA     string        `gorm:"type:varchar(255)" json:"current_commit_sha"`
	LatestCommitSHA      string        `gorm:"type:varchar(255)" json:"latest_commit_sha"`
	LatestCommitMessage  string        `gorm:"type:text" json:"latest_commit_message"`
	LatestCommitAuthor   string        `gorm:"type:varchar(255)" json:"latest_commit_author"`
	SyncStatus           GitSyncStatus `gorm:"type:varchar(50)" json:"sync_status"`
	SyncApprovalRequired bool          `gorm:"default:false" json:"sync_approval_required"`
	SyncApprovedBy       string        `gorm:"type:varchar(255)" json:"sync_approved_by,omitempty"`
	SyncApprovedAt       *time.Time    `json:"sync_approved_at,omitempty"`
	TotalChanges         int           `gorm:"default:0" json:"total_changes"`
	TotalAdditions       int           `gorm:"default:0" json:"total_additions"`
	TotalDeletions       int           `gorm:"default:0" json:"total_deletions"`
	ChangesSummary       string        `gorm:"type:text" json:"changes_summary"` // JSON array string
	LastSyncAttemptAt    *time.Time    `json:"last_sync_attempt_at,omitempty"`
	LastSyncAttemptError string        `gorm:"type:text" json:"last_sync_attempt_error,omitempty"`
	LastSuccessfulSyncAt *time.Time    `json:"last_successful_sync_at,omitempty"`
	NotificationSentAt   *time.Time    `json:"notification_sent_at,omitempty"`
}

// GitCommitInfo represents commit metadata
type GitCommitInfo struct {
	SHA       string    `json:"sha"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Tags      []string  `json:"tags,omitempty"`
	ParentSHA string    `json:"parent_sha"`
}

// GitDiffSummary represents a summary of git differences
type GitDiffSummary struct {
	TotalChanges   int           `json:"total_changes"`
	TotalAdditions int           `json:"total_additions"`
	TotalDeletions int           `json:"total_deletions"`
	FilesChanged   int           `json:"files_changed"`
	ChangedFiles   []GitChange   `json:"changed_files"`
	CommitsAhead   int           `json:"commits_ahead"`
	CommitsBehind  int           `json:"commits_behind"`
	IsClean        bool          `json:"is_clean"`
	LastCommit     GitCommitInfo `json:"last_commit"`
}

// SyncRequest represents a request to sync deployment with git
type SyncRequest struct {
	DeploymentID string `json:"deployment_id" binding:"required"`
	ProjectID    string `json:"project_id" binding:"required"`
	Force        bool   `json:"force,omitempty"`     // Skip approval if needed
	Reason       string `json:"reason,omitempty"`    // Approval reason
	AutoSync     bool   `json:"auto_sync,omitempty"` // Auto-sync future changes
}

// SyncApprovalRequest represents approval for pending sync
type SyncApprovalRequest struct {
	SyncTrackID string `json:"sync_track_id" binding:"required"`
	Approved    bool   `json:"approved"`
	Reason      string `json:"reason,omitempty"`
}

// GitAutoSyncConfig represents auto-sync configuration
type GitAutoSyncConfig struct {
	basemodel.BaseModel
	ProjectID         string `gorm:"index;type:varchar(255);not null" json:"project_id"`
	DeploymentID      string `gorm:"index;type:varchar(255);not null" json:"deployment_id"`
	Enabled           bool   `gorm:"default:false" json:"enabled"`
	RequireApproval   bool   `gorm:"default:false" json:"require_approval"`
	PollingInterval   int    `gorm:"default:0" json:"polling_interval"`    // seconds
	MaxConcurrent     int    `gorm:"default:1" json:"max_concurrent"`      // max parallel syncs
	OnlyProdReady     bool   `gorm:"default:false" json:"only_prod_ready"` // only sync tagged releases
	AllowedBranches   string `gorm:"type:text" json:"allowed_branches"`    // JSON array
	IgnorePaths       string `gorm:"type:text" json:"ignore_paths"`        // JSON array
	RequiredApprovers string `gorm:"type:text" json:"required_approvers"`  // JSON array
}

// DeploymentSyncHistory represents historical sync records
type DeploymentSyncHistory struct {
	basemodel.BaseModel
	DeploymentID      string `gorm:"index;type:varchar(255);not null" json:"deployment_id"`
	ProjectID         string `gorm:"index;type:varchar(255);not null" json:"project_id"`
	FromCommitSHA     string `gorm:"type:varchar(255)" json:"from_commit_sha"`
	ToCommitSHA       string `gorm:"type:varchar(255)" json:"to_commit_sha"`
	SyncType          string `gorm:"type:varchar(50)" json:"sync_type"` // "manual", "auto", "webhook"
	Status            string `gorm:"type:varchar(50)" json:"status"`    // "success", "failed", "rolled_back"
	TotalChanges      int    `gorm:"default:0" json:"total_changes"`
	Duration          int    `gorm:"default:0" json:"duration"` // seconds
	TriggeredBy       string `gorm:"type:varchar(255)" json:"triggered_by"`
	SyncError         string `gorm:"type:text" json:"sync_error,omitempty"`
	RollbackTriggered bool   `gorm:"default:false" json:"rollback_triggered"`
	RollbackReason    string `gorm:"type:text" json:"rollback_reason,omitempty"`
}
