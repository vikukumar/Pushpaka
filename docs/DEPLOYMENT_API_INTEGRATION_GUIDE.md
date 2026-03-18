# Deployment Management - API Integration & Examples

## API Integration Patterns

### Pattern 1: Async Actions with Polling

All deployment actions are asynchronous and return a 202 Accepted response with an action ID.

```go
// Client-side pattern
type ActionResponse struct {
    Message  string `json:"message"`
    ActionID string `json:"actionId"`
}

// Make request
resp, err := http.Post(
    fmt.Sprintf("http://localhost:8080/api/v1/deployments/%s/actions/restart", deploymentID),
    "application/json",
    nil,
)

body, _ := ioutil.ReadAll(resp.Body)
var actionResp ActionResponse
json.Unmarshal(body, &actionResp)

actionID := actionResp.ActionID

// Poll for completion
pollAction := func() {
    pollResp, _ := http.Get(
        fmt.Sprintf("http://localhost:8080/api/v1/actions/%s", actionID),
    )
    var action struct {
        Status string `json:"status"`
        Result string `json:"result"`
    }
    json.NewDecoder(pollResp.Body).Decode(&action)
    
    if action.Status == "success" {
        log.Printf("Action completed: %s", action.Result)
    } else if action.Status == "failed" {
        log.Printf("Action failed: %s", action.Result)
    } else {
        time.Sleep(2 * time.Second)
        pollAction()
    }
}

pollAction()
```

### Pattern 2: Request/Response with JSON Body

For actions that require parameters (like rollback with backupId):

```go
// Client making request
type RollbackRequest struct {
    BackupID string `json:"backupId"`
}

payload := RollbackRequest{BackupID: backup.ID}
jsonData, _ := json.Marshal(payload)

resp, err := http.Post(
    fmt.Sprintf("http://localhost:8080/api/v1/deployments/%s/actions/rollback", deploymentID),
    "application/json",
    bytes.NewBuffer(jsonData),
)

var actionResp ActionResponse
json.NewDecoder(resp.Body).Decode(&actionResp)
```

### Pattern 3: Query Parameters for Filtering

```go
// Get deployments filtered by status
url := fmt.Sprintf(
    "http://localhost:8080/api/v1/projects/%s/deployments?status=running&role=main",
    projectID,
)
resp, _ := http.Get(url)

var deployments struct {
    Deployments []map[string]interface{} `json:"deployments"`
}
json.NewDecoder(resp.Body).Decode(&deployments)
```

## Complete Integration Example

### Scenario: Deploy New Version to Testing, Promote if Healthy

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
)

type DeploymentClient struct {
    baseURL string
    client  *http.Client
}

