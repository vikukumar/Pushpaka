# Pushpaka API Reference

Complete API documentation for Pushpaka deployment system.

## Base URL

```
https://api.pushpaka.example.com/api/v1
```

## Authentication

All requests require authentication via API key:

```
Authorization: Bearer {API_KEY}
```

Get your API key from the dashboard: Settings → API Keys

## Response Format

All responses are JSON:

```json
{
  "success": true,
  "data": { /* response data */ },
  "error": null,
  "timestamp": "2026-03-18T10:00:00Z",
  "request_id": "req_abc123"
}
```

## Error Handling

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid project ID",
    "details": {
      "field": "project_id",
      "reason": "not found"
    }
  }
}
```

### Common Error Codes

| Code | HTTP | Meaning |
|------|------|---------|
| `UNAUTHORIZED` | 401 | Invalid API key |
| `FORBIDDEN` | 403 | No permission for resource |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `ALREADY_EXISTS` | 409 | Resource already exists |
| `RATE_LIMITED` | 429 | Too many requests |
| `SERVER_ERROR` | 500 | Internal server error |

## API Endpoints

### Projects

#### List Projects

```
GET /projects

Query Parameters:
  - limit: number (default: 20, max: 100)
  - offset: number (default: 0)
  - filter: string (search by name)

Response:
{
  "data": {
    "projects": [
      {
        "id": "proj_123abc",
        "name": "My App",
        "description": "Production app",
        "status": "active",
        "created_at": "2026-01-15T08:00:00Z",
        "deployments_count": 42
      }
    ],
    "total": 1,
    "limit": 20,
    "offset": 0
  }
}

Status Codes:
  200 - Success
  401 - Unauthorized
  429 - Rate limited
```

#### Get Project

```
GET /projects/{projectId}

Response:
{
  "data": {
    "id": "proj_123abc",
    "name": "My App",
    "description": "Production app",
    "status": "active",
    "owner": "john@example.com",
    "created_at": "2026-01-15T08:00:00Z",
    "updated_at": "2026-03-18T10:00:00Z",
    "plan": "professional",
    "deployment_limit": 1000,
    "storage_limit": 100,  // GB
    "team": {
      "id": "team_456def",
      "name": "Engineering",
      "members_count": 5
    }
  }
}

Status Codes:
  200 - Success
  404 - Project not found
```

#### Create Project

```
POST /projects

Request:
{
  "name": "New App",
  "description": "My application",
  "team_id": "team_456def"  // optional
}

Response:
{
  "data": {
    "id": "proj_123abc",
    "name": "New App",
    "status": "creating"
  }
}

Status Codes:
  201 - Created
  400 - Invalid data
  409 - Already exists
```

#### Update Project

```
PATCH /projects/{projectId}

Request:
{
  "name": "Updated Name",
  "description": "Updated description"
}

Response:
{
  "data": {
    "id": "proj_123abc",
    "name": "Updated Name",
    "updated_at": "2026-03-18T10:00:00Z"
  }
}

Status Codes:
  200 - Success
  404 - Not found
```

#### Delete Project

```
DELETE /projects/{projectId}

Response: 204 No Content

Status Codes:
  204 - Deleted
  404 - Not found
```

### Deployments

#### List Deployments

```
GET /projects/{projectId}/deployments

Query Parameters:
  - status: pending|deploying|deployed|failed|rolled_back
  - limit: number (default: 20)
  - offset: number (default: 0)

Response:
{
  "data": {
    "deployments": [
      {
        "id": "dep_123xyz",
        "version": "1.2.3",
        "commit": "abc123def",
        "status": "deployed",
        "deployed_at": "2026-03-18T09:00:00Z",
        "deployed_by": "john@example.com",
        "duration": 45
      }
    ],
    "total": 42
  }
}

Status Codes:
  200 - Success
  404 - Project not found
```

#### Get Deployment

```
GET /projects/{projectId}/deployments/{deploymentId}

