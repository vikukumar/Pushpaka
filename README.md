<div align="center">

<img src="branding/logo.svg" alt="Pushpaka Logo" width="320" />

# Pushpaka

### *Carry your code to the cloud effortlessly.*

[![Version](https://img.shields.io/badge/version-v1.0.0-6366f1?style=flat-square)](https://github.com/vikukumar/Pushpaka/releases)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-16.1-black?style=flat-square&logo=next.js)](https://nextjs.org)
[![React](https://img.shields.io/badge/React-19.2-61DAFB?style=flat-square&logo=react)](https://react.dev)
[![Tailwind](https://img.shields.io/badge/Tailwind-4.2-38BDF8?style=flat-square&logo=tailwindcss)](https://tailwindcss.com)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker)](https://docker.com)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Helm-326ce5?style=flat-square&logo=kubernetes)](helm/pushpaka)
[![License](https://img.shields.io/badge/license-MIT-22c55e?style=flat-square)](LICENSE)

**Pushpaka** is a production-grade self-hosted cloud deployment platform — deploy applications from any Git repository (public or private) with automated container builds, real-time logs, custom domains, dark/light theming, and Traefik-powered routing. Choose your deployment mode: single binary for development, Docker Compose for small deployments, or **Kubernetes with Helm charts** for enterprise scale.

🌐 **[Website](https://vikukumar.github.io/Pushpaka/)** - Beautiful documentation, installation guides, feature showcase, and release tracker  
📦 **[Helm Charts](https://vikukumar.github.io/Pushpaka/helm/)** - Production-ready Kubernetes deployment  
🤖 **[AI Chatbot](https://vikukumar.github.io/Pushpaka/)** - Real-time platform support right on the website  
📋 **[Releases](https://github.com/vikukumar/Pushpaka/releases)** - Version history, changelogs, upgrade guides

[Features](#features) · [Quick Start](#quick-start) · [Kubernetes Deployment](#kubernetes-deployment) · [Dev Mode](#dev-mode-single-binary) · [Architecture](#architecture) · [API](#api) · [Chatbot Support](#chatbot-support) · [Roadmap](ROADMAP.md)

</div>

---

## What is Pushpaka?

Pushpaka brings the Vercel/Render/Railway experience to your own infrastructure. It orchestrates the full deployment pipeline:

1. **Connect** a Git repository (public or private with PAT)
2. **Trigger** a deployment (manually, via API, or via webhook)
3. **Build** — auto-detects framework and generates a Dockerfile, or uses your own
4. **Deploy** the container via Docker or in-place process deployment
5. **Route** traffic via Traefik + optional custom domains + auto-SSL with Let's Encrypt
6. **Monitor** with real-time WebSocket log streaming and live system status

**Deployment Modes:**
- 🏃 **Dev Mode:** Single binary with SQLite — `pushpaka -dev`
- 🐳 **Docker Compose:** Full stack for small teams — `docker compose up`
- ☸️ **Kubernetes:** Enterprise-grade Helm charts — `helm install pushpaka oci://...`

---

## Features

### Platform
- 🚀 **One-click Git deployments** — public repos and private repos with Personal Access Token
- 🔒 **Private repository support** — PAT stored securely, never returned via API, redacted from logs
- 🐳 **Automatic Dockerization** — detects 15+ frameworks; generates optimized Dockerfile
- 🚫 **Docker-free direct deploy** — falls back to in-place process deployment
- ♻️ **Rollback support** — redeploy any previous deployment instantly
- 🔀 **Multi-project** — unlimited projects per user
- 👥 **Multi-user** — team-ready with role-based access (admin/user)
- 🗑️ **Project management** — create, update settings, and delete projects
- 📦 **Package detection** — auto-detects npm, pip, go mod, maven, etc.
- ⚡ **Rate limiting** — configurable request and build rate limits
- 🔄 **User configuration priority** — override deployments with custom settings

### Infrastructure
- 🔀 **Traefik v3 Reverse Proxy** — automatic TLS, routing, and load balancing
- 🔐 **Let's Encrypt SSL** — free, automatic, and renewing
- 📊 **Prometheus metrics** — export at `/api/v1/metrics` for Grafana
- ❤️ **Health checks** — `/health`, `/ready`, and `/system` status endpoints
- 🔧 **Worker stats** — live worker count, active jobs, idle count
- ☸️ **Kubernetes Ready** — production-grade Helm charts with autoscaling (2-5 replicas for API, 3-10 for workers)
- 🔢 **Horizontal Pod Autoscaling** — CPU/memory-based auto-scaling for all components
- 💾 **Persistent Storage** — PostgreSQL (20Gi), Redis (10Gi) with replication

### Developer Experience
- 📡 **Real-time logs** — WebSocket streaming during builds with level/stream filtering
- 🌍 **Custom domains** — map any domain to any project with auto-SSL
- 🔑 **Environment variables** — secure write-only storage
- 🌓 **Dark/light mode** — CSS-variable theming with localStorage persistence and system preference detection
- 📦 **Single binary dev mode** — `pushpaka -dev` starts everything with SQLite
- 🧰 **Package manager auto-detect** — auto-detects `npm` / `yarn` / `pnpm` / `bun`
- 🤖 **AI Chatbot** — OpenRouter-powered platform support 24/7 on website
- 📖 **Comprehensive documentation** — modern website with guides and tutorials

### Security
- 🔒 **JWT v5 + API key authentication**
- 🔑 **bcrypt password hashing** (cost 10)
- 🛡️ **Secure headers** (HSTS, CSP, X-Frame-Options, X-Content-Type-Options)
- 🚦 **Rate limiting** on all endpoints
- 🌐 **Configurable CORS**
- 🙈 **Git token redaction** — PAT never appears in logs
- 🔐 **Kubernetes RBAC** — role-based access control in K8s deployments
- 🚫 **Network policies** — Ingress/Egress rules for pod-to-pod communication

---

## Quick Start

### 1. Docker Compose (Development/Small Setup)

```bash
# Clone repository
git clone https://github.com/vikukumar/Pushpaka
cd Pushpaka

# Configure environment
cp .env.example .env
# Edit .env: set DOMAIN, JWT_SECRET, POSTGRES_PASSWORD, REDIS_PASSWORD, ACME_EMAIL

# Start services
docker compose up -d --build

# Access dashboard
open https://app.YOUR_DOMAIN
```

**Minimum `.env`:**
```env
DOMAIN=pushpaka.example.com
JWT_SECRET=$(openssl rand -hex 32)
POSTGRES_PASSWORD=your-super-secret-password
REDIS_PASSWORD=your-redis-secret
ACME_EMAIL=admin@example.com
```

### 2. Kubernetes with Helm (Production/Enterprise)

```bash
# Add Helm repository
helm repo add pushpaka https://vikukumar.github.io/Pushpaka/helm
helm repo update

# Create namespace
kubectl create namespace pushpaka

# Create values file
cat > values.yaml << EOF
domain: pushpaka.example.com
image:
  tag: v1.0.0
api:
  replicas: 3
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "1000m"
worker:
  replicas: 3
  resources:
    requests:
      memory: "1Gi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "2000m"
postgresql:
  enabled: true
  storage: 50Gi
redis:
  enabled: true
  storage: 20Gi
ssl:
  enabled: true
  email: admin@example.com
EOF

# Install Helm chart
helm install pushpaka pushpaka/pushpaka \
  --namespace pushpaka \
  --values values.yaml

# Verify deployment
kubectl get pods -n pushpaka
kubectl get svc -n pushpaka
```

For detailed Kubernetes setup, see the [Helm Installation Guide](https://vikukumar.github.io/Pushpaka/helm-install).

### 3. Dev Mode (Single Binary)

```bash
# Build backend
cd backend
go build -o pushpaka

# Run in dev mode (no database required)
./pushpaka -dev

# Build and run frontend (in another terminal)
cd frontend
npm install && npm run dev

# Access
open http://localhost:3000
```

---

## Architecture

```
                         Internet
                            |  HTTPS/WSS (Let's Encrypt)
               +------------v-------------+
               |      Traefik v3          |  Reverse Proxy / Load Balancer
               |   Port 80 / 443          |  Automatic TLS
               +------+----------+--------+
                      |          |
           +----------v---+  +---v-----------+
           |  Dashboard   |  |  Backend API  |
           |  (Next.js 16)|  |  (Go 1.25/   |
           |   :3000      |  |   Gin 1.12)  |
           +--------------+  +----+---------+
                                  |
               +------------------+-------------------+
               |                  |                   |
     +---------v------+  +--------v-------+  +--------v-------+
     |  PostgreSQL 17 |  |   Redis 8      |  |  Build Worker  |
     |  (Persistent)  |  |  (Job queue)   |  |  (Go process)  |
     +----------------+  +----------------+  +--------+-------+
                                                      |
                                        +-------------v-----------+
                                        |     Docker Engine       |
                                        |  git → build → run      |
                                        |  or direct deploy       |
                                        +-------------------------+

  Dev Mode (Single Binary with -dev flag):
  pushpaka -dev  →  API + embedded workers + SQLite (no external DB)

  All-in-One Mode (Default):
  pushpaka  →  API + embedded workers + Redis job queue

  Distributed Mode (Kubernetes):
  PUSHPAKA_COMPONENT=api     →  API replicas        (2-5)
  PUSHPAKA_COMPONENT=worker  →  Worker replicas     (3-10)
  +  PostgreSQL (20Gi)  +  Redis (10Gi)  +  Traefik (3 replicas)
```

---

## Kubernetes Deployment

Pushpaka provides production-ready Helm charts configured for enterprise deployments.

### What's Included in Helm Chart

**Components:**
- API server (2-5 replicas with HPA)
- Dashboard (2-3 replicas)
- Build worker (3-10 replicas with Docker socket mounting)
- PostgreSQL 17 (20Gi persistent volume)
- Redis 8 (10Gi persistent volume with replication)
- Traefik v3 (3 replicas, LoadBalancer)
- Cert-Manager (with Let's Encrypt)
- Prometheus & Grafana (optional monitoring)

**Features:**
- ✅ **Horizontal Pod Autoscaling** — CPU/memory-based scaling
- ✅ **Pod Disruption Budgets** — maintain service during upgrades
- ✅ **Network Policies** — secure inter-pod communication
- ✅ **RBAC** — service accounts and role bindings
- ✅ **Health Checks** — liveness and readiness probes
- ✅ **Resource Limits** — requests and limits for all pods
- ✅ **Pod Anti-affinity** — distribute pods across nodes
- ✅ **Persistent Storage** — databases and cache with backups

### Quick Helm Install

```bash
# Add repository
helm repo add pushpaka https://vikukumar.github.io/Pushpaka/helm
helm repo update

# Install with default values
helm install pushpaka pushpaka/pushpaka \
  --namespace pushpaka --create-namespace

# Verify
helm status pushpaka -n pushpaka
kubectl get pods -n pushpaka

# Upgrade
helm upgrade pushpaka pushpaka/pushpaka \
  --namespace pushpaka \
  --values custom-values.yaml

# Rollback
helm rollback pushpaka -n pushpaka
```

### Configuration

**Key values.yaml options:**

```yaml
# Domain and SSL
domain: pushpaka.example.com
acmeEmail: admin@example.com

# Component replicas
api:
  replicas: 3
dashboard:
  replicas: 2
worker:
  replicas: 5

# Auto-scaling (HPA)
autoscaling:
  enabled: true
  targetCPU: 80
  targetMemory: 85

# Storage
postgresql:
  storage: 50Gi
redis:
  storage: 20Gi

# Resource limits
resources:
  api:
    requests: { memory: "512Mi", cpu: "250m" }
    limits: { memory: "1Gi", cpu: "1000m" }
  worker:
    requests: { memory: "1Gi", cpu: "500m" }
    limits: { memory: "2Gi", cpu: "2000m" }

# Monitoring
monitoring:
  enabled: true
  prometheus:
    retention: 15d
  grafana:
    dashboards: true
```

For detailed configuration options and troubleshooting, see [Helm Installation Guide](https://vikukumar.github.io/Pushpaka/helm-install).

---

## Release Management

Pushpaka uses semantic versioning (v1.0.0, v1.1.0, etc.). Each release includes:

- **CHANGELOG.md** — Detailed changes and new features
- **FEATURES.md** — Feature list by category
- **COMPONENTS.md** — Component-level changes
- **Downloads** — Links to binaries, Docker images, Helm charts

**Current Release:** v1.0.0 (March 17, 2026)  
**Next Release:** v1.1.0 (Q2 2026) — Scheduled deployments, notifications, webhooks

View all releases and changelogs: [Releases](https://github.com/vikukumar/Pushpaka/releases)

---

## Chatbot Support

Pushpaka includes an **AI-powered chatbot** (powered by OpenRouter GPT-4 Turbo) available on the website. The chatbot provides 24/7 support for:

- 🚀 Installation and setup (Linux, macOS, Windows, Docker, Kubernetes)
- ⚙️ Configuration and troubleshooting
- 📖 Feature usage and best practices
- 🔧 API documentation and examples
- 🐛 Common issues and solutions
- 📋 Release notes and upgrade paths

**Using the Chatbot:**
1. Visit [Pushpaka Website](https://vikukumar.github.io/Pushpaka/)
2. Click the 💬 button in the bottom-right corner
3. Ask your question
4. Get instant AI-powered response

### Enable Chatbot in Your Deployment

To enable the chatbot, set the `OPENROUTER_API_KEY` GitHub Secret:

1. Go to [OpenRouter](https://openrouter.ai) and create an API key
2. In your GitHub repository: **Settings** → **Secrets and variables** → **Actions**
3. Create new repository secret `OPENROUTER_API_KEY` with your OpenRouter API key
4. The website will automatically use the chatbot feature

---

## Dev Mode (Single Binary)

Perfect for local development—no external database or Redis required!

```bash
# Build
cd backend && go build -o pushpaka

# Run
./pushpaka -dev

# In another terminal, start frontend
cd frontend && npm run dev

# Access
open http://localhost:3000
```

**Dev mode includes:**
- ✅ Embedded SQLite database
- ✅ In-process job queue (no Redis)
- ✅ All features enabled
- ✅ Perfect for development and testing

---

## API Overview

### Authentication

All API requests require a JWT token or API key.

```bash
# Get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password"
  }'

# Use token in subsequent requests
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/projects
```

### Core Endpoints

**Projects:**
```bash
GET    /api/v1/projects              # List projects
POST   /api/v1/projects              # Create project
GET    /api/v1/projects/{id}         # Get project details
PUT    /api/v1/projects/{id}         # Update project
DELETE /api/v1/projects/{id}         # Delete project
```

**Deployments:**
```bash
GET    /api/v1/projects/{id}/deployments      # List deployments
POST   /api/v1/projects/{id}/deployments      # Create deployment
GET    /api/v1/projects/{id}/deployments/{id} # Get deployment
POST   /api/v1/projects/{id}/deployments/{id}/rollback  # Rollback
```

**Logs:**
```bash
GET    /api/v1/projects/{id}/deployments/{id}/logs      # Get logs
WS     /api/v1/projects/{id}/deployments/{id}/logs/stream  # Real-time stream
```

**Metrics:**
```bash
GET    /api/v1/metrics               # Prometheus metrics
GET    /api/v1/health                # Health status
GET    /api/v1/system                # System information
```

For complete API documentation, see [API Docs](https://vikukumar.github.io/Pushpaka/api).

---

## Configuration

### Environment Variables

```bash
# Core
DOMAIN=pushpaka.example.com
JWT_SECRET=$(openssl rand -hex 32)
PORT=8080

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=pushpaka
POSTGRES_USER=pushpaka
POSTGRES_PASSWORD=very-secret

# Redis
REDIS_URL=redis://localhost:6379

# Git Integration
GIT_PROVIDER=github  # or gitlab, gitea

# TLS/SSL
ACME_EMAIL=admin@example.com
ACME_PROD=true

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# OpenRouter (for Chatbot)
OPENROUTER_API_KEY=<your-key>
OPENROUTER_MODEL=openai/gpt-4-turbo

# Feature Flags
ENABLE_PRIVATE_REPOS=true
ENABLE_CUSTOM_DOMAINS=true
ENABLE_DARK_MODE=true
```

For detailed configuration, see [Configuration Guide](https://vikukumar.github.io/Pushpaka/docs).

---

## Project Structure

```
Pushpaka/
├── backend/                 # Go backend (API, workers, deployment logic)
│   ├── main.go
│   ├── models/
│   ├── handlers/
│   ├── services/
│   ├── middleware/
│   └── go.mod
├── frontend/               # Next.js dashboard (React 19)
│   ├── app/
│   ├── components/
│   ├── lib/
│   ├── styles/
│   └── package.json
├── website/               # Astro documentation site
│   ├── src/
│   ├── pages/
│   ├── components/
│   └── astro.config.mjs
├── helm/                  # Kubernetes Helm charts
│   └── pushpaka/
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── templates/  (API, Dashboard, Worker, Ingress, etc.)
│       └── README.md
├── releases/              # Release documentation
│   ├── v1.0.0/
│   └── v1.1.0/
├── docker-compose.yml     # Single command production deployment
├── Dockerfile             # Multi-stage builds
├── .github/workflows/     # GitHub Actions (CI/CD, Helm release)
└── README.md
```

---

## Contributing

We welcome contributions! Areas where help is needed:

- 🐛 Bug fixes and improvements
- 📚 Documentation and tutorials
- 🔧 New framework detection
- 🎨 UI/UX enhancements
- 🧪 Test coverage
- 🌍 Translations and localization
- 🔌 Plugin development
- 📦 Deployment templates

### Development Setup

```bash
# Backend
cd backend
go mod tidy
go run main.go

# Frontend
cd frontend
npm install && npm run dev

# Website
cd website/src
npm install && npm run dev
```

See [Contributing Guide](CONTRIBUTING.md) for detailed instructions.

---

## Roadmap

**Current:** v1.0.0 ✅
- Core deployment pipeline
- Git integration
- Auto Dockerization
- Real-time logs
- Dark/light mode
- Kubernetes Helm charts
- AI Chatbot (OpenRouter)

**Planned:** v1.1.0 (Q2 2026)
- Scheduled deployments
- Email notifications
- Webhooks (pre/post-deployment)
- Mobile-friendly PWA
- Advanced API key management

**Future:** v1.2.0+ (Q3+ 2026)
- Multi-region deployments
- Database migration tools
- Advanced RBAC
- Git SSH authentication
- Cost analytics
- AI deployment optimization

[Full Roadmap](ROADMAP.md)

---

## Community

- 💬 **Discussions:** [GitHub Discussions](https://github.com/vikukumar/Pushpaka/discussions)
- 🐛 **Issues:** [GitHub Issues](https://github.com/vikukumar/Pushpaka/issues)
- 📧 **Email:** vikukumar@example.com
- 🌍 **Website:** [pushpaka.io](https://vikukumar.github.io/Pushpaka/)
- 💖 **Sponsor:** [GitHub Sponsors](https://github.com/sponsors/vikukumar)

---

## License

MIT License — see [LICENSE](LICENSE) file for details.

---

## Acknowledgments

Built with:
- 🔴 **Go** - Backend API and workers
- ⚛️ **React 19** - Interactive UI
- 🎨 **Tailwind CSS** - Beautiful styling
- 🔄 **Traefik v3** - Reverse proxy and routing
- 🐘 **PostgreSQL 17** - Data persistence
- 🚀 **Redis 8** - Job queue and caching
- ☸️ **Kubernetes** - Container orchestration
- 🌐 **Astro** - Documentation website
- 🤖 **OpenRouter** - AI chatbot

---

<div align="center">

**Made with ❤️ by [Viku Kumar](https://github.com/vikukumar)**

[⭐ Star on GitHub](https://github.com/vikukumar/Pushpaka) · [📧 Contact](mailto:vikukumar@example.com) · [🌍 Website](https://vikukumar.github.io/Pushpaka/)

</div>
