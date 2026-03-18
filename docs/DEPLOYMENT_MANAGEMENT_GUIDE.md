# Deployment Management System Documentation

## Overview

The Deployment Management System orchestrates multi-instance deployments with:
- **Main deployment**: Always-running production instance
- **Testing deployment**: Staging instance for validation before promotion
- **Backup deployments**: Historical versions for rollback capability
- **Code signatures**: Immutable snapshots of code state at deployment time
- **Graceful transitions**: Safe promotion with automatic fallback

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   API Handlers Layer                         │
│  (Start/Stop/Restart/Retry/Rollback/Sync/HealthCheck)      │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│              Deployment Management Service                   │
│  (Orchestration, Cloning, Backup, Promotion Logic)          │
└─────────────────────┬───────────────────────────────────────┘
                      │
      ┌───────────────┼───────────────┐
      │               │               │
┌─────▼──────┐ ┌─────▼──────┐ ┌─────▼──────┐
│ Repository │ │ Git Worker │ │ Deployment │
│   Layer    │ │            │ │   Worker   │
└─────┬──────┘ └─────┬──────┘ └─────┬──────┘
      │              │              │
      └──────────────┼──────────────┘
                     │
            ┌────────▼────────┐
            │   Database      │
            │   (6 tables)    │
            └─────────────────┘
```

## Database Schema

### deployment_code_signatures
Captures code state at deployment time
```sql
- id: primary key
- deployment_id: FK to deployments
- project_id: FK to projects
- commit_sha: git commit SHA
- commit_message: git commit message
- code_hash: SHA256 of entire codebase
- file_count: number of files deployed
- directory_path: path to code directory
- created_at: timestamp
```

### deployment_instances
Running deployment instances
```sql
- id: primary key
- deployment_id: FK to deployments
- project_id: FK to projects
- role: 'main' | 'testing' | 'backup'
- status: 'preparing' | 'running' | 'stopping' | 'stopped' | 'failed'
- port: service port
- container_id: docker container ID (if applicable)
- code_signature_id: FK to code signature
- instance_dir: directory path
- health_status: 'healthy' | 'unhealthy' | 'unknown'
- restart_count: number of automatic restarts
- started_at: timestamp
- stopped_at: timestamp
- updated_at: timestamp
```

### deployment_backups
Historical backups for rollback
```sql
- id: primary key
- deployment_id: FK to deployments
- project_id: FK to projects
- instance_id: FK to instance (source)
- code_signature_id: code state at backup time
- backup_path: filesystem path
- size: backup size in bytes
- reason: why backup was created
- is_restored: whether used for rollback
- restored_at: timestamp if restored
- created_at: timestamp
```

### deployment_actions
User-triggered actions on deployments
```sql
- id: primary key
- deployment_id: FK to deployments
- instance_id: FK to instance (optional)
- project_id: FK to projects
- user_id: who triggered
- action: 'start' | 'stop' | 'restart' | 'retry' | 'rollback' | 'sync'
- status: 'pending' | 'executing' | 'success' | 'failed'
- result: result message
- created_at/updated_at: timestamps
```

### project_deployment_stats
Aggregated statistics per project
```sql
- id: primary key
- project_id: FK to projects
- main_deployment_id: current main deployment
- testing_deployment_id: current testing deployment
- total_deployments: lifetime count
- successful_deploys: count
- failed_deploys: count
- total_backups: lifetime backup count
- last_deploy_at: most recent deployment
- avg_deploy_time: average deployment duration
```

### Modified: projects table
Added deployment configuration
```sql
- max_deployments: max concurrent deployments (default: 2)
- max_backups: max backup versions to keep (default: 3)
- git_clone_path: path to cloned repository
- clone_directory: base directory for clones
- main_deploy_id: current production deployment
```

## Workflow - Auto-Clone on Project Creation

### Trigger
When a project is created via API

### Flow
```
1. API: ProjectService.CreateProject()
   └─> Database: Insert project
   └─> Event: OnProjectCreated() → DeploymentManagementWorker

2. Worker: OnProjectCreated()
   └─> Call: DeploymentManagementService.InitializeProjectClone(ctx, project)

3. Service: InitializeProjectClone()
   └─> Create: /base-dir/{projectId} directory
   └─> Git: git clone -b {branch} --depth 1 {repo_url} {clone_path}
   └─> Update: project.GitClonePath = clone_path
   └─> Database: projectRepo.Update(project)

4. Result: Test pulls now happen from /base-dir/{projectId}
```

### Code Example
```go
// Called from worker
clonePath, err := dmService.InitializeProjectClone(ctx, project)
if err != nil {
    log.Printf("Clone failed: %v", err)
    // Project creation continues, clone can retry
}
// project.GitClonePath is now set to clonePath
```

## Workflow - Code Signature Capture

### Purpose
Create immutable snapshot of code state at deployment time

### Data Captured
- Commit SHA from git
- Commit message
- SHA256 hash of entire codebase (for comparison)
- File count (to detect missing files)
- Directory path

### Flow
```
1. Deployment triggered
   └─> Service: CaptureCodeSignature(deployment, project, sourcePath)

