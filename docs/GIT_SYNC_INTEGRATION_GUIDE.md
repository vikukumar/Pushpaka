# Git Sync Integration Guide

Complete guide for integrating git change tracking and synchronization into your Pushpaka deployment pipeline.

## Quick Start

### 1. Enable Git Repository Tracking

```go
// Add git_repo field to project
project := &models.Project{
    Name:    "my-app",
    GitRepo: "https://github.com/org/my-repo.git",
    Branch:  "main",
}

// Database will automatically track changes
```

### 2. Initialize Sync on Deployment

```go
// When creating a deployment
deployment := &models.Deployment{
    ProjectID: project.ID,
    CommitSHA: commitSHA,
    Branch:    "main",
}

// Initialize sync tracking
track, err := gitSyncService.InitializeSyncTracking(deployment, project)
if err != nil {
    return err
}

// Now git changes are automatically tracked
```

### 3. Check for Updates

```bash
# Check if new commits are available
curl GET http://localhost:8080/api/v1/deployments/{deploymentId}/git/check-updates

# Response shows:
# - sync_status: "out_of_sync" if new commits
# - latest_sha: newest commit hash
# - total_changes: files changed
```

### 4. Sync to Latest Code

```bash
# Manual sync to latest code
curl -X POST http://localhost:8080/api/v1/deployments/{deploymentId}/git/sync \
  -H "Content-Type: application/json" \
  -d '{
    "deployment_id": "dep_123",
    "force": false,
    "reason": "Deploying new feature"
  }'
```

### 5. Enable Auto-Sync

```bash
# Enable automatic synchronization
curl -X POST http://localhost:8080/api/v1/deployments/{deploymentId}/git/auto-sync \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "require_approval": false,
    "polling_interval": 3600,
    "allowed_branches": ["main", "develop"]
  }'
```

## Implementation Checklist

### Phase 1: Setup (Required)

- [ ] Add `git_repo` field to projects table
- [ ] Create git sync tracking tables (run migrations)
- [ ] Add `GitSyncRepository` to dependency injection
- [ ] Register `GitSyncService` in service layer
- [ ] Register `GitSyncHandler` in route handlers
- [ ] Start `GitSyncWorker` in main application

### Phase 2: Core Features (Recommended)

- [ ] Implement manual sync endpoint
- [ ] Implement auto-sync configuration
- [ ] Add git change tracking
- [ ] Add sync history recording
- [ ] Implement approval workflow
- [ ] Add notifications on sync events

### Phase 3: Advanced Features (Optional)

- [ ] Implement webhook integration
- [ ] Add sync metrics and monitoring
- [ ] Implement conditional syncs (ignore paths, branch filters)
- [ ] Add sync history archival
- [ ] Implement rollback on sync failure

## Configuration Guide

### Minimal Configuration (Dev/Testing)

```yaml
git_sync:
  enabled: true
  auto_sync:
    enabled: false       # Manual sync only
```

### Recommended Configuration (Production)

```yaml
git_sync:
  enabled: true
  auto_sync:
    enabled: true
    require_approval: true
    polling_interval: 3600      # Check every hour
    max_concurrent: 1            # One sync at a time
    allowed_branches:
      - main
      - production
    ignore_paths:
      - docs/
      - tests/
      - "*.md"
    required_approvers:
      - tech-lead@example.com
```

### Aggressive Configuration (Staging)

```yaml
git_sync:
  enabled: true
  auto_sync:
    enabled: true
    require_approval: false      # Auto-deploy
    polling_interval: 900        # Check every 15 minutes
    max_concurrent: 2
    allowed_branches:
      - develop
      - staging
    ignore_paths: []             # Deploy all changes
```

## Database Schema

### Tables Created

1. **git_sync_tracks** - Main sync status tracking
2. **git_changes** - Individual file changes
3. **git_auto_sync_config** - Auto-sync settings
4. **deployment_sync_history** - Historical records

### Key Fields