Response:
{
  "data": {
    "id": "dep_123xyz",
    "version": "1.2.3",
    "commit": "abc123def",
    "author": "john@example.com",
    "message": "feat: Add feature X",
    "status": "deployed",
    "tests": {
      "total": 45,
      "passed": 45,
      "failed": 0,
      "duration": 120
    },
    "approval": {
      "status": "approved",
      "approved_by": "dev-lead@example.com",
      "approved_at": "2026-03-18T08:50:00Z"
    },
    "deployment": {
      "started_at": "2026-03-18T08:55:00Z",
      "completed_at": "2026-03-18T09:00:00Z",
      "duration": 45,
      "status": "success"
    },
    "health": {
      "status": "healthy",
      "last_check": "2026-03-18T10:00:00Z",
      "checks": {
        "api": "passing",
        "database": "passing",
        "cache": "passing"
      }
    }
  }
}

Status Codes:
  200 - Success
  404 - Not found
```

#### Create Deployment

```
POST /projects/{projectId}/deployments

Request:
{
  "version": "1.2.3",
  "commit": "abc123def",
  "source": "github|manual|webhook",  // optional
  "environment": "staging|production"  // optional
}

Response:
{
  "data": {
    "id": "dep_123xyz",
    "version": "1.2.3",
    "status": "testing",
    "created_at": "2026-03-18T09:00:00Z"
  }
}

Status Codes:
  201 - Created
  400 - Invalid data
  429 - Rate limited
```

#### Approve Deployment

```
POST /projects/{projectId}/deployments/{deploymentId}/approve

Request:
{
  "approved": true,
  "reason": "Tested in staging"
}

Response:
{
  "data": {
    "id": "dep_123xyz",
    "status": "deploying",
    "approval": {
      "approved_by": "john@example.com",
      "approved_at": "2026-03-18T10:00:00Z"
    }
  }
}

Status Codes:
  200 - Success
  404 - Not found
  409 - Already approved/rejected
```

#### Rollback Deployment

```
POST /projects/{projectId}/deployments/{deploymentId}/rollback

Request:
{
  "reason": "Critical bug in latest version"
}

Response:
{
  "data": {
    "id": "dep_123xyz",
    "status": "rolling_back",
    "rollback": {
      "initiated_by": "john@example.com",
      "reason": "Critical bug in latest version",
      "target_version": "1.2.2"
    }
  }
}

Status Codes:
  200 - Success
  404 - Not found
  409 - Cannot rollback
```

### GitHub Integration

#### Connect GitHub Repository

```
POST /projects/{projectId}/github/connect

Request:
{
  "repository": "owner/repo",
  "branch": "main",
  "github_token": "ghp_xxxxx",
  "auto_approve": false,
  "require_passing_checks": true
}

Response:
{
  "data": {
    "id": "github_abc",
    "repository": "owner/repo",
    "branch": "main",
    "connected_at": "2026-03-18T10:00:00Z",
    "webhook_url": "https://pushpaka.example.com/webhooks/github/abc123",
    "status": "connected"
  }
}

Status Codes:
  201 - Connected
  400 - Invalid token or repository
```

#### Disconnect GitHub Repository

```
DELETE /projects/{projectId}/github/integration/{integrationId}

Response: 204 No Content

Status Codes:
  204 - Disconnected
  404 - Not found
```

#### Get GitHub Integration Status

```
GET /projects/{projectId}/github/integration/{integrationId}

Response:
{
  "data": {
    "id": "github_abc",
    "repository": "owner/repo",
    "branch": "main",
    "status": "connected",
    "last_sync": "2026-03-18T10:00:00Z",
    "webhook_status": "active",
    "auto_approve": false,
    "recent_pushes": [
      {
        "commit": "abc123def",
        "author": "john@example.com",
        "message": "feat: Add feature",
        "pushed_at": "2026-03-18T09:00:00Z"
      }
    ]
  }
}

Status Codes:
  200 - Success
  404 - Not found
```

### Health Checks

#### Get Health Status

```
GET /projects/{projectId}/health

Response:
{
  "data": {
    "status": "healthy",
    "timestamp": "2026-03-18T10:00:00Z",
    "last_deployment": "1.2.3",
    "uptime_percentage": 99.95,
    "checks": {
      "api": {
        "status": "passing",
        "response_time": 45,
        "last_check": "2026-03-18T10:00:00Z"
      },
      "database": {
        "status": "passing",
        "response_time": 120,
        "last_check": "2026-03-18T10:00:00Z"
      },
      "cache": {
        "status": "passing",
        "response_time": 5,
        "last_check": "2026-03-18T10:00:00Z"
      }
    },
    "metrics": {
      "cpu_usage": 35,
      "memory_usage": 62,
      "error_rate": 0.05,
      "request_rate": 1250
    }
  }
}

