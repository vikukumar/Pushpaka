# GitHub Actions CI/CD Pipeline Configuration

This document describes the GitHub Actions workflows for building, testing, and releasing Pushpaka.

## Workflows

### 1. **Release** (`release.yml`)
Automatically triggered on pushes to `main` with changes in key directories. Creates releases for all platforms.


**Triggers:**
- Push to `main` with changes in: `backend/`, `worker/`, `frontend/`, `cmd/`, `go.work`, `Dockerfile`, `.github/workflows/`
- Manual dispatch via `workflow_dispatch` with optional `force_version` input

**Output:**
- Multi-platform binaries (Linux, macOS, Windows)
- Frontend build artifacts
- Combined platform packages
- Docker image pushed to GHCR
- GitHub Release with all artifacts
- Checksums (SHA256SUMS)

**Version Management:**
- Reads `VERSION` file for base version (e.g., `1.0.0`)
- Auto-increments build counter (e.g., `1.0.0.0` → `1.0.0.1`)
- Resets counter when version changes (e.g., `1.0.0` → `1.0.1` starts at `1.0.1.0`)

### 2. **CI** (`ci.yml`)
Runs on every push and PR to `main` and `develop`. Tests code quality without releasing.

**Checks:**
- ✅ Go fmt, vet, and build checks (backend, worker)
- ✅ TypeScript type checking and ESLint
- ✅ Next.js build validation
- ✅ Dockerfile linting with Hadolint
- ✅ Security scanning with Trivy

**Failures block merges** to main/develop.

### 3. **Manual Release** (`manual-release.yml`)
Released on-demand via `workflow_dispatch` for selective component releases.

**Options:**
- `all` - Release everything (default)
- `backend-only` - Backend binaries only
- `worker-only` - Worker binaries only
- `frontend-only` - Frontend artifacts only
- `docker-only` - Docker image only

**Inputs:**
- `component` - Which component(s) to release
- `version` - Override version (optional, auto-increments if empty)

---

## Version Management

### File: `VERSION`
Located at repository root. Contains the base semantic version (e.g., `1.0.0`).

```bash
# To bump version
echo "1.0.1" > ./VERSION
git add VERSION
git commit -m "chore: bump version to 1.0.1"
git push
```

### Version Format
- **Full Version:** `v1.0.0.5` (v + base version + build counter)
- **Docker Tags:**
  - `latest` - Points to latest release
  - `v1.0.0.5` - Specific version tag
  - GHCR: `ghcr.io/your-username/pushpaka:v1.0.0.5`

### Workflow
1. First release of `1.0.0`: `v1.0.0.0`
2. Second build (same version): `v1.0.0.1`
3. Third build (same version): `v1.0.0.2`
4. Change version to `1.0.1`: `v1.0.1.0` (counter resets)

---

## Build Artifacts

### Backend
- `pushpaka-backend-linux-amd64`
- `pushpaka-backend-linux-arm64`
- `pushpaka-backend-darwin-amd64` (macOS)
- `pushpaka-backend-darwin-arm64` (macOS ARM)
- `pushpaka-backend-windows-amd64.exe`

### Worker
- `pushpaka-worker-linux-amd64`
- `pushpaka-worker-linux-arm64`
- `pushpaka-worker-darwin-amd64`
- `pushpaka-worker-darwin-arm64`
- `pushpaka-worker-windows-amd64.exe`

### Frontend
- `pushpaka-frontend-1.0.0.5.tar.gz` (Next.js build + public)

### Combined Packages
- `pushpaka-linux-1.0.0.5.tar.gz` (Backend + Worker + Frontend)
- `pushpaka-darwin-1.0.0.5.tar.gz` (Backend + Worker + Frontend)
- `pushpaka-windows-1.0.0.5.zip` (Backend + Worker + Frontend)

### Docker
- `ghcr.io/your-username/pushpaka:v1.0.0.5`
- `ghcr.io/your-username/pushpaka:latest`

---

## Configuration

### Setup for First Time

1. **Enable GitHub Container Registry:**
   - Go to: Settings → Packages and registries → Container registry
   - Ensure "Inherit access from repository" is checked

2. **Update VERSION file:**
   ```bash
   echo "1.0.0" > ./VERSION
   git add VERSION
   git commit -m "Initial version"
   ```

3. **Trigger first release:**
   - Option A: Push to `main` with changes to trigger auto-release
   - Option B: Go to Actions → "Build & Release" → "Run workflow" → Dispatch

### Environment Variables

Set in GitHub Actions Secrets (if needed):
- `GHCR_TOKEN` - Container registry token (auto-provided by `GITHUB_TOKEN`)
- `VERSION` - Can override (optional, auto-managed)