2. Service walks directory tree
   └─> Skip: .git, hidden directories, large binary files
   └─> Hash: SHA256 of all file contents
   └─> Count: number of files

3. Record created in deployment_code_signatures
   └─> Links deployment to specific code version
   └─> Enables code comparison between deployments

4. Used later for:
   └─> Verifying instance has correct code
   └─> Comparing testing vs main for validation
   └─> Backup/restore verification
```

### Code Example
```go
sig, err := dmService.CaptureCodeSignature(deployment, project, "")
if err != nil {
    return err
}
// sig.CodeHash: "a1b2c3d4e5f6..." (SHA256)
// sig.FileCount: 42
// sig.CommitSHA: "abc123def456..."
```

## Workflow - Multi-Instance Deployment

### Deployment Strategy
```
┌─────────────────────────────────────────────────────────┐
│               Initial State (new project)               │
│  Main: RUNNING v1.0 (current production)                │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│         New Deployment Triggered (v1.1 code)            │
│                                                          │
│  1. Capture code signature from git clone              │
│  2. Create testing instance with v1.1 code             │
│  3. Keep main v1.0 running (backup NOT created yet)    │
│                                                          │
│  Result:                                                │
│  Main: RUNNING v1.0                                     │
│  Testing: RUNNING v1.1                                 │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│      Promotion Decision: v1.1 Validated OK              │
│                                                          │
│  1. Create backup of v1.0 main (pre_promotion_backup)  │
│  2. Stop main v1.0 (graceful shutdown timeout)         │
│  3. Promote testing v1.1 to main role                  │
│  4. Move old main v1.0 to backup role                  │
│  5. Keep rotating: keep only MaxBackups (e.g., 3)      │
│                                                          │
│  Result:                                                │
│  Main: RUNNING v1.1                                     │
│  Backup1: STOPPED v1.0 (pre_promotion)                 │
│  Backup2: STOPPED v0.9 (removed if > MaxBackups)       │
└─────────────────────────────────────────────────────────┘
```

### Deployment Instance Creation
```go
instance, err := dmService.CreateDeploymentInstance(
    deployment,
    project,
    signature,    // Code signature from CaptureCodeSignature
    DeploymentRoleTesting,
)
// instance.InstanceDir: /base-dir/{projectId}/testing
// instance.Role: "testing"
// instance.Status: "stopped"
// instance.Port: from project config

// Code copied from source to instance:
//   {projectId}/source/ → {projectId}/testing/
```

## Workflow - Graceful Deployment Promotion

### Main to Testing Promotion
```
1. Service: PromoteTestingToMain(project, testing, main, createBackup=true)

2. If createBackup:
   a. Call: CreateBackup(main, project, "pre_promotion_rollback")
      - Copy /base-dir/{projectId}/main to /base-dir/{projectId}/backups/{id}
      - Record backup in database with reason
      - Clean old backups if > MaxBackups

3. Stop main gracefully:
   a. Set main.Status = "stopping"
   b. Sleep for gracefulShutdownTime (e.g., 30 seconds)
      - Allows connections to drain
      - Load balancer can mark unhealthy
      - Health checks will fail
   c. Set main.Status = "stopped"

4. Promote testing:
   a. Set testing.Role = "main"
   b. Set testing.Status = "running"
   c. Update database

5. Demote old main:
   a. Set old_main.Role = "backup"
   b. Keep for rollback if needed

6. Update project:
   a. Set project.MainDeployID = testing.DeploymentID
```

### Code Example
```go
err := dmService.PromoteTestingToMain(
    project,
    testingInstance,
    mainInstance,
    createBackup: true,
    gracefulShutdownTime: 30 * time.Second,
)
// Result: testing is now main, old main is backup
```

## Workflow - Backup and Rollback

### Creating Backups
```go
backup, err := dmService.CreateBackup(
    instance,
    project,
    "pre_promotion_rollback",
)
```

**Backup Operations:**
- Copy instance directory to `/base-dir/{projectId}/backups/{id}`
- Record in database with timestamp, size, reason
- Auto-cleanup old backups if exceeding MaxBackups

**Backup Storage:**
```
/base-dir/
├── project-id-1/
│   ├── main/              (current production code)
│   ├── testing/           (staging code)
│   └── backups/
│       ├── backup_20240115_143022/
│       ├── backup_20240114_090512/
│       └── backup_20240113_151234/
├── project-id-2/
│   ├── ...
```

### Restoring from Backup
```go
err := dmService.RestoreBackup(backup, instance)
```

**Restore Process:**
1. Backup current state to temp location (safety)
2. Clear instance directory
3. Copy backup contents to instance directory
4. Mark backup as restored in database
5. Restart instance

## API Endpoints

### Deployment Actions

#### Start Deployment
```
POST /api/v1/deployments/{id}/actions/start
Authorization: Bearer {token}

