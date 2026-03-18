# Pushpaka

> Deploy everywhere. Let AI run point.

Pushpaka is a self-hosted deployment platform for teams that want one system for Git-based delivery, AI-assisted operations, live runtime visibility, and infrastructure control.

It combines:
- multi-platform deployment across Docker, Kubernetes-oriented targets, and direct runtime fallback
- AI monitoring, AI log analysis, AI assistant chat, and RAG-backed operational context
- a built-in editor and web terminal for in-platform debugging and maintenance
- Traefik-powered routing, custom domains, TLS, and production-friendly deployment workflows

Pushpaka is designed for teams that want the speed of modern platform products without giving up control of infrastructure, runtime, or delivery workflows.

## What Pushpaka Does

Pushpaka acts as an internal deployment control plane.

A typical flow looks like this:
1. Connect a Git repository.
2. Configure the project, branch, runtime, and environment variables.
3. Trigger a deployment from the UI or API.
4. Build automatically using framework detection or an existing Dockerfile.
5. Deploy to the chosen target.
6. Inspect logs, health, metrics, AI analysis, editor state, and terminal access.
7. Roll back or re-deploy when needed.

## Core Product Capabilities

### AI Operations
- AI deployment log analysis for failed or unstable deployments
- AI monitoring alerts with configurable monitoring intervals
- AI assistant chat for contextual operational help
- RAG document support for internal runbooks and knowledge injection
- per-user AI configuration and AI usage tracking

### Deployment Platform
- Git-based deployments for public and private repositories
- private repository support with protected token handling
- automatic Dockerfile generation when a custom Dockerfile is not present
- direct runtime fallback when Docker is unavailable
- deployment history and rollback support
- multi-project and multi-user operation

### Runtime And Infrastructure
- Docker-based deployment workflows
- Kubernetes-aware project targets and infrastructure inspection flows
- Traefik-based ingress and routing
- custom domains with automatic TLS provisioning
- system health, readiness, and runtime visibility endpoints
- Prometheus metrics export
- worker statistics and queue-mode visibility

### Operator Experience
- real-time deployment log streaming
- built-in Monaco-based file editor
- source sync into the editor workspace
- web terminal access into running deployments
- project settings for build, runtime, resources, webhooks, and integrations
- dashboard support for AI, infrastructure, domains, notifications, and deployment operations

### Security And Access
- JWT authentication and API key support
- bcrypt password hashing
- secure headers and rate limiting
- configurable CORS
- redaction of sensitive Git credentials in logs and API flows
- GitHub and GitLab OAuth support in the backend

### Integrations
- GitHub and GitLab OAuth handlers
- incoming webhooks for deployment automation
- Slack, Discord, and SMTP notification configuration
- API-first workflows for internal automation and platform orchestration

## Runtime Modes

Pushpaka supports multiple operating modes depending on team size and environment.

### Single Binary Dev Mode
Best for local development and fast evaluation.

- SQLite instead of PostgreSQL
- embedded worker
- in-process queue
- no Docker, Redis, or PostgreSQL requirement to get started

Example:

```bash
cd cmd/pushpaka
go build -o pushpaka .
./pushpaka -dev
```

### All-In-One Mode
Best for simple self-hosted production setups.

- API and worker in one process
- shared runtime behavior
- simpler operational footprint

### Split Mode
Best for production scale and separation of concerns.

- `PUSHPAKA_COMPONENT=api`
- `PUSHPAKA_COMPONENT=worker`
- Redis-backed queue for worker scaling

## Quick Start

### Docker Compose

```bash
git clone https://github.com/vikukumar/Pushpaka
cd Pushpaka
cp .env.example .env
# configure DOMAIN, JWT_SECRET, POSTGRES_PASSWORD, REDIS_PASSWORD, ACME_EMAIL
docker compose up -d --build
```

### Frontend Dashboard

```bash
cd frontend
pnpm install
pnpm dev
```

### Helm

```bash
helm repo add pushpaka https://pushpaka.vikshro.in/helm
helm repo update
helm install pushpaka pushpaka/pushpaka
```

## Main Workflows

### Create And Deploy A Project
1. Create a project in the dashboard.
2. Add repository URL and branch.
3. Configure environment variables and build settings.
4. Choose deployment target if needed.
5. Trigger deployment.
6. Watch real-time logs and deployment state.

### Analyze A Failed Deployment
1. Open a deployment.
2. Review logs and status.
3. Run AI log analysis.
4. Inspect AI-generated summary or remediation hints.
5. Use editor or terminal if operational changes are required.
6. Re-deploy or roll back.

### Edit Source In Platform
1. Open the project editor.
2. Sync the working tree if needed.
3. Browse and edit text files.
4. Save changes.
5. Validate behavior with terminal or deployment actions.

### Manage Routing
1. Add a custom domain to the project.
2. Point DNS to the Pushpaka entrypoint.
3. Wait for TLS provisioning.
4. Validate routing and certificate state.

## API Surface

Pushpaka exposes an API for authentication, projects, deployments, logs, AI operations, domains, environment variables, notifications, webhooks, infrastructure controls, and health endpoints.

Representative endpoints include:
- `/api/v1/auth/*`
- `/api/v1/projects`
- `/api/v1/deployments`
- `/api/v1/logs/:id/stream`
- `/api/v1/ai/*`
- `/api/v1/notifications/*`
- `/api/v1/webhooks/*`
- `/api/v1/containers/*`
- `/api/v1/k8s/*`
- `/api/v1/health`
- `/api/v1/ready`
- `/api/v1/system`
- `/api/v1/metrics`

For deeper reference, see the documents in [`docs/`](docs).

## Repository Structure

```text
cmd/pushpaka/          Combined application entry point
backend/               Go API server and platform services
worker/                Build and deployment worker
frontend/              Next.js operator dashboard
website/               Astro marketing site and documentation
docs/                  Supporting platform and deployment documents
helm/                  Helm chart and release assets
infrastructure/        Routing and infrastructure configuration
branding/              Brand assets
migrations/            Database migrations
```

## Technology Stack

- Go backend and worker runtime
- Gin HTTP framework
- Next.js and React dashboard
- Astro documentation and marketing site
- PostgreSQL and SQLite support
- Redis queue support
- Traefik ingress and TLS automation
- Docker runtime integration
- Kubernetes infrastructure integration
- Monaco editor in the dashboard

## Documentation

Additional repository documents:
- [Documentation Hub](website/src/pages/docs.astro)
- [Architecture](docs/architecture.md)
- [Deployment Guide](docs/deployment.md)
- [Local Development](docs/local-dev.md)
- [Platform Overview](docs/platform-overview.md)
- [API Reference](docs/api.md)
- [Security Policy](SECURITY.md)
- [Contributing Guide](CONTRIBUTING.md)
- [License](LICENSE.md)
- [Copyright](COPYRIGHT.md)

## Website

Product site and documentation:

https://pushpaka.vikshro.in/

## Support

For support and collaboration:
- GitHub Issues for bugs and defects
- GitHub Discussions for usage and product questions
- repository documentation for installation, troubleshooting, and platform workflows

## License

Pushpaka is licensed under the MIT License.

See [LICENSE.md](LICENSE.md).
