# Release v1.0.0 - Initial Production Release

**Release Date:** March 17, 2026  
**Status:** Latest Stable

## Overview

First production-ready release of Pushpaka platform with complete deployment pipeline, monitoring, and management features.

## Changes

### ✨ New Features

- **One-Click Git Deployments** - Simple deployment workflow for public and private repositories
- **Automatic Dockerization** - Framework detection for 15+ frameworks (Next.js, React, Go, Python, etc.)
- **Real-Time Logs** - WebSocket-based live log streaming with filtering capabilities
- **Secure Private Repos** - Encrypted PAT storage for private repository access
- **Rollback Support** - Instant rollback to any previous deployment
- **Multi-Project** - Support for unlimited projects per user
- **Multi-User** - Team collaboration with role-based access control
- **Traefik v3 Routing** - Advanced reverse proxy with automatic SSL/TLS
- **Let's Encrypt Integration** - Free automatic SSL certificates with auto-renewal
- **Dashboard** - Modern Next.js dashboard with dark/light theme support
- **REST API** - Complete API for programmatic access
- **Prometheus Metrics** - Export metrics for monitoring and alerting
- **Health Checks** - Comprehensive health and readiness endpoints

### 🚀 Performance

- Sub-second deployment triggers
- Optimized container builds
- Parallel job processing
- In-process queue for dev mode
- ~100/100 Lighthouse score

### 🔒 Security

- JWT v5 authentication
- bcrypt password hashing (cost 10)
- Secure CORS headers
- Rate limiting on all endpoints
- Git token redaction from logs
- Secure environment variable storage

### 🛠️ Infrastructure

- Single-binary deployment in dev mode
- dockercompose production setup
- Kubernetes-ready with Helm charts
- PostgreSQL 17 support and SQLite fallback
- Redis 8 for job queue and caching
- Comprehensive logging

### 📦 Components

- Backend: Go 1.25 with Gin framework
- Frontend: Next.js 16.1.6 with React 19.2.4
- Worker: Go 1.25 with goroutines
- Dashboard: Next.js with Tailwind CSS
- Build System: Multi-platform (Linux, macOS, Windows)

## Downloads

### Binaries
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)
- Docker image available at: `ghcr.io/vikukumar/pushpaka:v1.0.0`

### Container
```bash
docker pull ghcr.io/vikukumar/pushpaka:v1.0.0
```

## Helm Installation

```bash
helm repo add pushpaka https://vikukumar.github.io/Pushpaka/helm
helm repo update
helm install pushpaka pushpaka/pushpaka -n pushpaka --create-namespace -f values.yaml
```

## Breaking Changes

None - this is the initial release.

## Known Issues

- Kubernetes integration testing ongoing
- Helm chart documentation in progress

## Migration Guide

N/A - Initial release, no migration needed.

## Thanks

Thanks to the Traefik, PostgreSQL, Go, and Kubernetes communities for the amazing tools!

## Support

- 🐛 Report bugs: https://github.com/vikukumar/Pushpaka/issues
- 💬 Discussions: https://github.com/vikukumar/Pushpaka/discussions
- 📚 Docs: https://vikukumar.github.io/Pushpaka/docs
