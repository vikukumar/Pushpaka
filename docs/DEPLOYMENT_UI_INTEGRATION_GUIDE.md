# Deployment Management UI Integration Guide

## Overview

This guide covers the frontend components and integration patterns for the Deployment Management System. The UI should display deployment status, allow user actions, and provide configuration options.

## Key UI Components

### 1. Deployment Status Dashboard

**Location**: Project Dashboard / Deployments Tab

**Purpose**: Show current state of all deployments for a project

**Display Information**:
```
┌─────────────────────────────────────────────────┐
│ PROJECT DEPLOYMENTS                             │
├─────────────────────────────────────────────────┤
│                                                  │
│ PRODUCTION (Main)                               │
│ ├─ Status: RUNNING ✓                           │
│ ├─ Version: 1.1.0                              │
│ ├─ Uptime: 7d 23h 12m                          │
│ ├─ Health: Healthy ✓                           │
│ ├─ Actions: [Stop] [Restart] [View Backups]    │
│ └─ Code: 42 files, hash: abc123...             │
│                                                  │
│ TESTING (Staging)                               │
│ ├─ Status: RUNNING ✓                           │
│ ├─ Version: 1.2.0-beta                         │
│ ├─ Uptime: 2h 15m                              │
│ ├─ Health: Healthy ✓                           │
│ ├─ Actions: [Promote] [Stop] [View Code]       │
│ └─ Code: 42 files, hash: def456...             │
│                                                  │
│ BACKUPS (3 available)                            │
│ ├─ Backup 1: v1.0.0 (5 days old, 2GB)          │
│ ├─ Backup 2: v0.9.0 (12 days old, 2GB)         │
│ └─ Backup 3: v0.8.0 (22 days old, 2GB)         │
│                                                  │
└─────────────────────────────────────────────────┘
```

**API Calls**:
- `GET /api/v1/projects/{projectId}/deployments` - List all deployments
- `GET /api/v1/deployments/{id}/status` - Get detailed status
- `GET /api/v1/deployments/{id}/backups` - List backups

### 2. Deployment Action Panel

**Location**: Per-deployment control area

**State Machine**:
```
        Created
           ↓
    [Deploy] → Preparing → Running → [Promote/Stop/Restart]
                  ↑                          ↓
                  └──────────────────────────┘
       [Retry]        ERROR
```

**Action Buttons**:
```
For Main Deployment:
- [Stop] - Stop deployment (before promoting new version)
- [Restart] - Restart if crashed
- [Health Check] - Verify status
- [View Logs] - See application logs

For Testing Deployment:
- [Promote to Main] - Promote to production
- [Stop] - Stop testing instance
- [Restart] - Restart testing instance
- [Sync Code] - Pull latest from git

For Backup:
- [Restore] - Restore from this backup
- [Delete] - Delete backup (if > MaxBackups)
```

**API Calls**:
```
Start:     POST /api/v1/deployments/{id}/actions/start
Stop:      POST /api/v1/deployments/{id}/actions/stop
Restart:   POST /api/v1/deployments/{id}/actions/restart
Retry:     POST /api/v1/deployments/{id}/actions/retry
Rollback:  POST /api/v1/deployments/{id}/actions/rollback (with backupId)
Sync:      POST /api/v1/deployments/{id}/actions/sync
```

**Response Handling**:
```javascript
// All actions return 202 Accepted with actionId
const response = await fetch(
  `/api/v1/deployments/${deploymentId}/actions/restart`,
  { method: 'POST' }
);

const { actionId } = await response.json();

// Poll for completion
const checkStatus = async () => {
  const action = await fetch(`/api/v1/actions/${actionId}`);
  const { status, result } = await action.json();
  
  if (status === 'success') {
    showNotification('Deployment restarted', result);
  } else if (status === 'failed') {
    showError('Deployment failed', result);
  } else if (status === 'executing') {
    setTimeout(checkStatus, 2000);  // Poll every 2s
  }
};

checkStatus();
```

### 3. Promotion Flow

**Location**: Testing deployment actions

**User Workflow**:
```
1. User clicks [Promote to Main]
   ↓
2. Confirmation dialog appears
   - Show main version that will be demoted
   - Show testing version that will be promoted
   - Ask for confirmation
   ↓
3. User confirms
   ↓
4. API call: POST /api/v1/deployments/{testingId}/actions/promote
   (Server creates backup of main, stops main, promotes testing)
   ↓
5. Poll status until complete
   ↓
6. Show success notification
   - "Deployment promoted successfully"
   - Show new main version
   - Show previous main in backups
```

