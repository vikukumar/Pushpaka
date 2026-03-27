# Deployment Guide

## Production Deployment

### Requirements

- VPS with at least 2 vCPU, 4GB RAM
- Docker + Docker Compose installed
- A domain name with DNS access
- Ports 80 and 443 open

---

## Step 1: Server Setup

```bash
# Ubuntu 24.04 LTS
sudo apt update && sudo apt upgrade -y

# Install Docker (includes Compose v2 plugin)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker --version          # Docker 27+
docker compose version    # Docker Compose v2.x
```

## Step 2: Deploy Pushpaka

```bash
# Clone the repository
git clone https://github.com/vikukumar/pushpaka /opt/pushpaka
cd /opt/pushpaka

# Configure environment
cp .env.example .env
nano .env
```

**Minimum required `.env` settings:**

```env
DOMAIN=pushpaka.example.com
JWT_SECRET=<generate-with-openssl>
POSTGRES_PASSWORD=very-secure-database-password
REDIS_PASSWORD=very-secure-redis-password
ACME_EMAIL=your-email@example.com
APP_ENV=production
BUILD_CLONE_DIR=/tmp/pushpaka-builds
BUILD_DEPLOY_DIR=/deploy/pushpaka
```

```bash
# Generate a strong JWT secret
openssl rand -hex 32
```

## Step 3: DNS Configuration

Point your domain to the server IP:

| Type | Name | Value |
|------|------|-------|
| A | `app.pushpaka.example.com` | `<server-ip>` |
| A | `api.pushpaka.example.com` | `<server-ip>` |
| A | `traefik.pushpaka.example.com` | `<server-ip>` |

Wait for DNS propagation (usually 5–15 minutes).

## Step 4: Launch

```bash
cd /opt/pushpaka

# Build and start all services
docker compose up -d --build

# Check logs
docker compose logs -f

# Migrations run automatically via postgres initdb.d on first launch
```

## Step 5: Verify

- Dashboard: `https://app.pushpaka.example.com`
- API health: `https://api.pushpaka.example.com/api/v1/health`
- System status: `https://api.pushpaka.example.com/api/v1/system`
- Traefik dashboard: `https://traefik.pushpaka.example.com`

## Updating

```bash
cd /opt/pushpaka
git pull origin main
docker compose up -d --build
```

## Backups

```bash
# Database backup
docker compose exec postgres pg_dump -U pushpaka pushpaka > backup_$(date +%Y%m%d).sql

# Restore
cat backup.sql | docker compose exec -T postgres psql -U pushpaka pushpaka
```

## Scaling Workers

```bash
# Horizontal: run 5 separate worker containers
docker compose up -d --scale pushpaka-worker=5
```

Or set `BUILD_WORKERS=5` in `.env` for vertical scaling within a single worker process (default: 3 goroutines).

## Monitoring

Prometheus metrics are available at:
```
https://api.pushpaka.example.com/api/v1/metrics
```

Scrape with Prometheus and visualize in Grafana.
