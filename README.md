<div align="center">

<img src="branding/logo.svg" alt="Pushpaka Logo" width="320" />

# Pushpaka

### *Carry your code to the cloud effortlessly.*

[![Version](https://img.shields.io/badge/version-v1.0.0-6366f1?style=flat-square)](https://github.com/yourusername/pushpaka)
[![Go](https://img.shields.io/badge/backend-Go%201.22-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Next.js](https://img.shields.io/badge/frontend-Next.js%2014-black?style=flat-square&logo=next.js)](https://nextjs.org)
[![Docker](https://img.shields.io/badge/infra-Docker-2496ED?style=flat-square&logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)

**Pushpaka** is a production-grade self-hosted cloud deployment platform — deploy applications from any Git repository with automated container builds, real-time logs, custom domains, and Traefik-powered routing.

[Quick Start](#quick-start) · [Features](#features) · [Architecture](#architecture) · [API](#api) · [Roadmap](#roadmap)

</div>

---

## What is Pushpaka?

Pushpaka brings the Vercel/Render/Railway experience to your own infrastructure. It orchestrates the full deployment pipeline:

1. **Connect** a Git repository
2. **Trigger** a deployment (manually or via webhook)
3. **Build** a Docker image automatically
4. **Deploy** the container with zero downtime
5. **Route** traffic via Traefik + optional custom domains
6. **Monitor** with real-time log streaming and health checks

---

## Architecture

```
┌─────────────┐        ┌───────────────┐       ┌───────────────┐
│  Dashboard  │◄──────►│   Backend API  │◄─────►│  PostgreSQL   │
│  (Next.js)  │  HTTPS │   (Go/Gin)    │       │  (Data store) │
└─────────────┘        └───────┬───────┘       └───────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   Traefik (v3)      │
                    │  Reverse Proxy+SSL  │
                    └──────────▲──────────┘
                               │
                    ┌──────────┴──────────┐
                    │  Redis (Job Queue)  │
                    └──────────▲──────────┘
                               │
                    ┌──────────┴──────────┐
                    │   Build Worker      │
                    │  (Go Process)       │
                    │  git → docker build │
                    │  docker run + route │
                    └─────────────────────┘
```

---

## Features

### Platform
- 🚀 **One-click Git deployments** — any public or private repo
- 🐳 **Automatic Dockerization** — detects Next.js, React, Go, Python, and more
- ♻️ **Rollback support** — redeploy any previous deployment instantly
- 🔀 **Multi-project** — unlimited projects per user
- 👥 **Multi-user** — team-ready with role-based access

### Infrastructure
- 🔀 **Traefik Reverse Proxy** — automatic TLS, routing, and load balancing
- 🔐 **Let's Encrypt SSL** — free, automatic, and renewing
- 📊 **Prometheus metrics** — export to Grafana
- ❤️ **Health checks** — `/health` and `/ready` endpoints

### Developer Experience
- 📡 **Real-time logs** — WebSocket streaming during builds
- 🌍 **Custom domains** — map any domain to any project
- 🔑 **Environment variables** — secure, write-only storage
- 🌓 **Dark/light mode** — polished dashboard UI

### Security
- 🔒 **JWT + API key authentication**
- 🔑 **bcrypt password hashing**
- 🛡️ **Secure headers** (HSTS, CSP, X-Frame-Options)
- 🚦 **Rate limiting** on all endpoints
- 🌐 **Configurable CORS**

---

## Quick Start

### Docker Compose (Recommended)

```bash
# Clone
git clone https://github.com/yourusername/pushpaka
cd pushpaka

# Configure
cp .env.example .env
# Edit .env: set DOMAIN, JWT_SECRET, POSTGRES_PASSWORD, REDIS_PASSWORD

# Launch
docker-compose up -d --build

# Open dashboard
open http://localhost:3000
```

**Default demo credentials** (after running `psql $DB -f scripts/seed.sql`):
- Email: `demo@pushpaka.app`
- Password: `Demo@1234`

---

## Project Structure

```
pushpaka/
├── backend/                  # Go API server
│   ├── cmd/server/main.go   # Entrypoint
│   └── internal/
│       ├── handlers/         # HTTP handlers
│       ├── services/         # Business logic
│       ├── repositories/     # Database layer
│       ├── models/           # Data models
│       ├── middleware/        # Auth, logging, security
│       ├── config/           # Configuration
│       └── router/           # Route definitions
│
├── worker/                   # Build & deploy workers
│   └── internal/worker/
│       └── build_worker.go   # Full pipeline: clone → build → run
│
├── frontend/                 # Next.js 14 dashboard
│   └── app/
│       ├── dashboard/        # Main app shell
│       ├── login/            # Auth pages
│       └── register/
│
├── migrations/               # SQL migrations (001–006)
├── infrastructure/           # Traefik dynamic config
├── branding/                 # Logo, favicon, OG image
├── scripts/                  # Seed data
├── docs/                     # Full documentation
└── docker-compose.yml        # Production stack
```

---

## API

Full documentation: [docs/api.md](docs/api.md)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register user |
| POST | `/auth/login` | Login |
| POST | `/projects` | Create project |
| GET | `/projects` | List projects |
| POST | `/deployments` | Trigger deployment |
| GET | `/deployments/:id` | Get deployment |
| POST | `/deployments/:id/rollback` | Rollback |
| GET | `/logs/:id` | Get deployment logs |
| WS | `/logs/:id/stream` | Stream logs live |
| POST | `/domains` | Add custom domain |
| POST | `/env` | Set env variable |
| GET | `/metrics` | Prometheus metrics |
| GET | `/health` | Health check |

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture.md](docs/architecture.md) | System architecture and design |
| [docs/api.md](docs/api.md) | Complete API reference |
| [docs/local-dev.md](docs/local-dev.md) | Local development setup |
| [docs/deployment.md](docs/deployment.md) | Production deployment guide |
| [docs/platform-overview.md](docs/platform-overview.md) | Platform concepts |

---

## Configuration

Key environment variables (see [`.env.example`](.env.example)):

| Variable | Default | Description |
|----------|---------|-------------|
| `DOMAIN` | `localhost` | Base domain |
| `JWT_SECRET` | — | **Required**: JWT signing secret |
| `POSTGRES_PASSWORD` | — | **Required**: Database password |
| `REDIS_PASSWORD` | — | **Required**: Redis password |
| `BUILD_WORKERS` | `3` | Concurrent build workers |
| `ACME_EMAIL` | — | Let's Encrypt contact email |

---

## Roadmap — v1.1.0

- [ ] GitHub / GitLab OAuth (one-click repo connect)
- [ ] Webhook auto-deploy on `git push`
- [ ] Pull Request preview deployments
- [ ] Blue-green zero-downtime deployments
- [ ] Docker Swarm multi-node support
- [ ] CPU/memory resource limits per project
- [ ] Slack / Discord / email notifications
- [ ] Web terminal (exec into containers)
- [ ] Audit log viewer in dashboard
- [ ] Build caching for faster deployments

---

## License

MIT © 2026 Pushpaka Contributors

---

<div align="center">
  <sub>Built with ❤️ — Pushpaka v1.0.0</sub>
</div>
