# Pushpaka Environment & Configuration Guide

## Overview

Pushpaka v1.0.0 now supports comprehensive configuration management with:
- **Multi-environment support** (development, staging, production)
- **PostgreSQL with SSL/TLS certificates**
- **Redis configuration with connection pooling**
- **Environment variable overrides** (12-factor app principle)
- **YAML configuration files** for centralized settings

## Quick Start

### 1. Development Mode (SQLite + In-Process Workers)

For local development without PostgreSQL or Redis:

```bash
# Copy environment template
cp .env.example .env

# Run with in-process queue (no Redis required)
./pushpaka -dev
```

### 2. Production Mode (PostgreSQL + Redis)

For production deployments:

```bash
# Copy production environment
cp .env.production.example .env

# Configure your database and Redis
# Edit .env and set:
# - DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
# - DB_SSL_MODE, DB_SSL_CERT_FILE (for SSL)
# - REDIS_HOST, REDIS_PORT, REDIS_PASSWORD

# Run application
./pushpaka
```

## Environment Variables

### PostgreSQL Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL server hostname |
| `DB_PORT` | 5432 | PostgreSQL server port |
| `DB_USER` | postgres | PostgreSQL username |
| `DB_PASSWORD` | (empty) | PostgreSQL password |
| `DB_NAME` | pushpaka | Database name |
| `DB_SSL_MODE` | disable | SSL mode: disable, prefer, require, verify-ca, verify-full |
| `DB_SSL_CERT_FILE` | (empty) | Path to CA certificate file |
| `DB_MAX_OPEN_CONNS` | 10 (dev), 30 (prod) | Maximum open connections |
| `DB_MAX_IDLE_CONNS` | 3 (dev), 10 (prod) | Maximum idle connections |
| `DB_CONN_MAX_LIFETIME` | 15m (dev), 5m (prod) | Connection lifetime before recycling |

### Redis Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | localhost | Redis server hostname |
| `REDIS_PORT` | 6379 | Redis server port |
| `REDIS_PASSWORD` | (empty) | Redis password (leave empty if not set) |
| `REDIS_DB` | 0 | Redis database number |
| `REDIS_MAX_RETRIES` | 3 | Maximum retry attempts |
| `REDIS_POOL_SIZE` | 5 (dev), 20 (prod) | Connection pool size |
| `REDIS_MIN_IDLE_CONNS` | 1 (dev), 5 (prod) | Minimum idle connections |
| `REDIS_MAX_CONN_AGE` | 10m | Maximum connection age |
| `REDIS_POOL_TIMEOUT` | 4s | Connection pool timeout |
| `REDIS_IDLE_TIMEOUT` | 5m | Idle connection timeout |
| `REDIS_READ_TIMEOUT` | 3s (dev), 5s (prod) | Read operation timeout |
| `REDIS_WRITE_TIMEOUT` | 3s (dev), 5s (prod) | Write operation timeout |

### Application Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | development | development, staging, production |
| `PORT` | 8080 | Application server port |
| `LOG_LEVEL` | info | debug, info, warn, error |
| `BASE_URL` | http://localhost:8080 | Public base URL |
| `CORS_ORIGINS` | http://localhost:3000 | Comma-separated CORS origins |

## PostgreSQL SSL/TLS Setup

### Development (No SSL)
```bash
DB_SSL_MODE=disable
DB_SSL_CERT_FILE=
```

### Production (Require SSL)

#### AWS RDS PostgreSQL
```bash
# Download CA certificate
wget https://truststore.amazonaws.com/global/certificates/rds-ca-bundle.pem -O /etc/ssl/certs/rds-ca-bundle.pem

# Set environment variables
DB_SSL_MODE=require
DB_SSL_CERT_FILE=/etc/ssl/certs/rds-ca-bundle.pem
```

#### Azure Database for PostgreSQL
```bash
# Download CA certificate
wget https://cacerts.digicert.com/DigiCertGlobalRootCA.crt -O /etc/ssl/certs/azure-postgres-ca.pem

# Set environment variables
DB_SSL_MODE=verify-full
DB_SSL_CERT_FILE=/etc/ssl/certs/azure-postgres-ca.pem
```

#### DigitalOcean Managed PostgreSQL
```bash
# Download CA certificate (already trusted on UNIX systems)
DB_SSL_MODE=require
DB_SSL_CERT_FILE=  # Leave empty, uses system CA store
```

#### Self-Hosted PostgreSQL
```bash
# Copy your CA certificate
cp /path/to/your/ca-cert.pem /etc/ssl/certs/postgres-ca.pem

# Set environment variables
DB_SSL_MODE=require
DB_SSL_CERT_FILE=/etc/ssl/certs/postgres-ca.pem
```

## SSL Mode Options

| Mode | Description | Production Ready |
|------|-------------|-----------------|
| `disable` | No encryption | ❌ No |
| `prefer` | Try SSL, fallback to unencrypted | ⚠️ Development only |
| `require` | Require SSL, no host validation | ✅ Yes |
| `verify-ca` | Require SSL with CA validation | ✅ Recommended |
| `verify-full` | Require SSL with CA and hostname validation | ✅ Best |

