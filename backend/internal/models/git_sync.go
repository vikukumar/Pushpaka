package models

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
	ID          string        `db:"id"            json:"id"`
	SyncTrackID string        `db:"sync_track_id" json:"sync_track_id"`
	FilePath    string        `db:"file_path"     json:"file_path"`
	ChangeType  GitChangeType `db:"change_type"   json:"change_type"`
	Additions   int           `db:"additions"     json:"additions"`
	Deletions   int           `db:"deletions"     json:"deletions"`
	OldContent  string        `db:"old_content"   json:"old_content,omitempty"`
	NewContent  string        `db:"new_content"   json:"new_content,omitempty"`
	CreatedAt   Time          `db:"created_at"    json:"created_at"`
}

// GitSyncTrack represents git synchronization tracking for a deployment
type GitSyncTrack struct {
	ID                   string        `db:"id"                      json:"id"`
	DeploymentID         string        `db:"deployment_id"           json:"deployment_id"`
	ProjectID            string        `db:"project_id"              json:"project_id"`
	Repository           string        `db:"repository"              json:"repository"`
	Branch               string        `db:"branch"                  json:"branch"`
	CurrentCommitSHA     string        `db:"current_commit_sha"      json:"current_commit_sha"`
	LatestCommitSHA      string        `db:"latest_commit_sha"       json:"latest_commit_sha"`
	LatestCommitMessage  string        `db:"latest_commit_message"   json:"latest_commit_message"`
	LatestCommitAuthor   string        `db:"latest_commit_author"    json:"latest_commit_author"`
	SyncStatus           GitSyncStatus `db:"sync_status"             json:"sync_status"`
	SyncApprovalRequired bool          `db:"sync_approval_required"  json:"sync_approval_required"`
	SyncApprovedBy       string        `db:"sync_approved_by"        json:"sync_approved_by,omitempty"`
	SyncApprovedAt       *Time         `db:"sync_approved_at"        json:"sync_approved_at,omitempty"`
	TotalChanges         int           `db:"total_changes"           json:"total_changes"`
	TotalAdditions       int           `db:"total_additions"         json:"total_additions"`
	TotalDeletions       int           `db:"total_deletions"         json:"total_deletions"`
	ChangesSummary       string        `db:"changes_summary"         json:"changes_summary"` // JSON
	LastSyncAttemptAt    *Time         `db:"last_sync_attempt_at"    json:"last_sync_attempt_at,omitempty"`
	LastSyncAttemptError string        `db:"last_sync_attempt_error" json:"last_sync_attempt_error,omitempty"`
	LastSuccessfulSyncAt *Time         `db:"last_successful_sync_at" json:"last_successful_sync_at,omitempty"`
	NotificationSentAt   *Time         `db:"notification_sent_at"    json:"notification_sent_at,omitempty"`
	CreatedAt            Time          `db:"created_at"              json:"created_at"`
	UpdatedAt            Time          `db:"updated_at"              json:"updated_at"`
}

// GitCommitInfo represents commit metadata
type GitCommitInfo struct {
	SHA       string `json:"sha"`
	Author    string `json:"author"`
	Message   string `json:"message"`
	Timestamp Time   `json:"timestamp"`
	URL       string `json:"url"`
	Tags      []string
	ParentSHA string `json:"parent_sha"`
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
	ID                string `db:"id"                   json:"id"`
	ProjectID         string `db:"project_id"           json:"project_id"`
	DeploymentID      string `db:"deployment_id"        json:"deployment_id"`
	Enabled           bool   `db:"enabled"              json:"enabled"`
	RequireApproval   bool   `db:"require_approval"     json:"require_approval"`
	PollingInterval   int    `db:"polling_interval"     json:"polling_interval"`   // seconds
	MaxConcurrent     int    `db:"max_concurrent"       json:"max_concurrent"`     // max parallel syncs
	OnlyProdReady     bool   `db:"only_prod_ready"      json:"only_prod_ready"`    // only sync tagged releases
	AllowedBranches   string `db:"allowed_branches"     json:"allowed_branches"`   // JSON array
	IgnorePaths       string `db:"ignore_paths"         json:"ignore_paths"`       // JSON array
	RequiredApprovers string `db:"required_approvers"   json:"required_approvers"` // JSON array
	CreatedAt         Time   `db:"created_at"           json:"created_at"`
	UpdatedAt         Time   `db:"updated_at"           json:"updated_at"`
}

// DeploymentSyncHistory represents historical sync records
type DeploymentSyncHistory struct {
	ID                string `db:"id"               json:"id"`
	DeploymentID      string `db:"deployment_id"    json:"deployment_id"`
	ProjectID         string `db:"project_id"       json:"project_id"`
	FromCommitSHA     string `db:"from_commit_sha"  json:"from_commit_sha"`
	ToCommitSHA       string `db:"to_commit_sha"    json:"to_commit_sha"`
	SyncType          string `db:"sync_type"        json:"sync_type"` // "manual", "auto", "webhook"
	Status            string `db:"status"           json:"status"`    // "success", "failed", "rolled_back"
	TotalChanges      int    `db:"total_changes"    json:"total_changes"`
	Duration          int    `db:"duration"         json:"duration"` // seconds
	TriggeredBy       string `db:"triggered_by"     json:"triggered_by"`
	SyncError         string `db:"sync_error"       json:"sync_error,omitempty"`
	RollbackTriggered bool   `db:"rollback_triggered" json:"rollback_triggered"`
	RollbackReason    string `db:"rollback_reason"  json:"rollback_reason,omitempty"`
	CreatedAt         Time   `db:"created_at"       json:"created_at"`
	UpdatedAt         Time   `db:"updated_at"       json:"updated_at"`
}
