# Feature: Git Change Tracking & Synchronization

Complete GitOps-style git change tracking and deployment synchronization system for Pushpaka.

## Overview

Track changes in your git repositories and keep your deployments automatically synchronized with the latest code. Similar to ArgoCD but optimized for containerized applications.

## Features

### 🔄 Automatic Change Detection
- **Real-time monitoring** of git repositories
- **File-level change tracking** (modified, added, deleted, renamed)
- **Detailed diffs** with addition/deletion counts
- **Multi-provider support** (GitHub, GitLab, self-hosted)

### 🎯 Multiple Synchronization Modes

**Manual Sync**
```bash
curl -X POST http://localhost:8080/api/v1/deployments/{id}/git/sync
```
Manually sync deployment to latest code with optional approval.

**Scheduled Auto-Sync**
```yaml
auto_sync:
  enabled: true
  polling_interval: 3600
  require_approval: false
```
Periodically check for updates and sync automatically.

**Webhook Integration**
Instant notification on git pushes (GitHub, GitLab webhooks).

### ✅ Approval Workflow
- Optional approval gates for controlled deployments
- Multiple approver support
- Approval reason tracking
- Timeout handling
- Complete audit trail

### 📊 Change Visibility
- Detailed file-level diffs
- Commit metadata (author, message, timestamp)
- Change summaries and statistics
- Before/after content comparison

### 📝 Complete History
- Full audit trail of every sync
- Sync type tracking (manual/auto/webhook)
- Error recording and debugging
- Rollback tracking and recovery

## Quick Start

### 1. Enable Git Tracking on Project

Set the git repository on your project:

```bash
curl -X PATCH http://localhost:8080/api/v1/projects/proj_123 \
  -H "Content-Type: application/json" \
  -d '{
    "git_repo": "https://github.com/org/my-repo.git"
  }'
```

### 2. Check for Updates

```bash
curl http://localhost:8080/api/v1/deployments/dep_123/git/check-updates

Response:
{
  "sync_status": "out_of_sync",
  "latest_sha": "abc123def",
  "current_sha": "old123sha",
  "total_changes": 15,
  "total_additions": 245,
  "total_deletions": 89
}
```

### 3. Manual Sync

```bash
curl -X POST http://localhost:8080/api/v1/deployments/dep_123/git/sync \
  -H "Content-Type: application/json" \
  -d '{"force": false}'

Response:
{
  "status": "syncing",
  "deployment_id": "dep_123"
}
```

### 4. Enable Auto-Sync

```bash
curl -X POST http://localhost:8080/api/v1/deployments/dep_123/git/auto-sync \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "require_approval": false,
    "polling_interval": 3600,
    "allowed_branches": ["main", "develop"]
  }'
```

### 5. View Sync History

```bash
curl http://localhost:8080/api/v1/deployments/dep_123/git/history?limit=20

Response:
{
  "total": 42,
  "history": [
    {
      "id": "sync_001",
      "from_commit_sha": "old123",
      "to_commit_sha": "new456",
      "sync_type": "manual",
      "status": "success",
      "total_changes": 15,
      "duration": 45,
      "created_at": "2026-03-18T10:00:00Z"
    },
    ...
  ]
}
```

## API Endpoints

### Monitoring & Status
- `GET /api/v1/deployments/{id}/git/status` - Current sync status
- `GET /api/v1/deployments/{id}/git/check-updates` - Check for new commits
- `GET /api/v1/deployments/{id}/git/changes` - View detailed file changes
- `GET /api/v1/deployments/{id}/git/history` - Sync history and records

### Synchronization
- `POST /api/v1/deployments/{id}/git/sync` - Manual sync to latest code
- `POST /api/v1/deployments/{id}/git/approve-sync` - Approve pending sync

### Auto-Sync Configuration
- `POST /api/v1/deployments/{id}/git/auto-sync` - Enable/configure auto-sync
- `GET /api/v1/deployments/{id}/git/auto-sync` - Get current config
- `PATCH /api/v1/deployments/{id}/git/auto-sync` - Update config
- `DELETE /api/v1/deployments/{id}/git/auto-sync` - Disable auto-sync

## Configuration Examples

### Minimal (Manual Sync Only)
```yaml
git_sync:
  enabled: true
  auto_sync:
    enabled: false
```

### Production (Approval Required)
```yaml
git_sync:
  auto_sync:
    enabled: true
    require_approval: true
    polling_interval: 3600
    allowed_branches:
      - main
      - production
    ignore_paths:
      - docs/
      - "*.md"
    required_approvers:
      - tech-lead@example.com
```

### Staging (Auto-Deploy)
```yaml
git_sync:
  auto_sync:
    enabled: true
    require_approval: false
    polling_interval: 900
    allowed_branches:
      - develop
      - staging
```

## Data Models

### GitSyncTrack
Tracks synchronization status between deployment and git repository:
- Current and latest commit SHAs
- Sync status (synced/out_of_sync/syncing/pending/failed)
- Change statistics
- Approval information

### GitChange
Records individual file changes:
- File path and change type
- Addition/deletion counts
- Optional content diffs

### GitAutoSyncConfig
Auto-sync configuration per deployment:
- Enable/disable toggle
- Approval requirement
- Polling interval
- Branch filters
- Path ignoring
- Required approvers

### DeploymentSyncHistory
Complete audit trail:
- From/to commits
- Sync type and status
- Duration and error info
- Rollback tracking

## Background Worker

The **GitSyncWorker** runs continuously:
- Polls every 10 seconds for pending operations
- Checks auto-sync configurations every 1 hour (configurable)
- Triggers syncs based on schedule and configuration
- Sends notifications on status changes
- Cleans up stale approval requests