```sql
-- Current status
current_commit_sha      -- Currently deployed commit
latest_commit_sha       -- Latest available commit
sync_status            -- synced|out_of_sync|syncing|pending|failed

-- Change tracking
total_changes          -- Files changed
total_additions        -- Lines added
total_deletions        -- Lines deleted

-- Approval workflow
sync_approval_required -- Needs approval?
sync_approved_by       -- Who approved
sync_approved_at       -- When approved

-- Timestamps
last_sync_attempt_at        -- Last attempt
last_successful_sync_at     -- Last success
notification_sent_at        -- Last notification
```

## API Endpoints Reference

### Status & Monitoring

```
GET    /api/v1/deployments/{id}/git/status           # Current sync status
GET    /api/v1/deployments/{id}/git/check-updates    # Check for new commits
GET    /api/v1/deployments/{id}/git/changes          # View detailed changes
GET    /api/v1/deployments/{id}/git/history          # Sync history
```

### Synchronization

```
POST   /api/v1/deployments/{id}/git/sync             # Manual sync
POST   /api/v1/deployments/{id}/git/approve-sync     # Approve pending sync
```

### Auto-Sync Configuration

```
POST   /api/v1/deployments/{id}/git/auto-sync        # Enable auto-sync
GET    /api/v1/deployments/{id}/git/auto-sync        # Get config
PATCH  /api/v1/deployments/{id}/git/auto-sync        # Update config
DELETE /api/v1/deployments/{id}/git/auto-sync        # Disable auto-sync
```

## Workflow Examples

### Example 1: Manual Sync Workflow

```
Developer pushes code
    ↓
GitHub webhook notifies Pushpaka
    ↓
GitSync Service detects new commit
    ↓
System marks deployment as "out_of_sync"
    ↓
Dashboard notification sent
    ↓
Admin clicks "Sync Now"
    ↓
Deployment updated with new code
    ↓
Health checks run
    ↓
Status updated to "synced"
```

### Example 2: Auto-Sync Workflow

```
GitSync Worker polls every 1 hour
    ↓
Detects new commit in "main" branch
    ↓
Checks auto-sync config
    ├─ Requires approval? → Notify approvers → Wait
    └─ No approval? → Proceed
    ↓
Clones repository and generates diff
    ↓
Updates deployment commit SHA
    ↓
Triggers deployment rebuild
    ↓
Runs tests
    ↓
Updates sync status to "synced"
    ↓
Records in sync history
```

### Example 3: Approval Workflow

```
Auto-sync triggered but approval required
    ↓
System creates "pending" sync request
    ↓
Notifications sent to required approvers
    ↓
Approver reviews changes in dashboard
    ↓
    ├─ Approves → Sync proceeds immediately
    └─ Rejects → Status back to "out_of_sync"
    ↓
Record approval in history
```

## Monitoring & Observability

### Key Metrics

```
Metric Name                    Description
────────────────────────────── ──────────────────────────
git_sync_status               Current status (synced/out_of_sync)
git_sync_duration_seconds     Time taken for last sync
git_sync_changes_total        Files changed in last sync
git_sync_failures_total       Failed sync attempts
git_sync_approvals_pending    Syncs waiting for approval
git_sync_last_check_seconds   Seconds since last check
```

### Dashboard Queries

```
# Deployments out of sync
SELECT count(*) FROM git_sync_tracks 
WHERE sync_status = 'out_of_sync'

# Average sync time
SELECT AVG(duration) FROM deployment_sync_history 
WHERE status = 'success'

# Most frequently synced deployments
SELECT deployment_id, count(*) as sync_count
FROM deployment_sync_history
GROUP BY deployment_id
ORDER BY sync_count DESC LIMIT 10
```

### Alert Conditions

```
Alert Name                         Condition
─────────────────────────────────  ──────────────────────────
out_of_sync_prolonged              sync_status = 'out_of_sync' 
                                   AND time > 24 hours

approval_timeout                   sync_status = 'pending'
                                   AND pending_time > 24 hours

sync_failure_consecutive           failures > 3 in last hour

high_change_volume                 total_changes > 100
                                   AND status != 'testing'
```

## Troubleshooting

### Issue: Sync Status Not Updating

