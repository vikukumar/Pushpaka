# Pushpaka API Documentation

## Base URL

```
https://api.your-domain.com/api/v1
```

For local development:
```
http://localhost:8080/api/v1
```

---

## Authentication

All protected endpoints require:

```
Authorization: Bearer <JWT_TOKEN>
```

Tokens are obtained via `/auth/login` or `/auth/register`.

---

## Endpoints

### Auth

#### POST /auth/register
Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "name": "Jane Doe",
  "password": "securepassword"
}
```

**Response `201`:**
```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Jane Doe",
    "role": "user",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

---

#### POST /auth/login
Authenticate and receive a JWT token.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response `200`:**
```json
{
  "token": "eyJhbGci...",
  "user": { ... }
}
```

---

### Projects

#### POST /projects
Create a new project.

**Request:**
```json
{
  "name": "my-app",
  "repo_url": "https://github.com/user/repo",
  "branch": "main",
  "build_command": "npm run build",
  "start_command": "npm start",
  "port": 3000,
  "framework": "nextjs"
}
```

**Response `201`:** Project object

---

#### GET /projects
List all projects for the authenticated user.

**Response `200`:**
```json
{
  "data": [ { ...project }, ... ]
}
```

---

#### GET /projects/:id
Get a single project.

---

#### DELETE /projects/:id
Delete a project and all its deployments.

---

### Deployments

#### POST /deployments
Trigger a new deployment.

**Request:**
```json
{
  "project_id": "uuid",
  "branch": "main",
  "commit_sha": "abc1234"
}
```

**Response `201`:** Deployment object with `status: "queued"`

---

#### GET /deployments
List all deployments. Supports pagination via `?limit=20&offset=0`.

---

#### GET /deployments/:id
Get a specific deployment.

---

#### POST /deployments/:id/rollback
Trigger a rollback to a previous deployment.

---

### Logs

#### GET /logs/:deployment_id
Get all logs for a deployment.

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "deployment_id": "uuid",
      "level": "info",
      "stream": "system",
      "message": "Build started",
      "created_at": "..."
    }
  ]
}
```

#### WebSocket GET /logs/:deployment_id/stream
Upgrade to WebSocket for real-time log streaming.

```
ws://api.localhost/api/v1/logs/{id}/stream?token=<JWT>
```

Each message is a JSON `DeploymentLog` object.

---

### Domains

#### POST /domains
Add a custom domain.

**Request:**
```json
{
  "project_id": "uuid",
  "domain": "app.yourdomain.com"
}
```

---

#### GET /domains
List all custom domains for the user.

---

#### DELETE /domains/:id
Remove a custom domain.

---

### Environment Variables

#### POST /env
Set or update an environment variable.

**Request:**
```json
{
  "project_id": "uuid",
  "key": "DATABASE_URL",
  "value": "postgres://..."
}
```

> Note: Values are never returned in responses for security.

---

#### GET /env?project_id=:id
List all env var keys (values are masked).

---

#### DELETE /env
Delete an environment variable.

**Request:**
```json
{
  "project_id": "uuid",
  "key": "DATABASE_URL"
}
```

---

### Health & Metrics

#### GET /health
Check system health including database and Redis.

#### GET /ready
Readiness probe for Kubernetes/Docker.

#### GET /metrics
Prometheus metrics endpoint.

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request (validation failed) |
| 401 | Unauthorized |
| 404 | Not Found |
| 409 | Conflict (duplicate) |
| 429 | Too Many Requests (rate limited) |
| 500 | Internal Server Error |
| 503 | Service Unavailable |
