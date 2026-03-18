# Deployment Restart & Health Tracking

Pushpaka provides automatic deployment restart, health monitoring, and recovery mechanisms to ensure continuous service availability.

## Overview

When a deployment restarts, Pushpaka automatically:
- Tracks deployment status changes
- Monitors health checks
- Gracefully restart services
- Maintain session state
- Log restart events
- Alert on repeated failures

## Deployment Status Tracking

### Status States

```go
type DeploymentStatus string

const (
    StatusPending    DeploymentStatus = "pending"      // Waiting to start
    StatusRunning    DeploymentStatus = "running"      // Currently running
    StatusRestarting DeploymentStatus = "restarting"   // Restart in progress
    StatusStopped    DeploymentStatus = "stopped"      // Stopped
    StatusError      DeploymentStatus = "error"        // Error state
    StatusHealthy    DeploymentStatus = "healthy"      // Healthy and ready
    StatusUnhealthy  DeploymentStatus = "unhealthy"    // Health check failed
)
```

### Status Transition Flow

```
pending
  ↓
running → healthcheck → healthy
  ↓         ↓
  error ← unhealthy
  ↓
restarting
  ↓
running (repeat)
```

## Restart Mechanism

### Graceful Restart

```bash
# API Endpoint to trigger restart
POST /api/v1/projects/{projectId}/deployments/{deploymentId}/restart

Request:
{
  "graceful": true,
  "timeout": 30,
  "drain_connections": true
}

Response:
{
  "status": "restarting",
  "restart_id": "restart_abc123",
  "estimated_time": 30,
  "started_at": "2026-03-18T10:00:00Z"
}
```

### Restart Process

1. **Pre-restart checks**
   - Verify deployment exists
   - Check current status
   - Verify permissions

2. **Graceful shutdown**
   - Send SIGTERM signal
   - Wait for graceful shutdown (timeout configurable)
   - Close connections
   - Flush pending operations

3. **Service restart**
   - Clear old processes
   - Start fresh instance
   - Wait for health checks
   - Verify service is responding

4. **Post-restart validation**
   - Run health checks
   - Verify all endpoints responding
   - Check dependencies
   - Monitor logs for errors

### Hard Restart (Force)

If graceful restart fails:

```bash
POST /api/v1/projects/{projectId}/deployments/{deploymentId}/restart?force=true

# Force restart without graceful shutdown
# Immediate process termination and restart
```

## Health Monitoring

### Health Check Configuration

```yaml
# In deployment values
healthCheck:
  enabled: true
  path: /health
  port: 8080
  initialDelay: 10s
  timeout: 5s
  periodSeconds: 10
  successThreshold: 1
  failureThreshold: 3

readinessCheck:
  enabled: true
  path: /ready
  initialDelay: 5s
  timeout: 3s
  periodSeconds: 5
  failureThreshold: 2
```

### Health Check Endpoints

**API Server:**
```bash
GET /health
Response: { "status": "healthy" }

GET /ready
Response: { "status": "ready", "dependencies": { "db": "ok", "redis": "ok" } }
```

**Dashboard:**
```bash
GET /api/health
Response: { "status": "up" }
```

**Worker:**
```bash
GET /internal/health
Response: { "status": "healthy", "jobs_processed": 1234 }
```

## Automatic Restart Policies

### Restart on Failure

```yaml
restartPolicy:
  enabled: true
  maxRestarts: 5
  restartWindow: 60  # seconds
  backoffMultiplier: 1.5
  initialBackoff: 5s
  maxBackoff: 300s
```

**Process:**
1. Service fails (crash, health check failure)
2. Wait for `initialBackoff` (5s)
3. Attempt restart
4. If fails again, wait for `initialBackoff * backoffMultiplier`
5. Continue until `maxRestarts` reached
6. If `maxRestarts` exceeded, alert admin

### Restart on Update

```yaml
restartPolicy:
  onUpdate:
    enabled: true
    strategy: "RollingUpdate"
    maxUnavailable: 1
    maxSurge: 1
```

**Process:**
1. New version detected
2. Start new container with new version
3. Route traffic to new container
4. Monitor health
5. Gradually remove old containers
6. Verify no errors in logs

## Deployment State Tracking

### Tracking Table Schema