---

## Performance Optimizations

### Go Build Cache
- Uses `actions/setup-go@v4` with `cache: true`
- Automatic caching of Go module downloads
- Shared across jobs

### Node Build Cache
- Uses `actions/setup-node@v4` with `cache: pnpm`
- pnpm lock file enables deterministic builds
- Automatic dependency caching

### Docker BuildKit
- Enabled via `docker/build-push-action@v5`
- GitHub Actions cache backend for layer caching
- `cache-from` and `cache-to` for build performance

### Parallelization
- Backend builds: 6 parallel jobs (3 OS × 2 arch)
- Worker builds: 6 parallel jobs
- All run concurrently where possible
- Typical build time: 8-12 minutes

---

## GitHub Release Format

### Release Page Example

```
Release v1.0.0.5

📦 Release v1.0.0.5

📋 Artifacts

Binaries:
- pushpaka-backend (Linux, macOS, Windows - amd64 & arm64)
- pushpaka-worker (Linux, macOS, Windows - amd64 & arm64)

Frontend:
- pushpaka-frontend-1.0.0.5.tar.gz (Next.js build artifacts)

Combined Packages:
- pushpaka-linux-1.0.0.5.tar.gz (Backend + Worker + Frontend)
- pushpaka-darwin-1.0.0.5.tar.gz (Backend + Worker + Frontend)
- pushpaka-windows-1.0.0.5.zip (Backend + Worker + Frontend)

🐳 Docker Image

docker pull ghcr.io/your-username/pushpaka:v1.0.0.5
docker pull ghcr.io/your-username/pushpaka:latest

🔐 Checksums
See SHA256SUMS file for verification.
```

---

## Workflow Triggers

| Workflow | Trigger | Auto-Release |
|----------|---------|--------------|
| **release.yml** | Push to `main` (path-based) | ✅ Yes (tag + release) |
| **ci.yml** | PR/push to `main`/`develop` | ❌ No (tests only) |
| **manual-release.yml** | Manual dispatch | ✅ Selective |

---

## Error Handling

### Build Failure
- Workflow fails immediately on error
- PR blocks until fixed
- Release is not created
- Check logs in Actions tab for details

### Version Conflict
- If `VERSION` file changed manually and conflicts exist
- Use `workflow_dispatch` with `force_version` to override
- Manual release always succeeds (no tag conflicts)

---

## Usage Examples

### Example 1: Release Everything Automatically
```bash
# Push changes to main
git add .
git commit -m "feat: add new feature"
git push origin main

# Workflow automatically:
# 1. Runs CI tests
# 2. Builds all binaries
# 3. Creates Docker image
# 4. Publishes release
# 5. Tags with v1.0.0.X
```

### Example 2: Manual Backend-Only Release
```bash
# Via GitHub UI:
# Actions → Manual Release → Run Workflow
# Component: backend-only
# Version: (leave empty for auto-increment)

# OR via GitHub CLI:
gh workflow run manual-release.yml \
  -f component=backend-only
```

### Example 3: Bump Minor Version
```bash
# Update VERSION file
echo "1.0.1" > ./VERSION
git add VERSION
git commit -m "chore: bump to 1.0.1"
git push

# Next build will be v1.0.1.0
```

### Example 4: Pull Docker Image
```bash
docker pull ghcr.io/your-username/pushpaka:latest
docker pull ghcr.io/your-username/pushpaka:v1.0.0.5

# Run with config
docker run -e PUSHPAKA_COMPONENT=all \
  ghcr.io/your-username/pushpaka:latest
```

---

## Troubleshooting

### Docker Image Not Pushing
- ✅ Check GitHub Container Registry permissions
- ✅ Verify `GITHUB_TOKEN` has package write permissions
- ✅ Check repo is public or token has private access

### Build Cache Not Working
- Clear cache: Actions → Manage → Delete all caches
- Re-run workflow to rebuild cache

### Version Counter Not Incrementing
- Verify `VERSION` file is at repo root
- Check git tags: `git tag -l`
- Use `workflow_dispatch` with `force_version` to fix

### Artifact Download Issues
- Artifacts available for 5 days (configurable)
- Check Actions tab for artifact links
- Use `gh release download` to get release artifacts

---

## Maintenance

### Update Go Version
Edit workflows: change `go-version: '1.25'` to desired version

### Update Node Version
Edit workflows: change `node-version: '20'` to desired version

### Change Image Registry
Find and replace: `ghcr.io/your-username/pushpaka`

### Update Build Cache Duration
In workflows: `retention-days: 5` (adjust as needed)
