# Git Change Tracking & Sync Implementation

Comprehensive guide for tracking git changes and synchronizing deployments with the latest code, similar to ArgoCD.

## Overview

Pushpaka now tracks git changes and automatically syncs deployments with the latest repository code. The system provides:

- **Real-time change detection** - Monitor for new commits
- **Automatic synchronization** - Deploy latest code on schedule or manually
- **Approval workflows** - Optional approval gates before syncing
- **Change visibility** - Detailed diff tracking between versions
- **History & rollback** - Complete audit trail of all syncs
- **ArgoCD-style GitOps** - Git as source of truth

## Architecture

```
┌─────────────────┐
│   Git Remote    │
│  (GitHub/Lab)   │
└────────┬────────┘
         │
         │ Webhook / Polling
         │
┌────────▼──────────────────────────────┐
│   Pushpaka GitSync Service            │
├────────────────────────────────────────┤
│ • Change Detection                     │
│ • Commit Tracking                      │
│ • Diff Generation                      │
│ • Auto-sync Logic                      │
└────────┬──────────────────────────────┘
         │
         │ Update & Track
         │
┌────────▼──────────────────────────────┐
│   Deployment Synchronization          │
├────────────────────────────────────────┤
│ • Git Sync Tracks (current status)     │
│ • Git Changes (file-level diffs)       │
│ • Sync History (audit trail)           │
│ • Auto-sync Config (settings)          │
└────────┬──────────────────────────────┘
         │
         │ Deploy & Run Tests
         │
┌────────▼──────────────────────────────┐
│   Active Deployment                   │
│  (Running Latest Code)                 │
└────────────────────────────────────────┘
```

## Data Models

### GitSyncTrack

Represents the synchronization status of a deployment.

```go
type GitSyncTrack struct {
    ID                      string        // Unique ID
    DeploymentID            string        // Reference to deployment
    ProjectID               string        // Reference to project
    Repository              string        // Git repository URL
    Branch                  string        // Git branch tracking
    
    CurrentCommitSHA        string        // Currently deployed commit
    LatestCommitSHA         string        // Latest commit in repository
    LatestCommitMessage     string        // Latest commit message
    LatestCommitAuthor      string        // Latest commit author
    
    SyncStatus              GitSyncStatus // Status (synced/out_of_sync/syncing/failed)
    SyncApprovalRequired    bool          // Needs approval to sync
    SyncApprovedBy          string        // Who approved the sync
    SyncApprovedAt          *Time         // When approved
    
    TotalChanges            int           // Number of files changed
    TotalAdditions          int           // Lines added across all files
    TotalDeletions          int           // Lines deleted across all files
    ChangesSummary          string        // JSON with detailed stats
    
    LastSyncAttemptAt       *Time         // Last sync attempt timestamp
    LastSyncAttemptError    string        // Error from last attempt
    LastSuccessfulSyncAt    *Time         // Last successful sync
    NotificationSentAt      *Time         // When notification was sent
    
    CreatedAt               Time
    UpdatedAt               Time
}
```

### GitChange

Represents changes to individual files.

```go
type GitChange struct {
    ID              string        // Unique ID
    SyncTrackID     string        // Reference to git sync track
    FilePath        string        // Path of changed file
    ChangeType      GitChangeType // added/modified/deleted/renamed
    
    Additions       int           // Lines added
    Deletions       int           // Lines deleted
    OldContent      string        // Previous content (optional)
    NewContent      string        // New content (optional)
    
    CreatedAt       Time
}
```

### GitAutoSyncConfig

Configuration for automatic synchronization.

```go
type GitAutoSyncConfig struct {
    ID                  string    // Unique ID
    ProjectID           string    // Reference to project
    DeploymentID        string    // Reference to deployment
    
    Enabled             bool      // Is auto-sync active
    RequireApproval     bool      // Need approval before sync
    PollingInterval     int       // Seconds between checks (default: 3600)
    MaxConcurrent       int       // Max parallel syncs (default: 1)
    OnlyProdReady       bool      // Only sync tagged releases
    
    AllowedBranches     string    // JSON array of allowed branches
    IgnorePaths         string    // JSON array of paths to ignore
    RequiredApprovers   string    // JSON array of required approvers
    
    CreatedAt           Time
    UpdatedAt           Time
}
```

