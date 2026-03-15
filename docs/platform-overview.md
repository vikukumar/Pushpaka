# Platform Overview

## What is Pushpaka?

Pushpaka is a **production-grade self-hosted cloud deployment platform** that lets developers deploy applications directly from Git repositories — similar to how Vercel, Render, and Railway work, but entirely self-hosted and under your control.

---

## Core Concepts

### Projects

A **Project** connects a Git repository to the Pushpaka platform. Each project has:

- A Git repository URL
- A target branch (default: `main`)
- Build and start commands
- An exposed port
- Environment variables
- Custom domains

### Deployments

A **Deployment** represents one run of the build and deploy pipeline for a project. Each deployment:

1. Clones the repository at a specific commit
2. Builds a Docker image
3. Runs the container on the platform
4. Exposes it via Traefik

Deployments are immutable — every push creates a new deployment.

### Workers

Background workers consume jobs from the Redis queue and:
- Clone repositories
- Build Docker images
- Deploy containers
- Stream logs back

### Environment Variables

Sensitive configuration is stored as environment variables, injected at container startup. Values are never returned via the API — only key names are visible.

### Custom Domains

Any domain can be mapped to a project. Traefik automatically routes traffic. DNS verification ensures ownership.

---

## Deployment States

| State | Meaning |
|-------|---------|
| `queued` | Job is in the Redis queue waiting for a worker |
| `building` | Worker is building the Docker image |
| `running` | Container is live and serving traffic |
| `failed` | Build or deployment failed — check logs |
| `stopped` | Container stopped or was replaced |

---

## Framework Support

Pushpaka auto-detects your framework and generates an appropriate Dockerfile:

| Framework | Detection |
|-----------|-----------|
| Next.js / React / Vue | `package.json` present |
| Go | `go.mod` present |
| Python | `requirements.txt` present |
| Docker | `Dockerfile` already present |
| Any | Falls through to generic |

You can override by providing a `Dockerfile` in your repository root.

---

## Security Model

- All API endpoints require JWT authentication
- Environment variable values are write-only via the API
- TLS is handled by Traefik + Let's Encrypt
- Deployed containers run with no privileged access
- The Docker socket is only accessible to the worker

---

## v1.1.0 Roadmap

- GitHub/GitLab OAuth integration
- Webhook auto-deploy on push events
- Preview deployments for pull requests
- Blue-green deployment strategy
- Horizontal scaling via Docker Swarm
- Container resource limits (CPU/memory)
- Slack / Discord webhook notifications
- Web terminal for running commands in containers
- Audit log viewer in dashboard