## Comparison with ArgoCD

| Feature | Pushpaka | ArgoCD |
|---------|----------|--------|
| Git Tracking | ✅ Yes | ✅ Yes |
| Auto-Sync | ✅ Scheduled | ✅ Continuous |
| Manual Sync | ✅ Yes | ✅ Yes |
| Approval Gates | ✅ Built-in | ⚠️ Limited |
| K8s Native | ❌ No | ✅ Yes |
| Docker Support | ✅ Direct | ⚠️ Via manifests |
| Change Tracking | ✅ Detailed | ✅ Detailed |
| History | ✅ Complete | ✅ Complete |

See [PUSHPAKA_VS_ARGOCD.md](PUSHPAKA_VS_ARGOCD.md) for detailed comparison.

## Documentation

Comprehensive documentation is available:

- **[GIT_SYNC_IMPLEMENTATION.md](GIT_SYNC_IMPLEMENTATION.md)** - Complete technical guide
- **[GIT_SYNC_INTEGRATION_GUIDE.md](GIT_SYNC_INTEGRATION_GUIDE.md)** - Setup and integration
- **[GIT_SYNC_IMPLEMENTATION_SUMMARY.md](GIT_SYNC_IMPLEMENTATION_SUMMARY.md)** - Overview
- **[PUSHPAKA_VS_ARGOCD.md](PUSHPAKA_VS_ARGOCD.md)** - ArgoCD comparison
- **[API_REFERENCE.md](API_REFERENCE.md)** - Full API documentation

## Monitoring

### Key Metrics
- `git_sync_status` - Current sync status per deployment
- `git_sync_duration_seconds` - Time to complete syncs
- `git_sync_changes_total` - Files changed
- `git_sync_failures` - Failed sync attempts
- `git_sync_pending_approvals` - Awaiting approval

### Alerts
- Out of sync for >24 hours
- Approval pending for >24 hours
- Sync failures (consecutive)
- High change volume

## Use Cases

### Development/Staging
Auto-deploy every commit to staging with quick feedback:
```yaml
auto_sync:
  enabled: true
  polling_interval: 300
  require_approval: false
```

### Production
Controlled deployments with approval gates:
```yaml
auto_sync:
  enabled: true
  polling_interval: 3600
  require_approval: true
```

### Manual Control
Users decide when to sync:
```yaml
auto_sync:
  enabled: false
```

## Workflow Example

```
1. Developer pushes to GitHub
        ↓
2. GitHub webhook notifies Pushpaka
        ↓
3. GitSync Service detects new commit
        ↓
4. Sys marks deployment as "out_of_sync"
        ↓
5. Dashboard shows change notification
        ↓
6. Admin clicks "Sync Now" or auto-sync triggers
        ↓
7. Deploymentdates with new code
        ↓
8. Post-deploy health checks run
        ↓
9. Status updated to "synced"
        ↓
10. Sync recorded in history with duration/stats
```

## Integration

### With GitHub
- Automatic webhook integration
- Commit metadata fetching
- File change tracking
- Branch management

### With Deployments
- Automatic commit SHA tracking
- Version management
- Rollback support

### With Notifications
- Out-of-sync alerts
- Approval requests
- Sync status updates
- Failure notifications

## Best Practices

✅ **DO**
- Start with manual sync
- Require approval for production
- Monitor sync history
- Test in staging first
- Use branch-specific policies
- Set reasonable polling intervals

❌ **DON'T**
- Use very short polling intervals (<300s)
- Auto-sync without monitoring
- Skip approval for production
- Deploy without testing
- Store credentials in config

## Troubleshooting

### Sync Status Not Updating
1. Check if GitSyncWorker is running
2. Verify git_repo field is set
3. Ensure git repo is accessible

### Auto-Sync Not Triggering
1. Verify auto_sync is enabled
2. Check polling interval
3. Confirm branch is allowed

### Changes Not Showing
1. Verify git credentials
2. Check branch exists
3. Confirm commit is reachable

See [GIT_SYNC_IMPLEMENTATION.md](GIT_SYNC_IMPLEMENTATION.md#troubleshooting) for detailed troubleshooting.

## File Structure

```
backend/
├── internal/
│   ├── models/
│   │   └── git_sync.go           # Data models
│   ├── repositories/
│   │   └── git_sync_repo.go      # Database operations
│   ├── services/
│   │   └── git_sync_service.go   # Business logic
│   └── handlers/
│       └── api/
│           └── git_sync_handler.go # HTTP endpoints
└── queue/
    └── git_sync_worker.go        # Background worker

migrations/
└── 002_git_sync_tables.sql       # Database schema
```

## Getting Started

1. **Run migrations** - Create git sync tables
2. **Start worker** - Enable background polling
3. **Register routes** - Add API endpoints
4. **Configure projects** - Set git_repo field
5. **Test it** - Manual sync first, then auto-sync

See [GIT_SYNC_INTEGRATION_GUIDE.md](GIT_SYNC_INTEGRATION_GUIDE.md) for step-by-step setup.

## Support

- Check the [documentation](GIT_SYNC_IMPLEMENTATION.md)
- Review [API reference](API_REFERENCE.md)
- See [troubleshooting guide](GIT_SYNC_INTEGRATION_GUIDE.md#troubleshooting)
- Compare with [ArgoCD](PUSHPAKA_VS_ARGOCD.md)

## Status

✅ **Complete & Ready for Integration**
- 6 code files (models, repo, service, handler, worker)
- 5 documentation files
- ~2500 lines of backend code
- ~3000 lines of documentation
- Database schema and migrations

---

**Feature**: Git Change Tracking & Synchronization
**Release**: March 18, 2026
**Version**: 1.0.0