## YAML Configuration

Advanced users can use `config/config.yaml` for centralized configuration:

### Example Structure

```yaml
development:
  server:
    port: 8080
    log_level: debug
    
  database:
    driver: sqlite
    path: pushpaka-dev.db

production:
  server:
    port: 8080
    log_level: warn
    
  database:
    driver: postgres
    host: ${DB_HOST}
    port: ${DB_PORT}
    user: ${DB_USER}
    password: ${DB_PASSWORD}
    name: ${DB_NAME}
    ssl_mode: require
    ssl_cert_file: ${DB_SSL_CERT_FILE}
    max_open_conns: 30
    max_idle_conns: 10
    
  redis:
    enabled: true
    host: ${REDIS_HOST}
    port: ${REDIS_PORT}
    password: ${REDIS_PASSWORD}
    pool_size: 20
```

## Programmatic Configuration

### Using DatabaseConfig

```go
import "github.com/vikukumar/pushpaka/internal/config"

// Load from environment variables
dbCfg := config.LoadDatabaseConfig("production")

// Build PostgreSQL URL
dsn := dbCfg.BuildPostgresURL()

// Open database with SSL support
db, err := database.NewPostgresWithConfig(dbCfg)
```

### Using RedisConfig

```go
import "github.com/vikukumar/pushpaka/internal/config"

// Load from environment variables
redisCfg := config.LoadRedisConfig("production")

// Build Redis URL
url := redisCfg.BuildRedisURL()

// Open Redis with full configuration
client, err := database.NewRedisWithConfig(redisCfg)
```

## Production Deployment Checklist

- [ ] Set `APP_ENV=production`
- [ ] Configure PostgreSQL with SSL:
  - [ ] Download CA certificate
  - [ ] Set `DB_SSL_MODE=require` or `verify-full`
  - [ ] Set `DB_SSL_CERT_FILE` to certificate path
  - [ ] Test connection: `psql "postgres://user:pass@host/dbname?sslmode=require&sslrootcert=/path/to/cert"`
- [ ] Configure Redis with authentication if needed
- [ ] Set strong JWT_SECRET (use: `openssl rand -hex 32`)
- [ ] Configure CORS_ORIGINS for your domain
- [ ] Set BUILD_CLONE_DIR and BUILD_DEPLOY_DIR to persistent storage
- [ ] Enable backup and recovery procedures
- [ ] Test failover and recovery
- [ ] Set up monitoring and alerting
- [ ] Document your configuration

## Performance Tuning

### PostgreSQL Connection Pool

For workload with N concurrent users:
- `DB_MAX_OPEN_CONNS`: 2-3x of N (e.g., 30-50 for 15-25 users)
- `DB_MAX_IDLE_CONNS`: 20-30% of max open (e.g., 10 for 30 max)
- `DB_CONN_MAX_LIFETIME`: 5-10 minutes (prevents connection staleness)

### Redis Connection Pool

For deployment job throughput:
- `REDIS_POOL_SIZE`: 10-20 for most workloads
- `REDIS_MIN_IDLE_CONNS`: 30% of pool size
- `REDIS_READ_TIMEOUT`: 3-5 seconds
- `REDIS_WRITE_TIMEOUT`: 3-5 seconds

## Troubleshooting

### PostgreSQL Connection Issues

```bash
# Test PostgreSQL connection
psql "postgres://user:password@host:5432/dbname?sslmode=disable"

# With SSL
psql "postgres://user:password@host:5432/dbname?sslmode=require&sslrootcert=/path/cert"

# Check certificate validity
openssl x509 -in /path/to/cert -text -noout
```

### Redis Connection Issues

```bash
# Test Redis connection
redis-cli -h localhost -p 6379 ping

# With authentication
redis-cli -h localhost -p 6379 -a password ping

# Check Redis info
redis-cli info
```

### SSL Certificate Errors

**Error: "certificate verify failed"**
- Check certificate path in `DB_SSL_CERT_FILE`
- Verify certificate PEM format (-----BEGIN CERTIFICATE-----)
- Test with `openssl s_client -connect host:5432 -CAfile cert.pem`

**Error: "no such file or directory"**
- Ensure certificate file path is absolute
- Check file permissions: `chmod 644 /path/to/cert`

## Migration Guide

### From Legacy Configuration

Old style (still supported):
```bash
DATABASE_URL=postgres://user:pass@host/db?sslmode=disable
REDIS_URL=redis://:password@localhost:6379/0
```

New style (recommended):
```bash
DB_HOST=host
DB_USER=user
DB_PASSWORD=pass
DB_NAME=db
DB_SSL_MODE=disable

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=password
REDIS_DB=0
```

## Files Reference

| File | Purpose |
|------|---------|
| `.env.example` | Template for development environment |
| `.env.production.example` | Template for production environment |
| `config/config.yaml.example` | Example YAML configuration |
| `backend/internal/config/config.go` | Main configuration loader |
| `backend/internal/config/database.go` | Database configuration classes |
| `backend/internal/database/postgres.go` | PostgreSQL connection functions |
| `backend/internal/database/redis.go` | Redis connection functions |
