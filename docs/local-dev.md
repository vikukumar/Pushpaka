# Local Development Guide

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker + Docker Compose
- Git

## Quick Start (Docker Compose)

```bash
# Clone the repository
git clone https://github.com/vikukumar/Pushpaka
cd pushpaka

# Copy environment variables
cp .env.example .env

# Start all services (dev mode)
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Dashboard: http://localhost:3000
# API:       http://localhost:8080
# Traefik:   http://localhost:8888
```

## Running Services Individually

### Database + Redis only

```bash
docker-compose up postgres redis
```

### Backend API

```bash
cd backend
cp ../.env.example ../.env
export $(cat ../.env | xargs)

# Download dependencies
go mod tidy

# Run
go run ./cmd/server
```

### Workers

```bash
cd worker
go mod tidy
go run main.go
```

### Frontend

```bash
cd frontend
npm install
cp .env.local.example .env.local
npm run dev
```

## Database

Run migrations manually:

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

Demo credentials: `demo@pushpaka.app` / `Demo@1234`

## Environment Variables

See `.env.example` for all available configuration options.

Critical variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `JWT_SECRET` | Yes | Secret for signing JWTs (min 32 chars) |
| `DATABASE_URL` | Yes | PostgreSQL DSN |
| `REDIS_URL` | Yes | Redis connection URL |
| `DOMAIN` | Production | Your domain (e.g. `pushpaka.example.com`) |
| `ACME_EMAIL` | Production | Email for Let's Encrypt SSL certs |

## Development Tips

- Backend hot-reload: install [Air](https://github.com/cosmtrek/air) and run `air` in `backend/`
- Frontend hot-reload: built-in with `npm run dev`
- View Redis queue: `redis-cli -a <password> LLEN pushpaka:deploy:queue`
- Tail logs: `docker-compose logs -f pushpaka-api pushpaka-worker`