```sql
CREATE TABLE deployment_restarts
(
    id              UUID PRIMARY KEY,
    deployment_id   UUID NOT NULL,
    restart_id      VARCHAR(64),
    restart_type    VARCHAR(32),      -- graceful, hard, auto, manual
    initiated_by    VARCHAR(256),      -- username or 'system'
    status          VARCHAR(32),      -- restarting, success, failed
    reason          TEXT,             -- why it was restarted
    duration_ms     INTEGER,
    exit_code       INTEGER,
    error_message   TEXT,
    started_at      TIMESTAMP,
    completed_at    TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deployment_restarts ON deployment_restarts(deployment_id, created_at DESC);
```

### Track API Calls

```go
// Log every restart event
POST /api/v1/projects/{projectId}/deployments/{deploymentId}/restart
→ INSERT INTO deployment_restarts (...)

// Query restart history
GET /api/v1/projects/{projectId}/deployments/{deploymentId}/restarts?limit=50
→ SELECT * FROM deployment_restarts WHERE deployment_id = ?
```

## Monitoring & Alerts

### Restart Metrics

```bash
# Exported to Prometheus
pushpaka_deployment_restarts_total{deployment="myapp"} 5
pushpaka_deployment_restart_duration_seconds{deployment="myapp"} 15.2
pushpaka_deployment_health_checks_failed_total{deployment="myapp"} 2
pushpaka_deployment_status{deployment="myapp", status="healthy"} 1
```

### Alert Rules

```yaml
# Prometheus alert rules
- alert: FrequentRestarts
  expr: rate(pushpaka_deployment_restarts_total[5m]) > 0.1
  for: 5m
  annotations:
    summary: "Deployment {{ $labels.deployment }} restarting too frequently"

- alert: HealthCheckFailures
  expr: pushpaka_deployment_health_checks_failed_total > 10
  for: 1m
  annotations:
    summary: "Deployment {{ $labels.deployment }} health checks failing"
```

## Server Restart Implementation

### Backend Handler

```go
// backend/handlers/deployment.go

func RestartDeployment(c *gin.Context) {
    projectID := c.Param("projectId")
    deploymentID := c.Param("deploymentId")
    
    var req struct {
        Graceful       bool   `json:"graceful" default:"true"`
        Timeout        int    `json:"timeout" default:"30"`
        DrainConnections bool `json:"drain_connections" default:"true"`
    }
    c.BindJSON(&req)
    
    // Get deployment
    deployment := repository.GetDeployment(projectID, deploymentID)
    if deployment == nil {
        c.JSON(404, gin.H{"error": "deployment not found"})
        return
    }
    
    // Check if running
    if deployment.Status != "running" {
        c.JSON(400, gin.H{"error": "deployment not running"})
        return
    }
    
    // Start restart process
    restartID := uuid.New().String()
    
    go func() {
        // Log restart event
        repository.LogRestart(deploymentID, restartID, "manual", "user initiated")
        
        if req.Graceful {
            // Graceful shutdown
            service.GracefulShutdown(deployment, req.Timeout)
        } else {
            // Hard restart
            service.ForcedShutdown(deployment)
        }
        
        // Wait for service to stop
        time.Sleep(2 * time.Second)
        
        // Restart service
        service.StartDeployment(deployment)
        
        // Monitor health
        for i := 0; i < 10; i++ {
            if service.HealthCheck(deployment) {
                repository.UpdateRestartStatus(restartID, "success")
                return
            }
            time.Sleep(time.Second)
        }
        
        repository.UpdateRestartStatus(restartID, "failed")
    }()
    
    c.JSON(200, gin.H{
        "status": "restarting",
        "restart_id": restartID,
    })
}

// Handle automatic restarts on failure
func AutoRestartWorker() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        deployments := repository.GetAll Deployments()
        for _, dep := range deployments {
            if !service.HealthCheck(dep) {
                if dep.RestartCount < dep.MaxRestarts {
                    service.StartDeployment(dep)
                    dep.RestartCount++
                    repository.Update(dep)
                } else {
                    // Alert admin
                    alerts.SendAlert("Max restarts exceeded for " + dep.ID)
                }
            }
        }
    }
}
```

### Database Migration