Response (202 Accepted):
{
  "message": "deployment start initiated",
  "actionId": "action-uuid"
}
```

#### Stop Deployment
```
POST /api/v1/deployments/{id}/actions/stop

Response (202 Accepted):
{
  "message": "deployment stop initiated",
  "actionId": "action-uuid"
}
```

#### Restart Deployment
```
POST /api/v1/deployments/{id}/actions/restart

Response (202 Accepted):
{
  "message": "deployment restart initiated",
  "actionId": "action-uuid"
}
```

#### Retry Failed Deployment
```
POST /api/v1/deployments/{id}/actions/retry

Response (202 Accepted):
{
  "message": "deployment retry initiated",
  "actionId": "action-uuid"
}
```

#### Rollback Deployment
```
POST /api/v1/deployments/{id}/actions/rollback
Content-Type: application/json

Request:
{
  "backupId": "backup-uuid"
}

Response (202 Accepted):
{
  "message": "deployment rollback initiated",
  "actionId": "action-uuid"
}
```

#### Sync with Latest Code
```
POST /api/v1/deployments/{id}/actions/sync

Response (202 Accepted):
{
  "message": "deployment sync initiated",
  "actionId": "action-uuid"
}
```

### Backup Management

#### Get Deployment Backups
```
GET /api/v1/deployments/{id}/backups

Response (200 OK):
{
  "deploymentId": "dep-id",
  "projectId": "proj-id",
  "backups": [
    {
      "id": "backup-id-1",
      "deploymentId": "dep-id",
      "reason": "pre_promotion_rollback",
      "size": 1048576,
      "createdAt": "2024-01-15T14:30:22Z",
      "isRestored": false
    }
  ]
}
```

#### Restore Specific Backup
```
POST /api/v1/deployments/{id}/backups/{backupId}/restore

Response (202 Accepted):
{
  "message": "backup restore initiated",
  "backupId": "backup-id-1",
  "actionId": "action-uuid"
}
```

### Deployment Status

#### Get Deployment Status
```
GET /api/v1/deployments/{id}/status

Response (200 OK):
{
  "id": "deployment-id",
  "projectId": "project-id",
  "status": "running",
  "branch": "main",
  "version": "1.0.0",
  "instances": [
    {
      "id": "instance-id",
      "role": "main",
      "status": "running",
      "port": 8080,
      "healthStatus": "healthy",
      "restartCount": 0,
      "startedAt": "2024-01-15T14:00:00Z",
      "codeSignature": {
        "commitSHA": "abc123...",
        "codeHash": "def456...",
        "fileCount": 42
      }
    }
  ],
  "updatedAt": "2024-01-15T14:30:22Z"
}
```

#### Get Project Deployments
```
GET /api/v1/projects/{projectId}/deployments

Response (200 OK):
{
  "projectId": "project-id",
  "deployments": [
    {
      "id": "main-deployment-id",
      "status": "running",
      "role": "main",
      "version": "1.1.0",
      "instances": 1
    },
    {
      "id": "testing-deployment-id",
      "status": "running",
      "role": "testing",
      "version": "1.2.0-beta",
      "instances": 1
    }
  ]
}
```

### Statistics

#### Get Project Deployment Stats
```
GET /api/v1/projects/{projectId}/deployment-stats

Response (200 OK):
{
  "projectId": "project-id",
  "totalDeployments": 15,
  "activeDeployments": 2,
  "mainDeployment": {
    "id": "dep-id-1",
    "version": "1.1.0",
    "uptime": "168h30m22s",
    "healthStatus": "healthy"
  },
  "testingDeployment": {
    "id": "dep-id-2",
    "version": "1.2.0-beta",
    "uptime": "2h15m33s",
    "healthStatus": "healthy"
  },
  "backupDeployments": 3,
  "totalBackupSize": 3145728,
  "lastDeploymentTime": "2024-01-15T14:00:00Z",
  "lastSyncTime": "2024-01-15T12:30:00Z",
  "averageDeploymentTime": "5m30s"
}
```

### Configuration

#### Update Deployment Limits
```
PATCH /api/v1/projects/{projectId}/deployment-limits
Content-Type: application/json

Request:
{
  "maxDeployments": 3,
  "maxBackups": 5
}

Response (200 OK):
{
  "message": "limits updated",
  "projectId": "project-id",
  "maxDeployments": 3,
  "maxBackups": 5
}
```

### Health & Monitoring

#### Deployment Health Check
```
POST /api/v1/deployments/{id}/health-check

