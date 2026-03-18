# Pushpaka vs ArgoCD: Git-Based Deployment Comparison

A comparison between Pushpaka's GitOps implementation and ArgoCD, highlighting similarities, differences, and integration patterns.

## Feature Comparison

| Feature | Pushpaka | ArgoCD | Notes |
|---------|----------|--------|-------|
| **Git Repository Tracking** | ✅ Yes | ✅ Yes | Both track source repositories |
| **Automatic Sync** | ✅ Yes | ✅ Yes | Pushpaka: configurable intervals; ArgoCD: continuous |
| **Manual Sync** | ✅ Yes | ✅ Yes | Both support on-demand syncing |
| **Diff Tracking** | ✅ Detailed | ✅ Detailed | File-level change tracking |
| **Change History** | ✅ Complete | ✅ Complete | Full audit trail of syncs |
| **Approval Workflow** | ✅ Yes | ⚠️ Limited | Pushpaka: built-in approval gates |
| **Branch Management** | ✅ Multi-branch | ✅ Single branch | Pushpaka: per-deployment configuration |
| **Rollback Support** | ✅ Yes | ✅ Yes | Automatic and manual rollback |
| **GitOps Principles** | ✅ Full | ✅ Full | Git as source of truth |
| **Change Notifications** | ✅ Yes | ✅ Yes | Slack, email, webhooks |
| **Kubernetes Integration** | ❌ No | ✅ Native | ArgoCD is K8s-native |
| **Docker Support** | ✅ Direct | ⚠️ Via manifests | Pushpaka: native Docker |

## Architecture Comparison

### ArgoCD Architecture
```
┌──────────────────┐
│  Git Repository  │
└────────┬─────────┘
         │
         │ Webhook / Polling
         │
    ┌────▼──────────────────┐
    │  ArgoCD Controller     │
    ├───────────────────────┤
    │ • Git Monitoring       │
    │ • Manifest Processing  │
    │ • Kubernetes Sync      │
    │ • Status Reporting     │
    └────┬──────────────────┘
         │
         │ kubectl apply
         │
    ┌────▼──────────────────┐
    │  Kubernetes Cluster    │
    │  (Deployments, Pods)   │
    └───────────────────────┘
```

### Pushpaka Architecture
```
┌──────────────────┐
│  Git Repository  │
└────────┬─────────┘
         │
         │ Webhook / Polling
         │
    ┌────▼──────────────────┐
    │  GitSync Service       │
    ├───────────────────────┤
    │ • Git Change Tracking  │
    │ • Diff Generation      │
    │ • Approval Workflow    │
    │ • Sync Orchestration   │
    └────┬──────────────────┘
         │
    ┌────┴───────────────────────┐
    │                            │
    ▼                            ▼
┌────────────────┐        ┌──────────────┐
│ Deployment     │        │ Docker/VM    │
│ Service        │        │ Runtime      │
└────────────────┘        └──────────────┘
```

## Integration Patterns

### Pattern 1: Pushpaka + ArgoCD

Use both tools for different layers:

```yaml
# Scenario: Microservices Architecture
Services:
  - Api Service: Managed by Pushpaka (Docker)
  - Frontend: Managed by Pushpaka (Docker)
  - Config: Managed by ArgoCD (K8s ConfigMaps)
  - Infrastructure: Managed by ArgoCD (K8s deployment)

Benefits:
  - Pushpaka: Direct deployment control
  - ArgoCD: Infrastructure as Code
  - Combined: Complete deployment pipeline
```

### Pattern 2: GitOps Workflow

Both support complete GitOps process:

```
1. Developer commits to git
   └─ Triggers webhook
      └─ Both Pushpaka and ArgoCD notified

2. Pushpaka detects change
   ├─ Parses git diff
   ├─ Generates change summary
   └─ Requests approval (if configured)

3. ArgoCD detects change
   ├─ Syncs K8s manifests
   ├─ Updates cluster state
   └─ Runs health checks

4. Both report status
   ├─ Sync history recorded
   ├─ Notifications sent
   └─ Dashboard updated
```

## Key Differences

### 1. Deployment Target

| Pushpaka | ArgoCD |
|----------|--------|
| Docker containers | Kubernetes resources |
| VMs / Servers | Container orchestration |
| Flexible runtime | Kubernetes-only |

### 2. Configuration Model

| Pushpaka | ArgoCD |
|----------|--------|
| Per-deployment config | Git-based manifests |
| API-driven | Code-driven |
| Dynamic updates | Declarative manifests |

### 3. Sync Strategy

| Pushpaka | ArgoCD |
|----------|--------|
| Manual or scheduled | Continuous (default) |
| Approval gates | Policy-based |
| Full control | Automated |

### 4. Change Handling

| Pushpaka | ArgoCD |
|----------|--------|
| Track all changes | Track manifest changes |
| File-level diffs | Resource-level diffs |
| Change notifications | Sync status reports |

## Migration Guide: ArgoCD to Pushpaka

If migrating from ArgoCD:

### Step 1: Inventory Existing Deployments
```bash
# Export ArgoCD applications
kubectl get applications -A -o yaml > argocd-apps.yaml

# Map to Pushpaka deployments
for app in argocd-apps.yaml:
  - Name → Pushpaka project name
  - Repo → Git repository
  - Branch → Git branch
  - Path → Application root path
```

