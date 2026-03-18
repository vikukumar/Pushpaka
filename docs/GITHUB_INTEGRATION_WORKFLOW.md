# GitHub Integration - Approval Workflow

Pushpaka auto-detects GitHub repositories and provides intelligent deployment workflow with approval gates, automated testing, and rollback capabilities.

## Overview

When you link a GitHub repository to Pushpaka:

1. **Track latest code** - Monitor new commits and pull requests
2. **Pre-deployment testing** - Run tests before deploying
3. **Approval workflow** - Require approval before deployment (optional auto-approve)
4. **Version tracking** - Track which version is deployed
5. **Rollback on failure** - Automatic rollback if deployment fails

## GitHub Repository Linking

### Connect Repository

```bash
POST /api/v1/projects/{projectId}/github/connect

Request:
{
  "repository": "owner/repo",
  "branch": "main",
  "github_token": "ghp_xxxxx",  # Personal Access Token (write:repo_hook required)
  "auto_approve": false,
  "required_reviews": 1,
  "require_passing_checks": true
}

Response 200 OK:
{
  "id": "github_integration_abc",
  "repository": "owner/repo",
  "branch": "main",
  "connected_at": "2026-03-18T10:00:00Z",
  "webhook_url": "https://pushpaka.example.com/webhooks/github/abc123",
  "status": "active"
}
```

### Webhook Setup

Pushpaka automatically:
1. Creates GitHub webhook for `push` and `pull_request` events
2. Registers webhook URL with GitHub
3. Validates webhook secret
4. Starts monitoring repository events

## Deployment Workflow

### State Machine

```
Idle
  ↓
New commit detected (webhook)
  ↓
Fetch latest code
  ↓
Run pre-deployment tests
  ├→ ✅ Tests pass → Pending Approval
  └→ ❌ Tests fail → Mark as failed, notify
       Pending Approval
       ├→ ✅ Auto-approved → Deploying
       ├→ ✅ Manual approval → Deploying
       └→ ❌ Rejected → Cancelled
              Deploying
              ├→ ✅ Deployment success → Active
              └→ ❌ Deployment failed → Rollback
                     Active (running new version)
                     Rollback (back to previous version)
```

## GitHub Workflow Configuration

### GitHub Actions Integration

Create `.github/workflows/pushpaka-deploy.yml`:

```yaml
name: Pushpaka Deployment Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Backend tests
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run Go tests
        working-directory: ./backend
        run: |
          go test -v ./...
          go build -o pushpaka

      # Frontend tests
      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        working-directory: ./frontend
        run: npm ci

      - name: Run linter
        working-directory: ./frontend
        run: npm run lint

      - name: Run tests
        working-directory: ./frontend
        run: npm run test

      - name: Build
        working-directory: ./frontend
        run: npm run build

      # Integration tests
      - name: Run integration tests
        run: |
          docker compose up -d
          npm run test:integration
          docker compose down

  notify-pushpaka:
    needs: test
    runs-on: ubuntu-latest
    if: success()
    steps:
      - uses: actions/checkout@v4

      - name: Notify Pushpaka
        run: |
          curl -X POST https://pushpaka.example.com/api/v1/github/webhook \
            -H "Content-Type: application/json" \
            -H "X-GitHub-Event: ${{ github.event_name }}" \
            -d @- <<'EOF'
          {
            "action": "${{ github.event.action }}",
            "repository": "${{ github.repository }}",
            "branch": "${{ github.ref_name }}",
            "commit": "${{ github.sha }}",
            "author": "${{ github.actor }}",
            "message": "${{ github.event.head_commit.message }}",
            "test_status": "success"
          }
          EOF
```

## Pre-Deployment Testing

### Test Suite Definition

```yaml
# .pushpaka/tests.yaml
tests:
  backend:
    - name: Unit Tests
      command: cd backend && go test -v ./...
      timeout: 300
      required: true
      
    - name: Integration Tests
      command: docker compose -f docker-compose.test.yml up && npm run test:integration
      timeout: 600
      required: true
      
  frontend:
    - name: Lint
      command: cd frontend && npm run lint
      timeout: 120
      required: true
      
    - name: Build
      command: cd frontend && npm run build
      timeout: 300
      required: true
      
  security:
    - name: Dependency Check
      command: npm audit --audit-level=moderate
      timeout: 180
      required: true
      
    - name: Code Scan
      command: trivy fs .
      timeout: 300
      required: false

approval:
  required: true
  auto_approve_on_success: false  # Requires manual approval
  timeout: 86400  # 24 hours to approve
  reviewers:
    - vikukumar
    - tech-leads
```

### Test Execution

```bash
# Pushpaka runs tests when new code is detected

# 1. Fetch latest code
git fetch origin
git checkout <new-commit>

# 2. Run pre-deployment tests
npm run test
go test ./...
trivy fs .

# 3. Grade results
PASS ✅ → proceed to approval
FAIL ❌ → notify and stop

# 4. Wait for approval
Send notification to approvers
Wait for approval (or auto-approve if configured)

# 5. Deploy
git checkout <new-commit>
./pushpaka deploy
```

## Approval Workflow

### Manual Approval

```bash
POST /api/v1/deployments/{deploymentId}/approve

Request:
{
  "approved": true,
  "reason": "Tested in staging, looks good"
}

Response 200 OK:
{
  "id": "dep_123",
  "status": "deploying",
  "approval": {
    "approved_by": "john@example.com",
    "approved_at": "2026-03-18T10:05:00Z",
    "reason": "Tested in staging, looks good"
  }
}
```

### Auto-Approval

