# Cloud Provider Configuration Examples

This file contains ready-to-use environment variable configurations for popular cloud providers.

## AWS RDS PostgreSQL + ElastiCache Redis

```bash
# .env (AWS deployment)
APP_ENV=production
PORT=8080
LOG_LEVEL=info

# RDS PostgreSQL
DB_HOST=mydb.c9akciq32.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_strong_password
DB_NAME=pushpaka_prod
DB_SSL_MODE=require
DB_SSL_CERT_FILE=/app/certs/rds-ca-bundle.pem
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=10

# ElastiCache Redis
REDIS_HOST=myredis.abc123.ng.0001.use1.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_auth_token
REDIS_DB=0
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=5

# Security
JWT_SECRET=$(openssl rand -hex 32)
BASE_URL=https://pushpaka.yourdomain.com
CORS_ORIGINS=https://app.yourdomain.com

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_app_id
GITHUB_CLIENT_SECRET=your_github_app_secret
```

### Setup Steps for AWS

1. **RDS PostgreSQL**
   ```bash
   # Download RDS CA certificate
   wget https://truststore.amazonaws.com/global/certificates/rds-ca-bundle.pem

   # Copy to your container/server
   mkdir -p /app/certs
   cp rds-ca-bundle.pem /app/certs/
   ```

2. **ElastiCache Redis**
   - Create AUTH token in ElastiCache console
   - Use the endpoint provided by AWS
   - Ensure security group allows inbound from application

3. **Deployment**
   ```bash
   docker build -t pushpaka:latest .
   docker run -p 8080:8080 \
     --env-file .env \
     -v /app/certs:/app/certs:ro \
     pushpaka:latest
   ```

## Azure Database for PostgreSQL + Azure Cache for Redis

```bash
# .env (Azure deployment)
APP_ENV=production
PORT=8080
LOG_LEVEL=info

# Azure PostgreSQL
DB_HOST=myserver.postgres.database.azure.com
DB_PORT=5432
DB_USER=dbadmin@myserver
DB_PASSWORD=your_strong_password
DB_NAME=pushpaka_prod
DB_SSL_MODE=verify-full
DB_SSL_CERT_FILE=/app/certs/azure-postgres-ca.pem
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=10

# Azure Cache for Redis (Premium with non-SSL disabled)
REDIS_HOST=myredis.redis.cache.windows.net
REDIS_PORT=6380  # Note: Azure uses port 6380 for SSL
REDIS_PASSWORD=your_azure_redis_key
REDIS_DB=0
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=5

# Security
JWT_SECRET=$(openssl rand -hex 32)
BASE_URL=https://pushpaka.yourdomain.com
CORS_ORIGINS=https://app.yourdomain.com

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_app_id
GITHUB_CLIENT_SECRET=your_github_app_secret
```

### Setup Steps for Azure

1. **Azure Database for PostgreSQL**
   ```bash
   # Get DigiCert Global Root CA
   wget https://cacerts.digicert.com/DigiCertGlobalRootCA.crt \
     -O /app/certs/azure-postgres-ca.pem

   # Verify SSL connection
   psql -h myserver.postgres.database.azure.com \
        -U dbadmin@myserver \
        -d postgres \
        -p 5432 \
        --set=sslmode=require \
        --set=sslrootcert=/app/certs/azure-postgres-ca.pem
   ```

2. **Azure Cache for Redis**
   - Enable "Non-SSL port" disabled (enforce SSL)
   - Use port 6380 instead of 6379
   - Copy Primary or Secondary key for authentication

## Google Cloud SQL PostgreSQL + Cloud Memorystore Redis

```bash
# .env (Google Cloud deployment)
APP_ENV=production
PORT=8080
LOG_LEVEL=info

# Cloud SQL PostgreSQL with Cloud SQL Proxy
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=pushpaka_user
DB_PASSWORD=your_strong_password
DB_NAME=pushpaka_prod
DB_SSL_MODE=disable  # Cloud SQL Proxy handles SSL
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=10

# Cloud Memorystore Redis (IP-based access within VPC)
REDIS_HOST=10.0.0.3  # Internal IP
REDIS_PORT=6379
REDIS_PASSWORD=""    # AUTH disabled for internal access
REDIS_DB=0
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=5

# Security
JWT_SECRET=$(openssl rand -hex 32)
BASE_URL=https://pushpaka.yourdomain.com
CORS_ORIGINS=https://app.yourdomain.com

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_app_id
GITHUB_CLIENT_SECRET=your_github_app_secret
```

### Setup Steps for Google Cloud

1. **Cloud SQL Proxy**
   ```bash
   # Download Cloud SQL Proxy
   curl https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 \
     -o cloud_sql_proxy
   chmod +x cloud_sql_proxy

   # Run proxy (connects through UNIX socket)
   ./cloud_sql_proxy -instances=PROJECT:REGION:INSTANCE=tcp:5432 \
     -credential_file=/path/to/service-account-key.json &
   ```

2. **Cloud Memorystore**
   - Use UNIX socket or internal IP within VPC
   - AUTH is optional for internal access
   - Configure firewall rules to allow access from application

## DigitalOcean App Platform + Managed Databases