### Step 2: Create Pushpaka Projects
```sh
curl -X POST http://pushpaka/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app",
    "git_repo": "https://github.com/org/repo",
    "description": "Migrated from ArgoCD"
  }'
```

### Step 3: Enable Git Sync
```sh
curl -X POST http://pushpaka/api/v1/deployments/{id}/git/auto-sync \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "require_approval": false,
    "polling_interval": 3600
  }'
```

### Step 4: Configure Webhooks
```yaml
# Update ArgoCD webhook to notify Pushpaka
webhooks:
  - url: http://pushpaka/api/v1/webhooks/github
    events:
      - push
      - pull_request
```

## Hybrid Deployment Strategy

### Scenario: Multi-Tier Application

```
┌─────────────────────────────────────┐
│         Git Repository              │
│  ├─ src/                (app code)  │
│  ├─ helm/               (K8s charts)│
│  ├─ docker/             (Dockerfile)│
│  └─ config/             (app config)│
└────────────┬────────────────────────┘
             │
    ┌────────┴─────────┐
    │                  │
    ▼                  ▼
┌─────────────────┐  ┌──────────────┐
│   Pushpaka      │  │   ArgoCD     │
├─────────────────┤  ├──────────────┤
│ • Build & Test  │  │ • Infrastructure
│ • Deploy App    │  │ • K8s Manifests
│ • Run Services  │  │ • Config Mgmt
└─────────────────┘  └──────────────┘
    │
    ▼
┌─────────────────┐
│  Docker Runtime │
│  (Services A,B) │
└─────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Kubernetes Cluster                 │
│  ├─ Config Server (from ArgoCD)    │
│  ├─ Service Mesh (from ArgoCD)     │
│  └─ Monitoring (from ArgoCD)       │
└─────────────────────────────────────┘
```

## Integration Checklist

- [ ] Set up Git webhook notifications
- [ ] Configure Pushpaka to track repositories
- [ ] Enable auto-sync with appropriate intervals
- [ ] Set up approval workflow if needed
- [ ] Configure notifications (Slack, email)
- [ ] Enable git change tracking
- [ ] Test manual sync workflow
- [ ] Test automatic sync on commit
- [ ] Set up monitoring and alerts
- [ ] Document deployment process for team
- [ ] Configure rollback procedures
- [ ] Set up sync history retention

## Recommendations

### Use Pushpaka When:
✅ Deploying to Docker/VM infrastructure
✅ Need approval gates before deployment
✅ Want detailed change tracking and notifications
✅ Prefer API-driven deployment configuration
✅ Need fine-grained control over sync policy

### Use ArgoCD When:
✅ Already using Kubernetes
✅ Need continuous deployment (always in sync)
✅ Want GitOps with declarative manifests
✅ Need multi-environment management
✅ Infrastructure is primarily K8s-based

### Use Both When:
✅ Hybrid architecture (Docker + K8s)
✅ Different teams managing different layers
✅ Need both app and infrastructure deployment
✅ Want maximum flexibility and control

## API Mapping

### ArgoCD AppProject → Pushpaka Project
```yaml
ArgoCD:
  apiVersion: argoproj.io/v1alpha1
  kind: AppProject
  metadata:
    name: my-project

Pushpaka (equivalent):
  POST /api/v1/projects
  {
    "name": "my-project",
    "git_repo": "https://github.com/org/repo"
  }
```

### ArgoCD Application → Pushpaka Deployment
```yaml
ArgoCD:
  apiVersion: argoproj.io/v1alpha1
  kind: Application
  spec:
    source:
      repoURL: https://github.com/org/repo
      targetRevision: main
      path: ./app

Pushpaka (equivalent):
  POST /api/v1/deployments
  {
    "project_id": "proj_123",
    "branch": "main",
    "commit_sha": "abc123..."
  }
```

## Performance Considerations

### Pushpaka
- **Polling overhead**: Minimal (configurable interval)
- **Git operations**: On-demand cloning
- **Change detection**: Fast (git diff)
- **Sync time**: Depends on deployment size

### ArgoCD
- **Polling overhead**: Continuous reconciliation
- **Manifest processing**: Fast
- **Sync time**: Fast (kubectl apply)
- **K8s API calls**: May impact cluster under load

### Optimization Tips
1. **Set appropriate polling intervals** (300-3600 seconds)
2. **Use ignore paths** to skip unnecessary syncs
3. **Cache recent syncs** to avoid duplicate work
4. **Batch notifications** to reduce alert noise
5. **Archive old sync history** to maintain performance

## Troubleshooting Cross-System Issues

### Issue: ArgoCD Out of Sync, Pushpaka In Sync

**Cause:** ArgoCD manages K8s resources, Pushpaka manages app deployment

**Solution:**
1. Check both deployment targets separately
2. Ensure git repo reflects actual state
3. Run manual sync on both systems
4. Verify deployment configuration in both

### Issue: Conflicts Between Two Sync Systems

**Prevention:**
- Clearly separate concerns (app vs infra)
- Use different git branches for each
- Document which system manages what
- Implement locking to prevent simultaneous syncs

---

**For more details:**
- See `GIT_SYNC_IMPLEMENTATION.md` for Pushpaka GitOps
- See ArgoCD documentation for Kubernetes workflows
- See `API_REFERENCE.md` for Pushpaka API details

**Last Updated:** March 18, 2026
