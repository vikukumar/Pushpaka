# Pushpaka Helm Chart README

## Overview

This is the official Helm chart for deploying Pushpaka on Kubernetes. It provides a complete, production-ready setup with:

- ✅ **API Server** - Backend with horizontal scaling
- ✅ **Dashboard** - Frontend with Next.js
- ✅ **Workers** - Build and deployment workers
- ✅ **PostgreSQL** - Production database
- ✅ **Redis** - Caching and job queue
- ✅ **Traefik** - Reverse proxy with auto SSL
- ✅ **Cert-Manager** - Let's Encrypt integration
- ✅ **Monitoring** - Prometheus & Grafana (optional)

## Quick Start

```bash
# Add repository
helm repo add pushpaka https://vikukumar.github.io/Pushpaka/helm
helm repo update

# Install
helm install pushpaka pushpaka/pushpaka -n pushpaka --create-namespace
```

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+  
- kubectl configured
- 2+ CPU, 4GB RAM minimum

## Features

### High Availability
- Multi-replica deployment for all components
- Pod Disruption Budgets
- Anti-affinity rules for pod distribution
- Health checks and readiness probes

### Auto Scaling
- Horizontal Pod Autoscaler for API, Dashboard, and Workers
- CPU and memory-based scaling
- Configurable min/max replicas

### Security
- RBAC enabled
- Pod security contexts
- Network policies
- Private secret management

### Monitoring
- Prometheus metrics export
- Grafana dashboards (optional)
- Health endpoints
- Liveness and readiness probes

### Data Persistence
- PostgreSQL with persistent volumes
- Redis replication
- Configurable storage classes
- Backup-ready setup

## Configuration

All configuration is done through `values.yaml`. Key settings:

```yaml
global:
  domain: "pushpaka.example.com"        # Your domain
  tls:
    enabled: true                        # Enable HTTPS

api:
  replicaCount: 2                       # API replicas
  autoscaling:
    enabled: true                       # Enable HPA
    minReplicas: 2
    maxReplicas: 5

postgresql:
  auth:
    password: "changeme"               # Change database password!

redis:
  auth:
    password: "changeme"               # Change redis password!
```

See [values.yaml](./values.yaml) for complete configuration options.

## Installation Methods

### Method 1: Default Installation
```bash
helm install pushpaka pushpaka/pushpaka -n pushpaka --create-namespace
```

### Method 2: Custom Values
```bash
helm install pushpaka pushpaka/pushpaka \
  -n pushpaka \
  --create-namespace \
  -f my-values.yaml
```

### Method 3: Individual Values
```bash
helm install pushpaka pushpaka/pushpaka \
  -n pushpaka \
  --create-namespace \
  --set global.domain=my.pushpaka.com \
  --set postgresql.auth.password=mysecretpass \
  --set redis.auth.password=myredispass
```

## Upgrade

```bash
# Update repo
helm repo update

# Upgrade release
helm upgrade pushpaka pushpaka/pushpaka -n pushpaka -f values.yaml

# Rollback if needed
helm rollback pushpaka -n pushpaka
```

## Uninstall

```bash
helm uninstall pushpaka -n pushpaka
kubectl delete namespace pushpaka
```

## Customization

### Use External Database
```yaml
postgresql:
  enabled: false

api:
  env:
    DATABASE_URL: "postgresql://user:pass@external.db:5432/pushpaka"
```

### Use External Redis
```yaml
redis:
  enabled: false

api:
  env:
    REDIS_URL: "redis://external.redis:6379"
```

### Custom Storage Class
```yaml
postgresql:
  primary:
    persistence:
      storageClassName: "fast-ssd"

redis:
  master:
    persistence:
      storageClassName: "fast-ssd"
```

### Disable Monitoring
```yaml
monitoring:
  enabled: false
```

## Troubleshooting

### Pods not starting?
```bash
kubectl describe pod POD_NAME -n pushpaka
kubectl logs POD_NAME -n pushpaka
```

### Check ingress
```bash
kubectl get ingress -n pushpaka
kubectl describe ingress pushpaka-api -n pushpaka
```

### Database connection issues?
```bash
# Verify PostgreSQL is running
kubectl get pods -n pushpaka | grep postgres

# Check connection
kubectl run -it --rm debug --image=busybox --restart=Never -n pushpaka -- \
  /bin/sh -c "wget -O- http://pushpaka-api:8080/api/v1/health"
```

## License

Same as Pushpaka project (MIT License)

## Support

- 📖 [Pushpaka Documentation](https://vikukumar.github.io/Pushpaka/)
- 🐛 [GitHub Issues](https://github.com/vikukumar/pushpaka/issues)
- 💬 [Discussions](https://github.com/vikukumar/pushpaka/discussions)