**Dialog Example**:
```
┌──────────────────────────────────────────┐
│ PROMOTE DEPLOYMENT TO MAIN               │
├──────────────────────────────────────────┤
│                                           │
│ Current Main Deployment:                 │
│  Version: 1.1.0                          │
│  Uptime: 7d 23h 12m                      │
│  Status: Running ✓                       │
│  Action: Will be moved to backups        │
│                                           │
│ New Main Deployment (from Testing):      │
│ Version: 1.2.0-beta                      │
│ Tests: All passed ✓                       │
│ Status: Running ✓                         │
│ Action: Will become main production      │
│                                           │
│ Shutdown Timeout: 30 seconds              │
│ Create Backup: Yes ✓                      │
│                                           │
│ [Cancel] [Promote]                        │
└──────────────────────────────────────────┘
```

### 4. Backup Management

**Location**: Deployments tab, Backups section

**Display**:
```
┌────────────────────────────────────────────────┐
│ BACKUPS (3 of 3)                              │
├────────────────────────────────────────────────┤
│                                                 │
│ Backup 1                                        │
│ ├─ Version: v1.0.0                            │
│ ├─ Reason: pre_promotion_rollback             │
│ ├─ Size: 2.1 GB                               │
│ ├─ Date: 2024-01-14 14:30:22                  │
│ ├─ Restored: No                               │
│ └─ Actions: [Restore] [Download] [Delete]     │
│                                                 │
│ Backup 2                                        │
│ ├─ Version: v0.9.0                            │
│ ├─ Reason: health_check_recovery              │
│ ├─ Size: 2.0 GB                               │
│ ├─ Date: 2024-01-12 09:15:11                  │
│ ├─ Restored: Yes (2024-01-13 11:22:30)        │
│ └─ Actions: [View] [Download] [Delete]        │
│                                                 │
│ Backup 3                                        │
│ ├─ Version: v0.8.0                            │
│ ├─ Reason: deployment_error_recovery          │
│ ├─ Size: 1.9 GB                               │
│ ├─ Date: 2024-01-08 16:45:00                  │
│ ├─ Restored: No                               │
│ └─ Actions: [Restore] [Download]              │
│                                                 │
│ Max Backups Setting: 3 [Edit]                  │
│                                                 │
└────────────────────────────────────────────────┘
```

**Restore Dialog**:
```
┌──────────────────────────────────────────┐
│ RESTORE BACKUP                           │
├──────────────────────────────────────────┤
│                                           │
│ Restoring from: Backup - v1.0.0          │
│ Date Created: 2024-01-14 14:30:22        │
│ Size: 2.1 GB                             │
│                                           │
│ Current Deployment:                      │
│ Version: 1.1.0                           │
│ Status: Running ✓                         │
│ Action: Will be backed up before restore │
│                                           │
│ WARNING:                                  │
│ This will stop current deployment and    │
│ restore previous version. Ensure all     │
│ dependent services are notified.         │
│                                           │
│ [Cancel] [Restore]                       │
└──────────────────────────────────────────┘
```

**API Calls**:
- `GET /api/v1/deployments/{id}/backups` - List backups
- `POST /api/v1/deployments/{id}/backups/{backupId}/restore` - Restore
- `DELETE /api/v1/deployments/{id}/backups/{backupId}` - Delete backup

### 5. Configuration Panel

**Location**: Project Settings → Deployment Configuration

**Display**:
```
┌─────────────────────────────────────────┐
│ DEPLOYMENT CONFIGURATION                │
├─────────────────────────────────────────┤
│                                          │
│ Max Concurrent Deployments              │
│ ├─ Current: 2                           │
│ ├─ Information: Controls how many       │
│ │  deployments can run simultaneously   │
│ │  (usually 1 main + 1 testing)         │
│ └─ [Input: 1-5] [Save]                  │
│                                          │
│ Max Backups to Keep                     │
│ ├─ Current: 3                           │
│ ├─ Information: Number of previous      │
│ │  versions retained for rollback       │
│ ├─ Disk Usage: ~6.5 GB (current)        │
│ └─ [Input: 1-10] [Save]                 │
│                                          │
│ Graceful Shutdown Timeout               │
│ ├─ Current: 30 seconds                  │
│ ├─ Information: Time to allow existing  │
│ │  connections to complete before force │
│ │  stopping                             │
│ └─ [Input: 5-300s] [Save]               │
│                                          │
│ Auto-Sync Updates                       │
│ ├─ Current: Enabled ✓                   │
│ ├─ Information: Automatically update    │
│ │  deployments when git changes detected│
│ └─ [Toggle: On/Off]                     │
│                                          │
│ Health Check Interval                   │
│ ├─ Current: 60 seconds                  │
│ ├─ Information: How often to check      │
│ │  deployment health                    │
│ └─ [Input: 10-300s] [Save]              │
│                                          │
└─────────────────────────────────────────┘
```