## API Endpoints

### 1. Check for Updates

**Endpoint:** `GET /api/v1/deployments/{deploymentId}/git/check-updates`

Check if new commits are available without syncing.

**Response:**
```json
{
  "sync_status": "out_of_sync",
  "latest_sha": "abc123def456",
  "current_sha": "old123commit",
  "total_changes": 15,
  "total_additions": 245,
  "total_deletions": 89,
  "latest_commit_message": "feat: Add new feature",
  "latest_commit_author": "john@example.com"
}
```

### 2. Manual Sync

**Endpoint:** `POST /api/v1/deployments/{deploymentId}/git/sync`

Manually trigger synchronization to latest code.

**Body:**
```json
{
  "deployment_id": "dep_123",
  "project_id": "proj_abc",
  "force": false,
  "reason": "Deploying new feature"
}
```

**Response:**
```json
{
  "status": "syncing",
  "deployment_id": "dep_123",
  "sync_track_id": "sync_xyz"
}
```

### 3. Get Sync Status

**Endpoint:** `GET /api/v1/deployments/{deploymentId}/git/status`

Get current synchronization status.

**Response:**
```json
{
  "deployment_id": "dep_123",
  "sync_status": "synced",
  "current_commit_sha": "abc123",
  "latest_commit_sha": "abc123",
  "is_clean": true,
  "total_changes": 0,
  "last_successful_sync_at": "2026-03-18T10:00:00Z"
}
```

### 4. Get Changes

**Endpoint:** `GET /api/v1/deployments/{deploymentId}/git/changes`

Get detailed file changes between commits.

**Response:**
```json
{
  "deployment_id": "dep_123",
  "total_changes": 15,
  "total_additions": 245,
  "total_deletions": 89,
  "changes": [
    {
      "file_path": "src/main.go",
      "change_type": "modified",
      "additions": 45,
      "deletions": 12
    },
    {
      "file_path": "docs/README.md",
      "change_type": "added",
      "additions": 200,
      "deletions": 0
    }
  ]
}
```

### 5. Get Sync History

**Endpoint:** `GET /api/v1/deployments/{deploymentId}/git/history`

Get historical sync records.

**Query Parameters:**
- `limit`: Number of records (default: 20)
- `offset`: Pagination offset (default: 0)

**Response:**
```json
{
  "deployment_id": "dep_123",
  "total": 42,
  "history": [
    {
      "id": "sync_hist_001",
      "from_commit_sha": "old123",
      "to_commit_sha": "new456",
      "sync_type": "manual",
      "status": "success",
      "total_changes": 15,
      "duration": 45,
      "triggered_by": "john@example.com",
      "created_at": "2026-03-18T10:00:00Z"
    }
  ]
}
```

### 6. Approve Sync

**Endpoint:** `POST /api/v1/deployments/{deploymentId}/git/approve-sync`

Approve a pending synchronization request.

**Body:**
```json
{
  "sync_track_id": "sync_xyz",
  "approved": true,
  "reason": "Tested in staging, looks good"
}
```

**Response:**
```json
{
  "status": "approved",
  "deployment_id": "dep_123",
  "sync_track_id": "sync_xyz",
  "approved_by": "john@example.com",
  "approved_at": "2026-03-18T10:00:00Z"
}
```

### 7. Enable Auto-Sync

**Endpoint:** `POST /api/v1/deployments/{deploymentId}/git/auto-sync`

Enable automatic synchronization.

**Body:**
```json
{
  "enabled": true,
  "require_approval": false,
  "polling_interval": 3600,
  "max_concurrent": 1,
  "only_prod_ready": false,
  "allowed_branches": ["main", "develop"],
  "ignore_paths": ["docs/", "*.md"],
  "required_approvers": ["tech-lead@example.com"]
}
```

