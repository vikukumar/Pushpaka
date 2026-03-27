# Local Development Guide

## Prerequisites

- **Go 1.25+** (`go version`)
- **Node.js 22+** with `pnpm` (`node --version`, `npm i -g pnpm`)
- **Git** (`git --version`)
- **Docker + Docker Compose** (optional — Pushpaka can run without Docker in dev mode)

---

## Fastest Option: Single Binary Dev Mode

The easiest local setup — **no Docker, Redis, or PostgreSQL required**. Everything runs in one process with SQLite.

```bash
# Build the combined binary
cd cmd/pushpaka
go build -o pushpaka.exe .   # Windows
# go build -o pushpaka .     # Linux/macOS

# Run in dev mode
./pushpaka.exe -dev
# Starts on :8080 with SQLite, embedded worker, in-process queue

# Frontend (separate terminal)
cd frontend
pnpm install
pnpm dev
# Open http://localhost:3000
```

Dev mode auto-sets:
- `DATABASE_DRIVER=sqlite` + `DATABASE_URL=pushpaka-dev.db`
- `REDIS_URL=""` (skipped — in-process channel used instead)
- `JWT_SECRET=dev-secret-change-in-production`
- `APP_ENV=development` (pretty console log output)
- `PORT=8080`

---

## Docker Compose (Full Stack)

```bash
# Clone the repository
git clone https://github.com/vikukumar/pushpaka
cd Pushpaka

# Copy environment variables
cp .env.example .env

# Start all services (dev overrides: exposed ports, debug logging)
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Dashboard: http://localhost:3000
# API:       http://localhost:8080
# Traefik:   http://localhost:8888
```

---

## Running Services Individually

### Database + Redis only

```bash
docker compose up postgres redis
```

### Backend API (Go)

```bash
# From workspace root
cd backend
go mod tidy
# Set required env vars (or use a .env file with direnv / source)
export DATABASE_DRIVER=sqlite
export DATABASE_URL=pushpaka-dev.db
export JWT_SECRET=dev-secret
export APP_ENV=development
go run ./cmd/server
```

Or run the combined binary with `-dev`:

```bash
cd cmd/pushpaka
go run . -dev
```

### Build Worker (standalone)

```bash
cd worker
go mod tidy
export DATABASE_DRIVER=sqlite
export DATABASE_URL=../pushpaka-dev.db
export REDIS_URL=redis://localhost:6379
go run main.go
```

### Frontend

```bash
cd frontend
pnpm install
pnpm dev           # http://localhost:3000
pnpm type-check    # TypeScript validation
pnpm build         # production build
```

---

## Database

### PostgreSQL (production)

Run migrations manually (they also run automatically via `initdb.d`):

```bash
psql $DATABASE_URL -f migrations/001_create_users.sql
psql $DATABASE_URL -f migrations/002_create_projects.sql
psql $DATABASE_URL -f migrations/003_create_deployments.sql
psql $DATABASE_URL -f migrations/004_create_domains.sql
psql $DATABASE_URL -f migrations/005_create_environment_variables.sql
psql $DATABASE_URL -f migrations/006_create_deployment_logs.sql
```

Load seed data:

```bash
psql $DATABASE_URL -f scripts/seed.sql
```

### SQLite (dev)

The database file `pushpaka-dev.db` is created automatically when running with `-dev`. Schema (including `is_private` / `git_token` columns) is applied on first open. Idempotent `ALTER TABLE` migrations handle upgrades.

---

## Environment Variables

See `.env.example` for all available options.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | — | Signing secret (min 32 chars) |
| `DATABASE_URL` | Yes | — | Postgres DSN or SQLite path |
| `DATABASE_DRIVER` | No | `postgres` | `sqlite` or `postgres` |
| `REDIS_URL` | Prod | — | Redis connection URL (empty = in-process) |
| `DOMAIN` | Prod | `localhost` | Base domain for Traefik |
| `ACME_EMAIL` | Prod | — | Email for Let's Encrypt SSL |
| `BUILD_WORKERS` | No | `3` | Concurrent build goroutines |
| `BUILD_CLONE_DIR` | No | `/tmp/pushpaka-builds` | Temp dir for git clones |
| `BUILD_DEPLOY_DIR` | No | `/deploy/pushpaka` | Dir for direct deployments |
| `PUSHPAKA_COMPONENT` | No | `all` | `api` / `worker` / `all` |

---

## Development Tips

- **Combined binary hot notes**: rebuild with `go build -C cmd/pushpaka -o pushpaka.exe .` after Go changes
- **Frontend hot-reload**: built-in with `pnpm dev` — no extra setup needed
- **Backend hot-reload**: install [Air](https://github.com/air-verse/air) and run `air` in `backend/`
- **View deploy queue**: `redis-cli -a <password> LLEN pushpaka:deploy:queue`
- **Tail logs**: `docker compose logs -f pushpaka-api pushpaka-worker`
- **TypeScript check**: `cd frontend && pnpm type-check`
- **Go build check**: `cd backend && go build ./...`