Response (200 OK):
{
  "deploymentId": "deployment-id",
  "status": "healthy",
  "responseTime": "45ms",
  "checkedAt": "2024-01-15T14:35:22Z"
}
```

#### Get Action Status
```
GET /api/v1/actions/{actionId}

Response (200 OK):
{
  "id": "action-id",
  "status": "success",
  "result": "deployment started",
  "createdAt": "2024-01-15T14:30:00Z",
  "updatedAt": "2024-01-15T14:30:15Z"
}
```

## Integration with Git Sync System

The Deployment Management System integrates with the git sync system:

1. **Project Clone**: Uses git sync detection to pull latest code
2. **Auto-Sync**: Detects code changes and captures new signatures
3. **Deployment Diff**: Compares code between deployments using signatures
4. **Rollback Safety**: Code signature ensures rollback integrity

## Worker Tasks

The `DeploymentManagementWorker` runs periodic tasks:

### processPendingTasks()
Runs every `tickInterval` (configurable, e.g., 5 minutes):
- Process pending actions (start/stop/restart)
- Check deployment health
- Cleanup old backups exceeding limits
- Clone repositories for new projects

### checkDeploymentHealth()
For each running instance:
- Make HTTP request to `http://localhost:{port}/health`
- If unhealthy and `restartCount < MAX_RESTARTS`: trigger restart
- Update `health_status` and `restart_count`

### cleanupOldBackups()
For each project:
- Count backups
- If count > `project.MaxBackups`: delete oldest
- Free up disk space

### processPendingProjectClones()
For projects without clone:
- Call `InitializeProjectClone()`
- Store path in `project.GitClonePath`

### Event Handlers
Can be called immediately:
- `OnProjectCreated()` - Trigger clone on creation
- `OnProjectUpdated()` - Refresh clone if branch changed
- `OnDeploymentTriggered()` - Execute full deployment workflow
- `OnActionTriggered()` - Execute deployment action

## Configuration

### Service Configuration
```go
dmService := services.NewDeploymentManagementService(
    dmRepo,
    projectRepo,
    deploymentRepo,
    "/opt/deployments",  // Base directory for clones/instances
)
```

### Worker Configuration
```go
worker := queue.NewDeploymentManagementWorker(
    projectRepo,
    deploymentRepo,
    dmRepo,
    dmService,
    5 * time.Minute,  // Check interval
)
worker.Start()
```

### Project Limits
- `MaxDeployments`: Max concurrent deployments (default: 2)
  - Usually 1 main + 1 testing
  - Can be increased for canary deployments
- `MaxBackups`: Max backup versions (default: 3)
  - Keeps 3 recent versions for quick rollback
  - Configurable per project

## Performance Considerations

### Directory Operations
- Copy operations optimized with concurrent workers
- Large files excluded from code hash (binaries)
- Shallow git clone (`--depth 1`) for faster cloning

### Database
- Indexes on frequently queried fields:
  - `project_id`, `deployment_id`, `instance_id`
  - `role`, `status`, `created_at`
- Statistics table for quick aggregation

### Health Checks
- Cached main deployment instance (most frequently accessed)
- Batch backup cleanup instead of one-by-one
- Lazy loading of backups/actions on request

## Error Handling & Recovery

### Clone Failures
- Logged but don't block project creation
- Worker retries on next tick
- User notified via project status

### Deployment Failures
- Automatic rollback to previous deployment
- Backup created before each promotion
- Manual retry option via API

### Unhealthy Instances
- Auto-restart up to max count
- After max restarts, marked as failed
- Manual intervention required

## Best Practices

1. **Graceful Shutdown**
   - Always set shutdown timeout before stopping
   - Allows existing connections to complete
   - Health checks will mark as unhealthy

2. **Code Signatures**
   - Always capture before and after deployment
   - Compare signatures when diagnosing issues
   - Use for backup validation

3. **Backup Strategy**
   - Always create backup before promotion
   - Schedule backup cleanup during low-traffic windows
   - Keep backups for at least 24 hours

4. **Monitoring**
   - Monitor action queue depth
   - Track average deployment time
   - Alert on repeated restart failures

5. **Testing Workflow**
   - Deploy to testing instance first
   - Run smoke tests before promotion
   - Use gradual traffic shift if possible

## Troubleshooting

### Clone Fails
- Check git credentials in project configuration
- Verify repository URL is reachable
- Check disk space for clone directory

### Deployment Stuck
- Check instance health status
- Review deployment action logs
- Check disk space on instance directory

### High Restart Count
- Review application logs
- Check health check endpoint implementation
- Consider increasing `MAX_RESTARTS` threshold

### Backup Cleanup Issues
- Verify backup directory permissions
- Check disk space availability
- Review cleanup logs for errors