**Response:**
```json
{
  "status": "enabled",
  "deployment_id": "dep_123",
  "config_id": "config_abc"
}
```

### 8. Get Auto-Sync Config

**Endpoint:** `GET /api/v1/deployments/{deploymentId}/git/auto-sync`

Get auto-sync configuration.

**Response:**
```json
{
  "id": "config_abc",
  "deployment_id": "dep_123",
  "enabled": true,
  "require_approval": false,
  "polling_interval": 3600,
  "max_concurrent": 1
}
```

### 9. Update Auto-Sync Config

**Endpoint:** `PATCH /api/v1/deployments/{deploymentId}/git/auto-sync`

Update auto-sync configuration.

**Body:**
```json
{
  "require_approval": true,
  "polling_interval": 1800,
  "allowed_branches": ["main"]
}
```

### 10. Disable Auto-Sync

**Endpoint:** `DELETE /api/v1/deployments/{deploymentId}/git/auto-sync`

Disable automatic synchronization.

**Response:**
```json
{
  "status": "disabled",
  "deployment_id": "dep_123"
}
```

## Synchronization Workflow

### Manual Sync Flow

```
┌─ User Requests Sync ─┐
│                      │
▼                      │
Check for Approval     │
│                      │
├─ Not Required ▶──────┤
│                      │
└─ Required ──────────►│
  ├─ Pending           │
  │  (await approval)  │
  │                    │
  └─ Approved ────────►│
                       │
                       ▼
                 Clone Repository
                       │
                       ▼
            Fetch Latest Commit
                       │
                       ▼
           Generate Diff & Changes
                       │
                       ▼
         Update Deployment Metadata
                       │
                       ▼
              Deploy New Version
                       │
                       ▼
           Run Post-Deployment Tests
                       │
                  ┌────┴────┐
                  │          │
              Success     Failed
                  │          │
                  ▼          ▼
            Mark Synced   Mark Failed
```

### Auto-Sync Flow

```
┌─ Polling Interval Elapsed ─┐
│                            │
▼                            │
Check Remote for Updates     │
│                            │
├─ No Changes ──────────────►│
│                            │
└─ Changes Detected ─────────┘
  │
  ▼
Parse Diff
  │
  ├─ Check Ignore Paths
  │  ├─ All Ignored ──────► Exit (no changes)
  │  └─ Some Changed ──►┐
  │                     │
  ├─ Check Allowed Branches
  │  ├─ Not Allowed ───► Exit
  │  └─ Allowed ───────┤
  │                    │
  ├─ Check Requirements
  │  ├─ Prod Ready Req. & Not Tagged ► Exit
  │  └─ Ready ─────────┤
  │                    │
  ▼                    │
Require Approval?      │
  │                    │
  ├─ No ──────────────►│
  │                    │
  └─ Yes ────────────►│ Notify Approvers
                       │ (pause for approval)
                       │
                       ▼
                  Auto-Sync Deploy
```

## Implementation Guide

### Step 1: Initialize Sync Tracking

When creating a deployment, initialize git sync tracking:

```go
// In deployment creation handler
deployment := &models.Deployment{...}
project := &models.Project{GitRepo: "https://github.com/org/repo"}

track, err := gitSyncService.InitializeSyncTracking(deployment, project)
if err != nil {
    // Handle error
}
```

### Step 2: Check for Updates Periodically

Implement a background worker to check for updates:

```go
// In a scheduled job (cron or timer)
func CheckForUpdatesWorker() {
    tracks, err := gitSyncRepo.GetOutOfSyncDeployments(100)
    if err != nil {
        log.Error().Err(err).Msg("failed to get out-of-sync tracks")
        return
    }

    for _, track := range tracks {
        if err := gitSyncService.CheckForUpdates(&track); err != nil {
            log.Error().Err(err).Msg("failed to check for updates")
            continue
        }

        // Notify about status changes
        if track.SyncStatus == models.GitSyncOutOfSync {
            notificationService.NotifyOutOfSync(track)
        }
    }
}
```