Status Codes:
  200 - Success
  404 - Not found
```

### Webhooks

#### Create Webhook

```
POST /projects/{projectId}/webhooks

Request:
{
  "url": "https://example.com/webhook",
  "events": ["deployment_started", "deployment_completed", "health_check_failed"],
  "secret": "webhook_secret_key"
}

Response:
{
  "data": {
    "id": "hook_123",
    "url": "https://example.com/webhook",
    "events": ["deployment_started", "deployment_completed"],
    "created_at": "2026-03-18T10:00:00Z",
    "status": "active"
  }
}

Status Codes:
  201 - Created
  400 - Invalid URL
```

#### Webhook Event

Pushpaka sends webhook events as POST requests:

```json
{
  "event": "deployment_completed",
  "timestamp": "2026-03-18T10:00:00Z",
  "data": {
    "deployment_id": "dep_123xyz",
    "version": "1.2.3",
    "status": "success",
    "duration": 45
  },
  "signature": "sha256=..."  // HMAC-SHA256(payload, secret)
}
```

#### Webhook Events

```
deployment_started
deployment_completed
deployment_failed
rollback_started
rollback_completed
health_check_passed
health_check_failed
approval_pending
approval_approved
approval_rejected
```

### Telemetry

#### Get Metrics

```
GET /projects/{projectId}/metrics

Query Parameters:
  - metric: cpu|memory|disk|network|requests|errors
  - from: ISO 8601 timestamp
  - to: ISO 8601 timestamp

Response:
{
  "data": {
    "metric": "cpu",
    "unit": "percent",
    "data_points": [
      {
        "timestamp": "2026-03-18T10:00:00Z",
        "value": 35.2
      }
    ]
  }
}

Status Codes:
  200 - Success
  400 - Invalid metric
```

### Settings

#### Get Project Settings

```
GET /projects/{projectId}/settings

Response:
{
  "data": {
    "deployment": {
      "timeout": 600,
      "max_concurrent": 2,
      "auto_rollback": true
    },
    "approval": {
      "required": true,
      "auto_approve": false,
      "timeout": 3600,
      "reviewers": ["dev-lead@example.com"]
    },
    "notifications": {
      "slack_enabled": true,
      "email_enabled": true,
      "channels": ["#deployments"]
    }
  }
}

Status Codes:
  200 - Success
```

#### Update Project Settings

```
PATCH /projects/{projectId}/settings

Request:
{
  "deployment": {
    "auto_rollback": true,
    "timeout": 600
  },
  "approval": {
    "required": true,
    "timeout": 3600
  }
}

Response:
{
  "data": {
    /* Updated settings */
  }
}

Status Codes:
  200 - Success
  400 - Invalid settings
```

## Rate Limiting

Rate limits per API key:
- 1000 requests per hour
- 100 concurrent requests

Response headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1711016400
```

## Pagination

Use `limit` and `offset` for pagination:

```
GET /projects?limit=20&offset=0

Response:
{
  "data": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "items": [...]
  }
}
```

## Versioning

API version in URL path: `/api/v1`

Breaking changes in new major versions. Current: v1

## SDKs

Official SDKs available:
- Go: `github.com/pushpaka/sdk-go`
- JavaScript: `npm install @pushpaka/sdk`
- Python: `pip install pushpaka`

Example (Go):
```go
import "github.com/pushpaka/sdk-go/pushpaka"

client := pushpaka.New("api_key")
project, err := client.GetProject("proj_123abc")
```

Example (JavaScript):
```javascript
const { PushpakaClient } = require("@pushpaka/sdk");

const client = new PushpakaClient("api_key");
const project = await client.getProject("proj_123abc");
```

Example (Python):
```python
from pushpaka import Client

client = Client("api_key")
project = client.get_project("proj_123abc")
```

---

**Last Updated:** March 18, 2026
**API Version:** v1.0.0