**Steps:**
1. Check GitSyncWorker is running: `GET /api/v1/health`
2. Verify git_repo field is set on project
3. Check git repository is accessible
4. Review error logs in sync history

### Issue: Auto-Sync Not Triggering

**Steps:**
1. Verify auto-sync is enabled: `GET /api/v1/deployments/{id}/git/auto-sync`
2. Check polling interval has passed
3. Verify branch is in allowed_branches list
4. Ensure no pending approvals blocking sync

### Issue: Changes Not Showing

**Steps:**
1. Verify git checkout is current
2. Check git credentials have repo access
3. Confirm branch exists in remote
4. Review git diff generation logs

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "repository not found" | Invalid git_repo URL | Update project git_repo field |
| "authentication failed" | No access to repo | Add SSH key / personal access token |
| "branch not found" | Branch doesn't exist | Verify branch name |
| "sync already in progress" | Lock timeout | Wait or force sync |

## Performance Tuning

### Polling Interval

```
Frequency | Interval | Use Case
────────  ──────    ──────────────────
Every 15m | 900s    | Staging/rapid iteration
Every 30m | 1800s   | Production critical
Every 1h  | 3600s   | Low-frequency updates
```

### Batch Processing

```
# Process multiple pending syncs
GitSyncWorker processes batches of:
- Up to 100 out-of-sync deployments per poll
- Up to 10 auto-sync configs per poll
- Prevents resource exhaustion
```

### Cleanup

```
Cleanup Task          | Frequency  | Purpose
──────────────────    ──────────   ────────────────────
Stale approval reqs   | 5 minutes  | Remove >24hr pending
History archival      | Daily      | Move old records
Failed sync retry     | 1 hour     | Retry failed syncs
Cache cleanup         | Hourly     | Remove temp clones
```

## Security Considerations

### Access Control

```go
// Require authentication for sync operations
// Only project admins can:
// - Enable/disable auto-sync
// - Force sync without approval
// - Update auto-sync configuration

// Required approvers must be explicitly configured
// Approval is recorded and audited
```

### Git Credentials

```
DON'T: Store credentials in deployment config
DO:    Use SSH keys or PATs in secure storage
DO:    Rotate credentials regularly
DO:    Limit credentials to minimum required repos
```

### Audit Trail

```
All syncs are recorded:
- Who triggered (user_id or system)
- When (timestamps)
- What changed (commit diffs)
- Why (approval reason)
- Result (success/failure)
```

## Integration with CI/CD

### GitOps Flow

```
1. Developer pushes to GitHub
        ↓
2. GitHub Actions run tests
        ↓
3. Tests pass/fail
        ↓
4. Webhook notifies Pushpaka
        ↓
5. Pushpaka detects new commit
        ↓
6. Auto-sync (if enabled) or Pending approval
        ↓
7. Deployment runs
        ↓
8. Post-deployment tests
        ↓
9. Health checks
        ↓
10. Status updated in dashboard
```

### Webhook Integration

```
POST /webhooks/github
{
  "event": "push",
  "repository": "org/repo",
  "branch": "main",
  "commits": ["abc123..."]
}

→ Pushpaka updates git state immediately
```

## Migration from Manual Deployments

### Step 1: Enable Git Tracking
```
Update projects table to add git_repo field
```

### Step 2: Record Historical Commits
```go
// Map existing deployments to commits
for each deployment {
    record commit_sha from deployment metadata
    initialize sync tracking
}
```

### Step 3: Enable Auto-Sync Gradually
```
1. Start with manual sync only
2. Test with staging deployments
3. Enable auto-sync for non-critical services
4. Gradually roll out to production
```

## Support & Resources

- **Documentation**: See `GIT_SYNC_IMPLEMENTATION.md`
- **API Reference**: See `API_REFERENCE.md`
- **ArgoCD Comparison**: See `PUSHPAKA_VS_ARGOCD.md`
- **GitHub Integration**: See `GITHUB_INTEGRATION_WORKFLOW.md`

---

**Last Updated:** March 18, 2026
**Version:** 1.0.0