```sql
-- migrations/2026_03_18_deployment_restarts.sql

CREATE TABLE IF NOT EXISTS deployment_restarts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id),
    restart_id VARCHAR(64) UNIQUE,
    restart_type VARCHAR(32),
    initiated_by VARCHAR(256),
    status VARCHAR(32),
    reason TEXT,
    duration_ms INTEGER,
    exit_code INTEGER,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deployment_restarts_deployment_id 
ON deployment_restarts(deployment_id, created_at DESC);

CREATE INDEX idx_deployment_restarts_status 
ON deployment_restarts(status, created_at DESC);

-- Add restart tracking to deployments table
ALTER TABLE deployments 
ADD COLUMN IF NOT EXISTS last_restart_at TIMESTAMP;

ALTER TABLE deployments 
ADD COLUMN IF NOT EXISTS restart_count INTEGER DEFAULT 0;

ALTER TABLE deployments 
ADD COLUMN IF NOT EXISTS max_restarts INTEGER DEFAULT 5;
```

## Dashboard Integration

### Restart Status Widget

```tsx
// frontend/components/DeploymentRestartStatus.tsx

export function DeploymentRestartStatus({ deployment }) {
  const [restarts, setRestarts] = useState([]);
  const [isRestarting, setIsRestarting] = useState(false);

  const handleRestart = async (graceful = true) => {
    setIsRestarting(true);
    try {
      const response = await fetch(
        `/api/v1/projects/${projectId}/deployments/${deployment.id}/restart`,
        {
          method: 'POST',
          body: JSON.stringify({ graceful })
        }
      );
      const data = await response.json();
      
      // Poll for status
      const checkStatus = setInterval(async () => {
        const status = await fetch(`/api/v1/restarts/${data.restart_id}`);
        if (status.complete) {
          clearInterval(checkStatus);
          setIsRestarting(false);
        }
      }, 1000);
    } catch (error) {
      console.error('Restart failed:', error);
      setIsRestarting(false);
    }
  };

  return (
    <div className="deployment-restart-widget">
      <h3>Deployment Status</h3>
      
      <div className={`status-indicator ${deployment.status}`}>
        {deployment.status}
      </div>

      {isRestarting && <LoadingSpinner />}

      <div className="restart-buttons">
        <button 
          onClick={() => handleRestart(true)}
          disabled={isRestarting}
        >
          Graceful Restart
        </button>
        <button
          onClick={() => {
            if (confirm('Force restart will immediately stop the service. Continue?')) {
              handleRestart(false);
            }
          }}
          disabled={isRestarting}
          className="danger"
        >
          Force Restart
        </button>
      </div>

      <div className="restart-history">
        <h4>Recent Restarts</h4>
        {restarts.map((restart) => (
          <div key={restart.id} className="restart-item">
            <span className="timestamp">{restart.started_at}</span>
            <span className="type">{restart.restart_type}</span>
            <span className={`status ${restart.status}`}>{restart.status}</span>
            <span className="reason">{restart.reason}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
```

## API Reference

### Restart Deployment

```
POST /api/v1/projects/{projectId}/deployments/{deploymentId}/restart

Request:
{
  "graceful": true,           // optional, default: true
  "timeout": 30,              // optional, seconds, default: 30
  "drain_connections": true   // optional, default: true
}

Response 200 OK:
{
  "status": "restarting",
  "restart_id": "restart_abc123",
  "estimated_time": 30,
  "started_at": "2026-03-18T10:00:00Z"
}
```

### Get Restart Status

```
GET /api/v1/restarts/{restartId}

Response 200 OK:
{
  "id": "restart_abc123",
  "deployment_id": "dep_xyz",
  "status": "restarting",      // restarting, success, failed
  "progress": 50,
  "started_at": "2026-03-18T10:00:00Z",
  "estimated_completion": "2026-03-18T10:00:30Z"
}
```

### Get Restart History

```
GET /api/v1/projects/{projectId}/deployments/{deploymentId}/restarts?limit=50

Response 200 OK:
{
  "restarts": [
    {
      "id": "restart_123",
      "restart_type": "manual",
      "initiated_by": "john@example.com",
      "status": "success",
      "duration_ms": 15000,
      "reason": "User requested restart",
      "started_at": "2026-03-18T10:00:00Z",
      "completed_at": "2026-03-18T10:00:15Z"
    }
  ],
  "total": 1
}
```

## Best Practices

✅ **DO:**
- Monitor restart frequency
- Set appropriate timeout values
- Use graceful restarts by default
- Alert on repeated failures
- Log restart reasons
- Test restart procedure regularly

❌ **DON'T:**
- Force restart without good reason
- Ignore frequent restarts
- Restart during critical operations
- Skip health checks
- Restart without backups
- Ignore restart error logs

---

**Last Updated:** March 18, 2026  
**Version:** 1.0