### Step 3: Enable Auto-Sync

Allow users to enable automatic synchronization:

```go
config := &models.GitAutoSyncConfig{
    Enabled:         true,
    RequireApproval: true,
    PollingInterval: 3600,
    AllowedBranches: `["main", "develop"]`,
}

err := gitSyncService.EnableAutoSync(deploymentID, config)
```

### Step 4: Handle Manual Sync

Expose manual sync endpoints:

```go
// API handler for manual sync
func (h *GitSyncHandler) SyncDeployment(c *gin.Context) {
    var req models.SyncRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    userID := c.GetString("user_id")
    if err := h.gitSyncService.SyncDeployment(req.DeploymentID, userID, req.Force); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "syncing"})
}
```

## Configuration Examples

### Example 1: Auto-Sync Main Branch Only

```yaml
auto_sync:
  enabled: true
  require_approval: false
  polling_interval: 1800  # 30 minutes
  allowed_branches:
    - main
  ignore_paths:
    - docs/
    - tests/
    - "*.md"
```

### Example 2: Production with Approval

```yaml
auto_sync:
  enabled: true
  require_approval: true
  polling_interval: 3600  # 1 hour
  required_approvers:
    - tech-lead@example.com
    - devops@example.com
  allowed_branches:
    - main
  only_prod_ready: true  # Requires git tags
```

### Example 3: Staging with Auto-Deploy

```yaml
auto_sync:
  enabled: true
  require_approval: false
  polling_interval: 900  # 15 minutes
  allowed_branches:
    - develop
    - staging
  ignore_paths: []  # Deploy all changes
  max_concurrent: 2
```

## Monitoring & Alerts

### Key Metrics

- **Sync Status Distribution**: Track deployment sync states
- **Sync Duration**: Average time to complete syncs
- **Sync Success Rate**: Percentage of successful syncs
- **Time Out of Sync**: How long deployments lag behind
- **Approval Timing**: Average time to approval

### Alert Conditions

```yaml
alerts:
  sync_failed:
    condition: sync_status == "failed"
    threshold: 1
    duration: 5m
    
  out_of_sync_prolonged:
    condition: sync_status == "out_of_sync" AND time_since_last_sync > 24h
    threshold: 1
    
  approval_timeout:
    condition: sync_status == "pending" AND pending_duration > 24h
    threshold: 1
```

## Best Practices

✅ **DO:**
- Start with manual sync, then enable auto-sync
- Require approval for production deployments
- Monitor sync history and error logs
- Test changes in staging first
- Use branch-specific auto-sync policies
- Keep polling intervals reasonable (3600s = 1 hour)
- Document sync requirements for your team

❌ **DON'T:**
- Use very short polling intervals (< 300s)
- Auto-sync without monitoring
- Ignore approval requirements for production
- Sync without running tests
- Deploy to production without staging first
- Keep very detailed old history (archive it)

## Troubleshooting

### Issue: Sync Status Stuck in "Syncing"

**Solution:**
1. Check deployment service logs
2. Verify git repository is accessible
3. Check for authentication issues
4. Review deployment constraints

### Issue: Changes Not Detected

**Solution:**
1. Verify git_repo field is set on project
2. Check branch name is correct
3. Confirm git credentials/tokens have repo access
4. Check polling interval isn't too long

### Issue: Approval Keeps Pending

**Solution:**
1. Verify all required approvers are set
2. Check approver email addresses
3. Review notification settings
4. Ensure approval endpoint is correctly integrated

## Integration Points

- **GitHub Integration**: `GITHUB_INTEGRATION_WORKFLOW.md`
- **Deployment API**: `API_REFERENCE.md`
- **Notifications**: See `notification_service.go`
- **WebHooks**: See `webhook_service.go`

---

**Last Updated:** March 18, 2026
**Version:** 1.0.0