**API Calls**:
- `PATCH /api/v1/projects/{projectId}/deployment-limits`
- `GET /api/v1/projects/{projectId}/deployment-config`
- `PATCH /api/v1/projects/{projectId}/deployment-config`

### 6. Statistics & Monitoring

**Location**: Project Dashboard → Deployment Stats

**Display**:
```
┌─────────────────────────────────────────┐
│ DEPLOYMENT STATISTICS                   │
├─────────────────────────────────────────┤
│                                          │
│ Lifetime Statistics                     │
│ ├─ Total Deployments: 15               │
│ ├─ Successful: 14 (93.3%)              │
│ ├─ Failed: 1 (6.7%)                    │
│ └─ Average Deploy Time: 5m 30s          │
│                                          │
│ Current Status                          │
│ ├─ Active Deployments: 2               │
│ │  ├─ Main: Running (7d 23h uptime)    │
│ │  └─ Testing: Running (2h 15m uptime) │
│ └─ Backup Versions: 3 (6.5 GB)         │
│                                          │
│ Performance Metrics                     │
│ ├─ Last Deployment: 5m 22s ago         │
│ ├─ Deployment Time Range: 4m 10s - 8m 5s │
│ ├─ Health Check Frequency: Every 60s   │
│ └─ Unhealthy Restarts (7d): 0          │
│                                          │
│ Recent Activity                         │
│ ├─ 2024-01-15 14:30 - v1.1 deployed   │
│ ├─ 2024-01-14 09:15 - v1.0 promoted   │
│ ├─ 2024-01-13 11:22 - Backup restored │
│ └─ 2024-01-12 16:45 - Auto sync        │
│                                          │
└─────────────────────────────────────────┘
```

**API Calls**:
- `GET /api/v1/projects/{projectId}/deployment-stats`
- `GET /api/v1/deployments/{id}/status`
- `GET /api/v1/actions/{actionId}` (periodic polling)

### 7. Health Status Indicator

**Display Variations**:
```
Healthy               Unhealthy            Unknown
│ Status: ✓ Green    │ Status: ✗ Red       │ Status: ? Gray
│ Response: OK       │ Response: Timeout   │ Response: No data
│ Last Check: 2m ago │ Last Check: 5m ago  │ Last Check: Never
└ Auto-restart: No   └ Restart attempts: 2 └ Waiting for first check
```

**Real-time Updates**:
- Poll `/api/v1/deployments/{id}/status` every 30 seconds
- Update health indicator and restart count
- Show notification if transitions from healthy to unhealthy
- Show notification if auto-restart triggered

## Example React Component

