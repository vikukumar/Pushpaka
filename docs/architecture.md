# Architecture

## Overview

Pushpaka is a distributed, containerized deployment platform. It is designed for self-hosting on a single VPS or a multi-node infrastructure.

```
┌──────────────────────────────────────────────────────────────────┐
│                         Internet                                 │
└──────────────────────────────┬───────────────────────────────────┘
                               │ HTTPS
                    ┌──────────▼──────────┐
                    │    Traefik (v3)      │   Reverse Proxy / TLS
                    │   Port 80 / 443     │   LetsEncrypt Auto-SSL
                    └──┬──────────────┬───┘
                       │              │
              ┌────────▼───┐   ┌──────▼────────┐
              │  Dashboard  │   │  Backend API   │
              │  (Next.js)  │   │  (Go / Gin)   │
              │   :3000     │   │    :8080       │
              └────────────┘   └──────┬────────┘
                                      │
               ┌──────────────────────┼──────────────────────┐
               │                      │                       │
     ┌─────────▼──────┐    ┌──────────▼──────┐   ┌──────────▼──────┐
     │  PostgreSQL     │    │     Redis        │   │  Build Worker   │
     │  (Data store)  │    │   (Job queue)    │   │  (Go Process)   │
     └────────────────┘    └────────────────┘   └──────────┬──────┘
                                                            │
                                           ┌────────────────▼──────────┐
                                           │   Docker Engine             │
                                           │   (Container Runtime)      │
                                           │   Deployed user containers  │
                                           └───────────────────────────┘
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

### Frontend Dashboard (Next.js)
- React Server Components + Client components
- Zustand for state management
- React Query for data fetching
- Real-time log streaming via WebSocket

### Build Worker (Go)
- Polls Redis queue for deployment jobs
- Clones Git repository
- Detects or generates Dockerfile
- Builds Docker image
- Runs container with Traefik labels
- Streams logs to PostgreSQL

### PostgreSQL
- Primary data store
- Schema: users, projects, deployments, domains, environment_variables, deployment_logs

### Redis
- Job queue: `pushpaka:deploy:queue`
- FIFO with BRPOP/LPUSH

## Deployment Flow

```
User → POST /deployments
         │
         └── Create Deployment record (status: queued)
         └── Serialize DeploymentJob → Redis queue

Worker picks up job (BRPOP)
         │
         └── Update status: building
         └── git clone <repo> --branch <branch>
         └── Generate Dockerfile if missing
         └── docker build -t <image_tag> .
         └── docker run -d --network pushpaka-network <labels> <image>
         └── Update status: running  (or failed)
         └── Stream all output → deployment_logs table

Dashboard → WebSocket /logs/:id/stream
         │
         └── Poll deployment_logs every 500ms
         └── Push new entries to browser
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
