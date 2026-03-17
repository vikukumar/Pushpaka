# Platform Overview

## What is Pushpaka?

Pushpaka is a **production-grade self-hosted cloud deployment platform** that lets developers deploy applications directly from Git repositories — similar to how Vercel, Render, and Railway work, but entirely self-hosted and under your control.

---

## Core Concepts

### Projects

A **Project** connects a Git repository to the Pushpaka platform. Each project has:

- A Git repository URL (public or private)
- A target branch (default: `main`)
- Build and start commands
- An exposed port
- Environment variables
- Custom domains
- Private flag + Personal Access Token (PAT) for private repos

Projects can be created, updated (name, branch, commands, PAT, private flag), and deleted from the dashboard.

### Deployments

A **Deployment** represents one run of the build and deploy pipeline for a project. Each deployment:

1. Clones the repository at a specific commit
2. Builds a Docker image
3. Runs the container on the platform
4. Exposes it via Traefik

Deployments are immutable — every push creates a new deployment.

### Workers

Background workers consume jobs from the queue and:

- Clone repositories (including private repos via PAT — token redacted from all logs)
- Auto-detect package manager (`npm` / `yarn` / `pnpm` / `bun`) with PATH fallback
- Generate optimized Dockerfile if not present in the repo
- Build Docker images (or fall back to direct in-place deployment when Docker is unavailable)
- Deploy containers with Traefik routing labels
- Stream structured logs (level + stream) to the database in real time

In **dev mode** (`pushpaka -dev`), workers run embedded in the API process using an in-process channel queue. Worker count and active job stats are exposed via `/api/v1/system`.

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
| Next.js / React / Vue / Node | `package.json` present |
| Go | `go.mod` present |
| Python | `requirements.txt` or `pyproject.toml` present |
| Docker | `Dockerfile` already present (used as-is) |
| Any | Falls through to a generic build container |

Package manager is also auto-detected (`package.json` lock files) and falls back to `npm` if no lock file is found or the detected PM is not in PATH.

You can always override by providing a `Dockerfile` in your repository root.

---

## Security Model

- All API endpoints require JWT authentication
- Environment variable values are write-only via the API
- TLS is handled by Traefik + Let's Encrypt
- Deployed containers run with no privileged access
- The Docker socket is only accessible to the worker

---

## Theming

The dashboard supports full dark and light mode via a custom CSS-variable-based theme system:

- Preference is stored in `localStorage` (key: `pushpaka_theme`)
- Falls back to system `prefers-color-scheme` on first visit
- Animated toggle pill in the header
- `.dark` / `.light` class toggled on `<html>`, matched by CSS `var(--*)` color tokens

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