```typescript
// DeploymentStatusCard.tsx
import React, { useEffect, useState } from 'react';

interface DeploymentInstance {
  id: string;
  role: 'main' | 'testing' | 'backup';
  status: 'preparing' | 'running' | 'stopping' | 'stopped';
  healthStatus: 'healthy' | 'unhealthy' | 'unknown';
  version: string;
  uptime: string;
  restart_count: number;
}

export const DeploymentStatusCard: React.FC<{ deploymentId: string }> = ({
  deploymentId,
}) => {
  const [deployment, setDeployment] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [actionInProgress, setActionInProgress] = useState<string | null>(null);

  useEffect(() => {
    const fetchStatus = async () => {
      const response = await fetch(
        `/api/v1/deployments/${deploymentId}/status`
      );
      const data = await response.json();
      setDeployment(data);
      setLoading(false);
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 30000); // Poll every 30s
    return () => clearInterval(interval);
  }, [deploymentId]);

  const handleAction = async (action: string) => {
    setActionInProgress(action);
    const response = await fetch(
      `/api/v1/deployments/${deploymentId}/actions/${action}`,
      { method: 'POST' }
    );
    const { actionId } = await response.json();

    // Poll for completion
    const checkCompletion = async () => {
      const statusResp = await fetch(`/api/v1/actions/${actionId}`);
      const { status, result } = await statusResp.json();

      if (status === 'success' || status === 'failed') {
        setActionInProgress(null);
        // Refetch deployment status
        const updatedResp = await fetch(
          `/api/v1/deployments/${deploymentId}/status`
        );
        const updatedData = await updatedResp.json();
        setDeployment(updatedData);
      } else {
        setTimeout(checkCompletion, 2000);
      }
    };

    checkCompletion();
  };

  if (loading) return <div>Loading...</div>;

  const instance = deployment.instances[0];

  return (
    <div className="deployment-card">
      <h3>{deployment.role?.toUpperCase() || 'DEPLOYMENT'}</h3>

      <div className="status">
        <span className={`badge ${instance.status}`}>
          {instance.status}
        </span>
        <span className={`health ${instance.healthStatus}`}>
          {instance.healthStatus}
        </span>
      </div>

      <div className="info">
        <p>Version: {deployment.version}</p>
        <p>Uptime: {instance.uptime}</p>
        {instance.health_status === 'unhealthy' && (
          <p className="warning">
            Restart attempts: {instance.restart_count}
          </p>
        )}
      </div>

      <div className="actions">
        {instance.role === 'main' && (
          <>
            <button
              onClick={() => handleAction('stop')}
              disabled={actionInProgress === 'stop'}
            >
              Stop
            </button>
            <button
              onClick={() => handleAction('restart')}
              disabled={actionInProgress === 'restart'}
            >
              Restart
            </button>
          </>
        )}

        {instance.role === 'testing' && (
          <>
            <button
              onClick={() => handleAction('promote')}
              disabled={actionInProgress === 'promote'}
              className="primary"
            >
              Promote to Main
            </button>
            <button
              onClick={() => handleAction('stop')}
              disabled={actionInProgress === 'stop'}
            >
              Stop
            </button>
          </>
        )}

        <button
          onClick={() => handleAction('health-check')}
          disabled={actionInProgress === 'health-check'}
        >
          Health Check
        </button>
      </div>
    </div>
  );
};
```

## UI States & Transitions

### Loading State
```
┌─────────────────────────────┐
│ Deployment Status           │
│ Fetching...    ⟳ Loading   │
│                             │
│ [Disabled buttons]          │
└─────────────────────────────┘
```

### Action In-Progress State
```
┌─────────────────────────────┐
│ Deployment Status           │
│ Status: running             │
│ Health: ⟳ Checking...      │
│                             │
│ [Stop] [Restart ⟳] [Check] │
│   (Action buttons disabled) │
└─────────────────────────────┘
```

### Error State
```
┌─────────────────────────────┐
│ ⚠ DEPLOYMENT ERROR          │
│                             │
│ Failed to connect health    │
│ endpoint. Auto-restart      │
│ triggered (2 of 3 attempts).│
│                             │
│ [Retry] [View Details]      │
└─────────────────────────────┘
```

## Performance Optimization

1. **Caching**: 
   - Cache deployment status for 5 seconds
   - Only refresh on user action

2. **Polling Strategy**:
   - Poll every 30 seconds for normal updates
   - Poll every 2 seconds during action execution
   - Stop polling when action completes

3. **Batch Updates**:
   - Fetch all project deployments in one request
   - Not individual deployment endpoints

4. **Lazy Loading**:
   - Don't load backups until user clicks "View Backups"
   - Don't load logs until user clicks "View Logs"

## Error Handling

```javascript
const handleError = (error) => {
  if (response.status === 404) {
    showError('Deployment not found');
  } else if (response.status === 409) {
    showError('Another action is in progress');
  } else if (response.status === 429) {
    showError('Too many requests, please try again');
  } else if (response.status >= 500) {
    showError('Server error, please try again');
  }
};
```

## Accessibility

- Use ARIA labels for action buttons
- Color should not be only indicator (use icons/text too)
- Support keyboard navigation (Tab through buttons)
- Provide loading/status text updates for screen readers

## Testing

- Test each action button (start/stop/restart/promote)
- Test polling with delayed responses
- Test error scenarios (404, 500, timeout)
- Test rapid action clicks (should be prevented)
- Test promotion from testing to main
- Test rollback from backup