```bash
# .env (DigitalOcean deployment)
APP_ENV=production
PORT=8080
LOG_LEVEL=info

# DigitalOcean Managed PostgreSQL
DB_HOST=db-postgresql-nyc1-12345-do-user-123456-0.a.db.ondigitalocean.com
DB_PORT=25060
DB_USER=doadmin
DB_PASSWORD=your_strong_password
DB_NAME=defaultdb
DB_SSL_MODE=require
DB_SSL_CERT_FILE=  # DigitalOcean uses standard system certs
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=10

# DigitalOcean Managed Redis
REDIS_HOST=redis-nyc1-12345-do-user-123456-0.a.db.ondigitalocean.com
REDIS_PORT=25061
REDIS_PASSWORD=your_redis_password
REDIS_DB=0
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=5

# Security
JWT_SECRET=$(openssl rand -hex 32)
BASE_URL=https://pushpaka.yourdomain.com
CORS_ORIGINS=https://app.yourdomain.com

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_app_id
GITHUB_CLIENT_SECRET=your_github_app_secret
```

### Setup Steps for DigitalOcean

1. **App Platform Deployment**
   - Create database resources in DigitalOcean console
   - Copy connection strings and credentials
   - No certificate files needed (uses system CA bundle)

2. **Network Configuration**
   - Enable "Restrict Inbound" in database settings
   - Add App Platform app to trusted sources

## Heroku PostgreSQL + Redis

```bash
# .env (Heroku deployment - auto-injected via DATABASE_URL and REDIS_URL)
APP_ENV=production
PORT=$PORT  # Heroku injects PORT
LOG_LEVEL=info

# Heroku PostgreSQL (via standard DATABASE_URL)
# Heroku automatically exports: DATABASE_URL=postgres://user:pass@host:port/db
# It handles SSL by default, no configuration needed

# Heroku Redis (via REDIS_URL)
# Heroku automatically exports: REDIS_URL=redis://default:pass@host:port

# For explicit configuration if needed:
DB_HOST=${DATABASE_URL_HOST}
DB_PORT=${DATABASE_URL_PORT}
DB_USER=${DATABASE_URL_USER}
DB_PASSWORD=${DATABASE_URL_PASSWORD}
DB_NAME=${DATABASE_URL_DBNAME}
DB_SSL_MODE=require

REDIS_HOST=${REDIS_URL_HOST}
REDIS_PORT=${REDIS_URL_PORT}
REDIS_PASSWORD=${REDIS_URL_PASSWORD}

# Security
JWT_SECRET=$(openssl rand -hex 32)
BASE_URL=https://yourapp.herokuapp.com
CORS_ORIGINS=https://app.yourdomain.com

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_app_id
GITHUB_CLIENT_SECRET=your_github_app_secret
```

### Setup Steps for Heroku

1. **Provision Databases**
   ```bash
   heroku addons:create heroku-postgresql:standard-0
   heroku addons:create heroku-redis:premium-0
   ```

2. **Configure App**
   ```bash
   heroku config:set JWT_SECRET=$(openssl rand -hex 32)
   heroku config:set GITHUB_CLIENT_ID=xxx
   heroku config:set BASE_URL=https://yourapp.herokuapp.com
   ```

3. **Deploy**
   ```bash
   git push heroku main
   heroku logs --tail
   ```

## Self-Hosted (Kubernetes)

```bash
# .env (Kubernetes deployment)
APP_ENV=production
PORT=8080
LOG_LEVEL=info

# PostgreSQL Service (internal Kubernetes DNS)
DB_HOST=postgres-service.default.svc.cluster.local
DB_PORT=5432
DB_USER=pushpaka
DB_PASSWORD=your_strong_password  # Use Kubernetes Secret
DB_NAME=pushpaka_prod
DB_SSL_MODE=disable  # Internal network, optional
DB_SSL_CERT_FILE=
DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=10

# Redis Service (internal Kubernetes DNS)
REDIS_HOST=redis-service.default.svc.cluster.local
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password  # Use Kubernetes Secret
REDIS_DB=0
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=5

# Security
JWT_SECRET=your_secret_from_kubernetes_secret
BASE_URL=https://pushpaka.yourdomain.com
CORS_ORIGINS=https://app.yourdomain.com

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_app_id
GITHUB_CLIENT_SECRET=your_github_app_secret
```

### Setup Steps for Kubernetes

1. **Create Secrets**
   ```bash
   kubectl create secret generic pushpaka-db \
     --from-literal=password=your_strong_password
   kubectl create secret generic pushpaka-redis \
     --from-literal=password=your_redis_password
   ```

2. **Deploy with Helm or Kustomize**
   ```bash
   helm install pushpaka ./helm -f values-production.yaml
   ```

## Health Check URLs

Test your configuration:

```bash
# PostgreSQL
psql 'postgres://user:password@host:port/db?sslmode=require'

# Redis
redis-cli -h host -p port -a password ping

# Application Health
curl -i http://localhost:8080/health
```

## Common Issues & Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| "SSL: CERTIFICATE_VERIFY_FAILED" | Wrong cert path | Verify `DB_SSL_CERT_FILE` absolute path |
| "connection refused" | Wrong host/port | Test with `psql` or `redis-cli` |
| "authentication failed" | Wrong password | Double-check credentials in secrets manager |
| "pool timeout" | Too many connections | Increase `DB_MAX_OPEN_CONNS` |
| "connection pool exhausted" | Low pool size | Increase `REDIS_POOL_SIZE` |