func NewDeploymentClient(baseURL string) *DeploymentClient {
    return &DeploymentClient{
        baseURL: baseURL,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

// Step 1: Get current project deployments
func (c *DeploymentClient) GetProjectDeployments(projectID string) (map[string]interface{}, error) {
    resp, err := c.client.Get(
        fmt.Sprintf("%s/api/v1/projects/%s/deployments", c.baseURL, projectID),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

// Step 2: Trigger deployment action
func (c *DeploymentClient) TriggerAction(deploymentID, action string) (string, error) {
    resp, err := c.client.Post(
        fmt.Sprintf("%s/api/v1/deployments/%s/actions/%s", c.baseURL, deploymentID, action),
        "application/json",
        nil,
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var response struct {
        ActionID string `json:"actionId"`
    }
    json.NewDecoder(resp.Body).Decode(&response)
    return response.ActionID, nil
}

// Step 3: Poll action until completion
func (c *DeploymentClient) WaitForAction(ctx context.Context, actionID string) (string, error) {
    for {
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        default:
        }

        resp, err := c.client.Get(
            fmt.Sprintf("%s/api/v1/actions/%s", c.baseURL, actionID),
        )
        if err != nil {
            return "", err
        }

        var action struct {
            Status string `json:"status"`
            Result string `json:"result"`
        }
        json.NewDecoder(resp.Body).Decode(&action)

        if action.Status == "success" {
            return action.Result, nil
        } else if action.Status == "failed" {
            return "", fmt.Errorf("action failed: %s", action.Result)
        }

        time.Sleep(2 * time.Second)
    }
}

// Step 4: Check deployment health
func (c *DeploymentClient) CheckHealth(deploymentID string) (bool, error) {
    resp, err := c.client.Post(
        fmt.Sprintf("%s/api/v1/deployments/%s/health-check", c.baseURL, deploymentID),
        "application/json",
        nil,
    )
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    var healthResp struct {
        Status string `json:"status"`
    }
    json.NewDecoder(resp.Body).Decode(&healthResp)
    return healthResp.Status == "healthy", nil
}

// Step 5: Get deployment statistics
func (c *DeploymentClient) GetStats(projectID string) (map[string]interface{}, error) {
    resp, err := c.client.Get(
        fmt.Sprintf("%s/api/v1/projects/%s/deployment-stats", c.baseURL, projectID),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var stats map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&stats)
    return stats, nil
}

// Main workflow
func (c *DeploymentClient) DeployAndPromoteWorkflow(
    ctx context.Context,
    projectID string,
    healthCheckWaitTime time.Duration,
) error {
    // Step 1: Get current deployments
    log.Println("Getting current deployments...")
    deployments, err := c.GetProjectDeployments(projectID)
    if err != nil {
        return fmt.Errorf("failed to get deployments: %w", err)
    }

    testingDeployments := deployments["deployments"].([]interface{})
    if len(testingDeployments) < 2 {
        return fmt.Errorf("no testing deployment available")
    }

    testingDeploy := testingDeployments[1].(map[string]interface{})
    testingID := testingDeploy["id"].(string)

    log.Printf("Found testing deployment: %s", testingID)

    // Step 2: Wait for health (if just deployed)
    log.Println("Waiting for deployment health checks...")
    time.Sleep(healthCheckWaitTime)

    // Step 3: Check health
    healthy, err := c.CheckHealth(testingID)
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }

    if !healthy {
        return fmt.Errorf("deployment not healthy, skipping promotion")
    }

    log.Println("Deployment is healthy, promoting to main...")

    // Step 4: Trigger promotion
    actionID, err := c.TriggerAction(testingID, "promote")
    if err != nil {
        return fmt.Errorf("failed to trigger promotion: %w", err)
    }

    log.Printf("Promotion action triggered: %s", actionID)

    // Step 5: Wait for completion
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    result, err := c.WaitForAction(ctx, actionID)
    if err != nil {
        return fmt.Errorf("promotion failed: %w", err)
    }

    log.Printf("Promotion completed: %s", result)

    // Step 6: Get stats
    stats, err := c.GetStats(projectID)
    if err != nil {
        return fmt.Errorf("failed to get stats: %w", err)
    }

    log.Printf("New deployment stats: %+v", stats)
    return nil
}

// Usage
func main() {
    client := NewDeploymentClient("http://localhost:8080")
    ctx := context.Background()

    err := client.DeployAndPromoteWorkflow(ctx, "project-id-123", 30*time.Second)
    if err != nil {
        log.Fatalf("Workflow failed: %v", err)
    }

    log.Println("Workflow completed successfully")
}
```

## Service Integration Patterns

### Pattern: Direct Service Call from Other Services

```go
// Other services can directly call DeploymentManagementService

// Example: GitSyncService wants to sync and trigger deployment
func (s *GitSyncService) SyncAndDeploy(ctx context.Context, project *models.Project) error {
    // 1. Perform git sync
    changes, err := s.detectChanges(project)
    if err != nil {
        return err
    }

    if len(changes) == 0 {
        return nil  // No changes to deploy
    }

    // 2. Refresh project clone with latest code
    if err := s.dmService.RefreshProjectClone(ctx, project, project.Branch); err != nil {
        return fmt.Errorf("failed to refresh clone: %w", err)
    }

    // 3. Create new deployment with updated code
    deployment := &models.Deployment{
        ID:        uuid.New().String(),
        ProjectID: project.ID,
        Branch:    project.Branch,
        Status:    "pending",
        CreatedAt: models.Time(time.Now()),
    }

    // 4. Capture code signature (new version)
    sig, err := s.dmService.CaptureCodeSignature(deployment, project, "")
    if err != nil {
        return fmt.Errorf("failed to capture signature: %w", err)
    }

    log.Printf("New code signature: %s (hash: %s)", sig.ID, sig.CodeHash)

    // 5. Create testing instance
    testingInstance, err := s.dmService.CreateDeploymentInstance(
        deployment,
        project,
        sig,
        models.DeploymentRoleTesting,
    )
    if err != nil {
        return fmt.Errorf("failed to create testing instance: %w", err)
    }

    log.Printf("Testing instance created: %s", testingInstance.ID)

    // 6. Record action for tracking
    action, err := s.dmService.RecordAction(
        deployment.ID,
        testingInstance.ID,
        project.ID,
        "git-sync-auto",
        models.DeploymentActionSync,
    )
    if err != nil {
        log.Printf("Failed to record action: %v", err)
    }

    // 7. Start testing instance
    // TODO: Implement start logic in service
    s.dmService.UpdateActionStatus(action.ID, "executing", "")

    return nil
}
```

## Hook Integration - When to Call DeploymentManagementWorker Events

### When Project is Created
```go
// In ProjectService.CreateProject()
func (s *ProjectService) CreateProject(ctx context.Context, proj *models.Project) error {
    // 1. Create project in database
    if err := s.projectRepo.Create(proj); err != nil {
        return err
    }

    // 2. Trigger immediate clone (via worker event)
    go s.deploymentWorker.OnProjectCreated(ctx, proj)

    return nil
}
```

### When Project is Updated (Repo/Branch Change)
```go
// In ProjectService.UpdateProject()
func (s *ProjectService) UpdateProject(ctx context.Context, proj *models.Project) error {
    // Get old project to check what changed
    oldProject, err := s.projectRepo.GetByID(proj.ID)
    if err != nil {
        return err
    }

    oldBranch := oldProject.Branch

    // Update project
    if err := s.projectRepo.Update(proj); err != nil {
        return err
    }

    // If branch or repo URL changed, refresh clone
    if proj.RepoURL != oldProject.RepoURL || proj.Branch != oldBranch {
        go s.deploymentWorker.OnProjectUpdated(ctx, proj, oldBranch)
    }

    return nil
}
```

### When Deployment is Triggered
```go
// In DeploymentService
func (s *DeploymentService) TriggerDeployment(ctx context.Context, deployment *models.Deployment) error {
    // 1. Create deployment
    if err := s.deploymentRepo.Create(deployment); err != nil {
        return err
    }

    // 2. Trigger deployment workflow (async)
    go s.worker.OnDeploymentTriggered(ctx, deployment)

    return nil
}
```

### When User Triggers Action
```go
// In DeploymentManagementHandler
func (h *DeploymentManagementHandler) RestartDeployment(c *gin.Context) {
    deploymentID := c.Param("id")
    userID := c.GetString("user_id")

    // 1. Record action
    action, err := h.dmService.RecordAction(
        deploymentID,
        "",
        deployment.ProjectID,
        userID,
        models.DeploymentActionRestart,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record action"})
        return
    }

    // 2. Trigger worker to execute action
    go h.worker.OnActionTriggered(c.Request.Context(), action)

    c.JSON(http.StatusAccepted, gin.H{
        "message":  "deployment restart initiated",
        "actionId": action.ID,
    })
}
```

## Error Handling Patterns

### Pattern: Graceful Degradation

```go
// If backup creation fails, continue deployment (don't block)
backup, err := s.dmService.CreateBackup(instance, project, "pre_deployment")
if err != nil {
    log.Printf("WARNING: Failed to create backup: %v", err)
    // Continue anyway - backup is nice-to-have, not required
}

// Continue with deployment even if backup failed
err = s.dmService.PromoteTestingToMain(...)
```

### Pattern: Retry Logic

```go
// Retry failed operations
func retryWithBackoff(fn func() error, maxAttempts int) error {
    var lastErr error
    for i := 0; i < maxAttempts; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        lastErr = err

        backoffTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        log.Printf("Attempt %d failed, retrying in %v: %v", i+1, backoffTime, err)
        time.Sleep(backoffTime)
    }
    return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// Usage
err := retryWithBackoff(func() error {
    return s.dmService.InitializeProjectClone(ctx, project)
}, 3)
```

### Pattern: Operation Timeout

```go
// All operations should have timeouts
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

sig, err := s.dmService.CaptureCodeSignature(deployment, project, "")
if err == context.DeadlineExceeded {
    log.Printf("Code signature capture timed out")
    return err
}
```

## Monitoring & Observability

### Logging Pattern

```go
func (s *DeploymentManagementService) CreateDeploymentInstance(
    deployment *models.Deployment,
    project *models.Project,
    sig *models.DeploymentCodeSignature,
    role models.DeploymentRole,
) (*models.DeploymentInstance, error) {
    log.Printf(
        "[DEPLOYMENT] Creating instance: deployment=%s, project=%s, role=%s",
        deployment.ID,
        project.ID,
        role,
    )

    // Create instance directory
    instanceDir := filepath.Join(s.baseDir, project.ID, string(role))
    if err := os.MkdirAll(instanceDir, 0755); err != nil {
        log.Printf("[ERROR] Failed to create instance directory: %v", err)
        return nil, err
    }

    log.Printf("[DEPLOYMENT] Instance directory created: %s", instanceDir)

    // ... rest of logic
}
```

### Metrics Pattern

```go
// Track action counts and status
type ActionMetrics struct {
    TotalActions      int
    SuccessfulActions int
    FailedActions     int
    AvgExecutionTime  time.Duration
}

// Update metrics after action completes
metrics.TotalActions++
if action.Status == "success" {
    metrics.SuccessfulActions++
} else if action.Status == "failed" {
    metrics.FailedActions++
}
```

## Testing Patterns

### Test: Clone on Project Creation

```go
func TestProjectCloneOnCreation(t *testing.T) {
    // Setup
    dmService := NewMockDeploymentManagementService()
    projectService := NewProjectService(dmService)

    project := &models.Project{
        ID:      "test-project-1",
        RepoURL: "https://github.com/user/repo.git",
        Branch:  "main",
    }

    // Execute
    err := projectService.CreateProject(context.Background(), project)

    // Verify
    assert.NoError(t, err)
    assert.NotEmpty(t, project.GitClonePath)
    assert.True(t, dirExists(project.GitClonePath))
}
```

### Test: Action Workflow

```go
func TestRestartDeploymentAction(t *testing.T) {
    // Setup
    handler := NewDeploymentManagementHandler(dmService, deployService)

    // Execute
    handleRestartDeployment(handler, "deployment-id-123")

    // Verify
    action := dmService.GetLastAction()
    assert.Equal(t, "restart", action.Action)
    assert.Equal(t, "pending", action.Status)
}
```

## Performance Tuning

### Concurrent Operations

```go
// Process multiple backups in parallel
var wg sync.WaitGroup
for _, backup := range oldBackups {
    wg.Add(1)
    go func(b *models.DeploymentBackup) {
        defer wg.Done()
        if err := os.RemoveAll(b.BackupPath); err != nil {
            log.Printf("Error removing backup: %v", err)
        }
    }(backup)
}
wg.Wait()
```

### Caching

```go
// Cache main deployment instance (frequently accessed)
type CachedDeploymentService struct {
    mainInstanceCache *models.DeploymentInstance
    cacheTTL          time.Duration
    lastCacheTime     time.Time
    service           *DeploymentManagementService
}

func (c *CachedDeploymentService) GetMainInstance(projectID string) (*models.DeploymentInstance, error) {
    if time.Since(c.lastCacheTime) < c.cacheTTL && c.mainInstanceCache != nil {
        return c.mainInstanceCache, nil
    }

    instance, err := c.service.GetMainDeploymentInstance(projectID)
    if err != nil {
        return nil, err
    }

    c.mainInstanceCache = instance
    c.lastCacheTime = time.Now()
    return instance, nil
}
```

## Troubleshooting Common Issues

### Issue: Clone Fails
```go
// Check credentials first
if project.IsPrivate && project.GitToken == "" {
    log.Printf("Private repo requires token")
    return "", errors.New("missing git token")
}

// Check network connectivity
if !canConnect(project.RepoURL) {
    log.Printf("Cannot reach repository: %s", project.RepoURL)
    return "", errors.New("network error")
}

// Retry with exponential backoff
return retryWithBackoff(func() error {
    return cloneRepo(project)
}, 3)
```

### Issue: Disk Space Full
```go
// Check disk space before backup
available, err := getDiskSpace(s.baseDir)
if err != nil {
    return nil, err
}

estimatedBackupSize := getDirectorySize(instanceDir)
if available < estimatedBackupSize*2 {  // 2x safety margin
    return nil, fmt.Errorf("insufficient disk space")
}

// Clean old backups more aggressively
cleanupOldBackups(project, maxBackups: 1)
```

### Issue: Health Check Hangs
```go
// Always use timeout
client := &http.Client{
    Timeout: 5 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout: 3 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout: 3 * time.Second,
    },
}

status, err := getHealthStatus(client, instance.Port)
```
