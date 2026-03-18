# Quick Reference - Environment & Configuration

## 🚀 Quick Start

### Development
```bash
cp .env.example .env
./pushpaka -dev    # SQLite + in-process workers
```

### Production
```bash
# Download certificate (AWS example)
wget https://truststore.amazonaws.com/rds-ca-bundle.pem

# Copy template
cp .env.production.example .env

# Configure (edit .env with your values)
DB_HOST=your-db.example.com
DB_SSL_MODE=require
DB_SSL_CERT_FILE=/path/to/cert.pem
REDIS_HOST=your-redis.example.com

# Run
./pushpaka
```

## 📋 Essential Environment Variables

### PostgreSQL (Required for Production)
```bash
DB_HOST=localhost              # Server hostname
DB_PORT=5432                   # Server port
DB_USER=postgres               # Username
DB_PASSWORD=secure_password    # Password (use secret manager!)
DB_NAME=pushpaka              # Database name
```

### PostgreSQL SSL (Optional, Recommended for Production)
```bash
DB_SSL_MODE=require            # require / verify-ca / verify-full
DB_SSL_CERT_FILE=/path/to/ca.pem  # CA certificate path
```

### Redis (Optional, Required for Multi-Process)
```bash
REDIS_HOST=localhost           # Server hostname
REDIS_PORT=6379               # Server port
REDIS_PASSWORD=               # Password (if needed)
REDIS_DB=0                    # Database number
```

### Application
```bash
APP_ENV=production            # development / staging / production
PORT=8080                     # Server port
JWT_SECRET=<random>           # Generate: openssl rand -hex 32
BASE_URL=https://pushpaka.example.com
CORS_ORIGINS=https://app.example.com
```

## 🔒 SSL Modes Explained

| Mode | Description | Use Case |
|------|-------------|----------|
| `disable` | No encryption | Development only |
| `prefer` | Try SSL, fallback | Staging/transition |
| `require` | Require SSL, no validation | Production (basic) |
| `verify-ca` | Require SSL + CA validation | Production (secure) |
| `verify-full` | Require SSL + hostname validation | Production (best) |

## ☁️ Cloud Provider Quick Links

### AWS RDS
```bash
# Download CA
wget https://truststore.amazonaws.com/rds-ca-bundle.pem

# Environment
DB_SSL_MODE=require
DB_SSL_CERT_FILE=/path/to/rds-ca-bundle.pem
```

### Azure PostgreSQL
```bash
# Download CA
wget https://cacerts.digicert.com/DigiCertGlobalRootCA.crt

# Environment
DB_SSL_MODE=verify-full
DB_SSL_CERT_FILE=/path/to/azure-ca.crt
```

### Google Cloud SQL
```bash
# Use Cloud SQL Proxy instead of direct SSL
DB_HOST=127.0.0.1    # Proxy listens here
DB_PORT=5432
DB_SSL_MODE=disable  # Proxy handles SSL
```

### DigitalOcean
```bash
# Standard CA certificates work
DB_SSL_MODE=require
DB_SSL_CERT_FILE=    # Leave empty, uses system CA
```

## 🧪 Testing Connections

### PostgreSQL
```bash
# Without SSL
psql -h localhost -p 5432 -U postgres -d pushpaka

# With SSL
psql "postgres://user:pass@host:5432/db?sslmode=require&sslrootcert=/path/cert"
```

### Redis
```bash
# Basic
redis-cli -h localhost -p 6379 ping

# With password
redis-cli -h localhost -p 6379 -a password ping
```

## 📊 Connection Pool Sizes

### Development (SQLite/small load)
```
PostgreSQL: max=10, idle=3
Redis: pool=5, min_idle=1
```

### Staging (Medium load)
```
PostgreSQL: max=20, idle=5
Redis: pool=10, min_idle=2
```

### Production (High load)
```
PostgreSQL: max=30, idle=10  # Adjust for your workload
Redis: pool=20, min_idle=5   # Adjust for job throughput
```

## 🔧 Custom Configuration

### Using Environment Variables Only
```bash
# PostgreSQL
export DB_HOST=db.example.com
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secret
export DB_NAME=pushpaka
export DB_SSL_MODE=require
export DB_SSL_CERT_FILE=/etc/ssl/certs/ca.pem

# Redis
export REDIS_HOST=redis.example.com
export REDIS_PORT=6379
export REDIS_PASSWORD=redis_secret

# Run
./pushpaka
```

### Using .env File
```bash
# Copy template
cp .env.production.example .env

# Edit with your values
# Then run - automatically loads from .env
./pushpaka
```

### Using YAML Configuration
```bash
# Copy template
cp config/config.yaml.example config/config.yaml

# Edit YAML with your settings
# Set mode via APP_ENV
export APP_ENV=production
./pushpaka
```

## ✅ Pre-Deployment Checklist

- [ ] Database connectivity: `psql` connects successfully
- [ ] SSL certificate: File exists and is readable
- [ ] Redis connectivity: `redis-cli ping` returns PONG
- [ ] All variables set: `env | grep -E "DB_|REDIS_|JWT_"`
- [ ] JWT secret generated: `openssl rand -hex 32`
- [ ] Log level appropriate: `LOG_LEVEL=warn` for production
- [ ] Base URL correct: `BASE_URL=your-domain`

## 🐛 Troubleshooting Quick Guide

| Problem | Check | Fix |
|---------|-------|-----|
| "Connection refused" | Host/port correct? | Verify `DB_HOST` and firewall |
| "SSL certificate error" | Certificate exists? | Check `DB_SSL_CERT_FILE` path |
| "Database unreachable" | Network route? | Test with `psql` directly |
| "Pool exhausted" | Max connections set? | Increase `DB_MAX_OPEN_CONNS` |
| "Redis timeout" | Redis running? | Verify `REDIS_HOST:REDIS_PORT` |
| "SSL mode not recognized" | Typo in mode? | Use: require/verify-ca/verify-full |

## 📚 Full Documentation

- **Setup Guide**: `docs/CONFIG_GUIDE.md`
- **Architecture**: `docs/CONFIG_ARCHITECTURE.md`  
- **Cloud Providers**: `CLOUD_PROVIDER_CONFIG.md`
- **Code Examples**: `CONFIG_INTEGRATION_EXAMPLES.md`

## 🎯 Next Steps

1. Copy appropriate `.env` file (`.env.example` for dev, `.env.production.example` for prod)
2. Set database and Redis connection details
3. For SSL: Download CA certificate and set `DB_SSL_CERT_FILE`
4. Test connections with `psql` and `redis-cli`
5. Generate JWT secret: `openssl rand -hex 32`
6. Run application and verify startup logs
7. Test API endpoints and verify database/Redis access

---

**For detailed documentation, see:** `docs/CONFIG_GUIDE.md`