```yaml
# Configuration for auto-approval
approval:
  auto_approve: true
  conditions:
    - all_tests_pass: true      # ✅ Required
    - no_breaking_changes: true  # ✅ Detected via analysis
    - approved_by_github: false  # Branch protection rules
```

### Approval Dashboard UI

```tsx
// frontend/components/DeploymentApproval.tsx

export function DeploymentApprovalPanel({ deployment }) {
  return (
    <div className="approval-panel">
      <h3>Deployment Pending Approval</h3>

      <div className="deployment-info">
        <p><strong>Version:</strong> {deployment.version}</p>
        <p><strong>Commit:</strong> {deployment.commit}</p>
        <p><strong>Author:</strong> {deployment.author}</p>
        <p><strong>Changes:</strong> {deployment.changes_count} files changed</p>
      </div>

      <div className="test-results">
        <h4>Pre-Deployment Tests</h4>
        {deployment.tests.map(test => (
          <div key={test.id} className={`test-result status-${test.status}`}>
            <span>{test.name}</span>
            <span className="status">{test.status}</span>
            <span className="time">{test.duration}s</span>
          </div>
        ))}
      </div>

      <div className="approval-buttons">
        <button 
          className="approve-btn"
          onClick={() => handleApprove(true)}
        >
          ✅ Approve & Deploy
        </button>
        <button
          className="reject-btn"
          onClick={() => handleApprove(false)}
        >
          ❌ Reject
        </button>
      </div>

      <textarea
        placeholder="Approval reason/notes"
        onChange={(e) => setReason(e.target.value)}
      />
    </div>
  );
}
```

## Version Tracking

### Tracking Deployments

```sql
CREATE TABLE deployment_versions (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL,
  version VARCHAR(32),
  commit_hash VARCHAR(64),
  author VARCHAR(256),
  message TEXT,
  branch VARCHAR(128),
  status VARCHAR(32),      -- pending, deployed, failed, rolled_back
  test_status VARCHAR(32), -- passed, failed, skipped
  deployed_by VARCHAR(256),
  deployed_at TIMESTAMP,
  rolled_back_at TIMESTAMP,
  rollback_reason TEXT,
  created_at TIMESTAMP
);
```

### Version API

```bash
GET /api/v1/projects/{projectId}/versions

Response:
{
  "versions": [
    {
      "version": "1.2.3",
      "commit": "abc123def",
      "author": "john@example.com",
      "deployed_at": "2026-03-18T09:00:00Z",
      "status": "active",
      "previous_version": "1.2.2",
      "deployment_time": 45,
      "tests": {
        "passed": 45,
        "failed": 0,
        "skipped": 2
      }
    }
  ]
}
```

## Automatic Rollback

### Rollback Triggers

```yaml
rollback:
  enabled: true
  triggers:
    - health_checks_failing:
        threshold: 3           # 3 consecutive failures
        window: 300            # seconds
    - error_rate_spike:
        threshold: 5           # 5% error rate
        baseline: 0.5          # 0.5% normal
        window: 60
    - cpu_memory_spike:
        cpu_threshold: 90      # % usage
        memory_threshold: 85   # % usage
        duration: 120          # seconds
    - critical_exceptions:
        keywords: [panic, fatal, deadlock]
        count: 5
```

### Rollback Process

```bash
# Automatic rollback triggered

# 1. Detect issue
Deployment error detected: Health check failing

# 2. Decision
Automatic rollback enabled? YES → Proceed
Rollback already attempted? NO → Proceed

# 3. Notify
Send alert to admins
Disable auto-approval for config
Create incident ticket

# 4. Rollback
git checkout previous-commit
./pushpaka deploy

# 5. Verify
Run health checks
Verify no errors
Confirm rollback successful

# 6. Investigation
Log detailed incident report
Mark deployment as failed
Require manual review before retry
```

## Notifications

### Approval Notifications

```
📬 New Deployment Pending Approval

Version: 1.2.3
Commit: abc123def456 (feat: Add new feature)
Author: john@example.com
Tests: ✅ All passed (45/45)

Action Required:
👤 Approvers: @vikukumar, @tech-leads
⏱️ Timeout: 24 hours

[✅ Approve] [❌ Reject] [View Details]
```

### Deployment Notifications

```
✅ Deployment Successful

Version: 1.2.3
Environment: Production
Deployed by: john@example.com
Duration: 45 seconds
Previous: 1.2.2

[View Logs] [Rollback]
```

### Failure Notifications

```
❌ Deployment Failed

Version: 1.2.3
Error: Health check failed
Logs: [View]
Auto-rollback: Starting...

[View Incident] [Retry] [Force Deploy]
```

## Best Practices

✅ **DO:**
- Always run tests before deployment
- Require approvals for production
- Monitor deployment health
- Set automatic rollback triggers
- Keep approval timeout reasonable
- Document deployment history

❌ **DON'T:**
- Skip tests for urgent deployments
- Auto-approve without testing
- Deploy during maintenance windows
- Ignore rollback triggers
- Forget to update version numbers
- Skip health checks

## Configuration Examples

### High-Safety (Manual approval, auto-rollback)
```yaml
approval:
  required: true
  auto_approve: false
  timeout: 3600

rollback:
  enabled: true
  auto_rollback_on_failure: true
```

### Medium-Safety (Auto-approve on test pass, rollback enabled)
```yaml
approval:
  auto_approve_on_success: true
  timeout: 300

rollback:
  enabled: true
  triggers: [health_checks_failing, error_rate_spike]
```

### Fast-Deploy (Auto-approve, minimal rollback)
```yaml
approval:
  auto_approve: true

rollback:
  enabled: true
  triggers: [critical_exceptions]
```

---

**Last Updated:** March 18, 2026
