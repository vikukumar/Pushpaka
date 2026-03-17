# Architecture

## Overview

Pushpaka is a distributed, containerized deployment platform designed for self-hosting on a single VPS or multi-node infrastructure. It also runs as a **single binary in dev mode** (no external dependencies).

```
                         Internet
                            |  HTTPS/WSS
               +------------v-------------+
               |      Traefik v3          |  Reverse Proxy / TLS
               |   Port 80 / 443          |  Let's Encrypt Auto-SSL
               +------+----------+--------+
                      |          |
           +----------v---+  +---v-----------+
           | Dashboard    |  | Backend API   |
           | Next.js 16   |  | Go 1.25/      |
           | React 19     |  | Gin 1.12      |
           | :3000        |  | :8080         |
           +--------------+  +----+---------+
                                  |
               +------------------+-------------------+
               |                  |                   |
     +---------v------+  +--------v-------+  +--------v-------+
     | PostgreSQL 16  |  |  Redis 7       |  | Build Worker   |
     | (Data store)   |  | (Job queue)    |  | (Go process)   |
     +----------------+  +----------------+  +--------+-------+
                                                      |
                                        +-------------v-----------+
                                        |     Docker Engine       |
                                        |  git -> build -> run    |
                                        |  or direct in-place     |
                                        +-------------------------+

  Dev mode (single binary, -dev flag):
  pushpaka -dev  =>  API + embedded worker + SQLite + in-process queue
```

## Services

### Traefik
- Accepts all HTTP/HTTPS traffic
- Routes to backend API and dashboard
- Handles TLS via Let's Encrypt
- Labels on deployed containers for automatic routing

### Backend API (Go/Gin)
- RESTful API + WebSocket (for log streaming)
- Clean architecture: handlers → services → repositories
- JWT authentication + API key support
- Prometheus metrics at `/api/v1/metrics`

### Frontend Dashboard (Next.js 16 / React 19)
- App Router with Server + Client Components
- Zustand 5 for state management
- TanStack Query 5 for data fetching and caching
- Real-time log streaming via WebSocket (JWT-authenticated)
- Custom CSS-variable-based dark/light theme system
- Tailwind CSS 4 with responsive, animated UI

### Build Worker (Go 1.25)
- Consumes jobs from Redis queue (production) or in-process channel (dev mode)
- Clones Git repositories including private repos via embedded PAT in HTTPS URL (token never logged)
- Auto-detects package manager: npm / yarn / pnpm / bun (PATH fallback)
- Detects or generates an optimized Dockerfile per framework
- Falls back to direct in-place deployment if Docker is unavailable
- Builds Docker image and runs container with Traefik labels
- Streams structured logs (level + stream) to database in real time
- Tracks worker/job lifecycle counters exposed on `/api/v1/system`

### PostgreSQL
- Primary data store
- Schema: users, projects, deployments, domains, environment_variables, deployment_logs

### Redis 7 (production only)
- Job queue: `pushpaka:deploy:queue`
- FIFO with BRPOP/LPUSH
- Skipped in dev mode — replaced by in-process Go channel

### In-process Queue (dev mode)
- Buffered Go channel in `queue.InProcess`
- Exposes `TotalWorkers()` / `ActiveJobs()` counters
- Zero external dependencies

## Deployment Flow

```
User → POST /api/v1/deployments
         |
         +-- Create Deployment record (status: queued)
         +-- Serialize DeploymentJob JSON -> queue
              (Redis LPUSH  -or-  in-process channel Push)

Worker dequeues job (BRPOP / channel recv)
         |
         +-- Update status: building
         +-- git clone <repo> [token embedded if private]
         +-- Auto-detect package manager (npm/yarn/pnpm/bun)
         +-- Generate Dockerfile if not present
         +-- docker available?
             YES: docker build -t <image_tag> .
                  docker run -d --network pushpaka-network <traefik-labels>
             NO:  install deps && run start command in-place (BUILD_DEPLOY_DIR)
         +-- Update status: running (or failed)
         +-- Stream all output -> deployment_logs table

Dashboard -> WebSocket /api/v1/logs/:id/stream (JWT auth)
         |
         +-- Poll deployment_logs every 500ms
         +-- Push new entries to browser in real time
```

## Security

- Passwords hashed with bcrypt (cost 10)
- JWT HS256 with configurable expiry
- API keys as UUID v4
- Rate limiting via middleware
- Secure headers (HSTS, CSP, X-Frame-Options, etc.)
- CORS restricted to configured origins
- Environment variable values never returned via API
- Docker socket mounted read-only in Traefik, RW only in worker
