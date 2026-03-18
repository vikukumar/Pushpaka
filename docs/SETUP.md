# Pushpaka CI/CD Pipeline Setup Guide

Complete guide for GitHub Actions CI/CD with versioning, multi-platform builds, and Docker releases.

## 📋 Prerequisites

- GitHub repository with Actions enabled
- Docker configured in GitHub (or GHCR access)
- Go 1.25+ (local testing)
- pnpm (local testing)
- Git

## 🚀 Quick Start (2 minutes)

### 1. Verify VERSION file exists
```bash
cat VERSION  # Should output: 1.0.0
```

### 2. Push to main
```bash
git add .
git commit -m "feat: initial setup"
git push origin main
```

### 3. Watch GitHub Actions
Go to: **Your Repo** → **Actions** → **Build & Release** → Watch the build!

---

## 📦 What Gets Released

### When you push to `main`:
✅ **Backend binaries** (3 platforms × 2 architectures = 6 binaries)
✅ **Worker binaries** (6 binaries)
✅ **Frontend** (Next.js build artifacts)
✅ **Combined packages** (Backend + Worker + Frontend per platform)
✅ **Docker image** (pushed to GHCR)
✅ **GitHub Release** (with all artifacts)

### Build counter auto-increments:
- First: `v1.0.0.0`
- Second: `v1.0.0.1`
- Third: `v1.0.0.2`
- After version bump to `1.0.1`: `v1.0.1.0` (resets)

---

## 🔧 Configuration

### Update VERSION
To create a new version:

```bash
# Edit VERSION file
echo "1.0.1" > VERSION

# Commit and push
git add VERSION
git commit -m "chore: bump version to 1.0.1"
git push origin main

# Next build: v1.0.1.0 (counter resets)
```

**Linux/macOS:**
```bash
./github/scripts/manage-workflows.sh bump 1.0.1
```

**Windows (PowerShell):**
```powershell
.\\.github\scripts\manage-workflows.ps1 bump 1.0.1
```

### Update Go Version
Edit `.github/workflows/release.yml`:
```yaml
- uses: actions/setup-go@v4
  with:
    go-version: '1.26'  # Change this
```

### Update Node Version
Edit `.github/workflows/release.yml`:
```yaml
- uses: actions/setup-node@v4
  with:
    node-version: '22'  # Change this
```

---

## 📊 Workflows Overview

| Workflow | Trigger | Output | Status |
|----------|---------|--------|--------|
| `release.yml` | Push to `main` (auto) | Full release | ✅ Complete |
| `ci.yml` | PR/Push | Tests only | ✅ Complete |
| `manual-release.yml` | Manual dispatch | Selective | ✅ Complete |

### Detailed Workflow Files

**Location:** `.github/workflows/`
- `release.yml` - 50KB, 580 lines
- `ci.yml` - 15KB, 150 lines
- `manual-release.yml` - 25KB, 250 lines

**Workflow Scripts:**

**Location:** `.github/scripts/`
- `manage-workflows.sh` - Linux/macOS helper
- `manage-workflows.ps1` - Windows PowerShell helper

---

## 🏗️ Build Process

### Step 1: Version Calculation
- Reads `VERSION` file
- Checks latest git tag
- Auto-increments counter or resets

### Step 2: Parallel Builds (all at once)
```
Backend (6 binaries)    Worker (6 binaries)    Frontend (1 artifact)
├─ linux-amd64          ├─ linux-amd64         └─ .next build
├─ linux-arm64          ├─ linux-arm64
├─ darwin-amd64         ├─ darwin-amd64
├─ darwin-arm64         ├─ darwin-arm64
├─ windows-amd64        └─ windows-amd64
└─ (windows-arm64 skipped - not supported)

All run concurrently → ~8-12 minutes total
```

### Step 3: Docker Build
- Builds single image with all components
- Pushes to GHCR (github container registry)
- Tags: `latest` and `v1.0.0.5`

### Step 4: Release Creation
- Creates GitHub Release
- Uploads all binaries
- Generates SHA256SUMS
- Creates combined platform packages

---

## 🐳 Docker Image

### Pull Latest
```bash
docker pull ghcr.io/your-username/pushpaka:latest
docker pull ghcr.io/your-username/pushpaka:v1.0.0.5
```

### Run Configuration
Set environment variables to control which components start:

```bash
# Run API + Workers (default)
docker run -e PUSHPAKA_COMPONENT=all \
  ghcr.io/your-username/pushpaka:latest

# Run API only
docker run -e PUSHPAKA_COMPONENT=api \
  ghcr.io/your-username/pushpaka:latest

# Run Workers only
docker run -e PUSHPAKA_COMPONENT=worker \
  ghcr.io/your-username/pushpaka:latest
```

### With Ports & Volumes
```bash
docker run -d \
  --name pushpaka \
  -p 8080:8080 \
  -v /deploy:/deploy/pushpaka \
  -e PUSHPAKA_COMPONENT=all \
  -e DB_URL="postgres://user:pass@localhost/pushpaka" \
  ghcr.io/your-username/pushpaka:latest
```

---

## 📥 Using Released Artifacts

### Option 1: GitHub Release Page
- Go to **Releases** → Latest
- Download binaries or combined packages
- Extract and run

### Option 2: Direct Binary
```bash
# Linux
./pushpaka-backend-linux-amd64

# macOS
./pushpaka-backend-darwin-amd64

# Windows
.\pushpaka-backend-windows-amd64.exe
```

### Option 3: Package Download
```bash
# Linux
tar -xzf pushpaka-linux-1.0.0.5.tar.gz
cd pushpaka-linux-1.0.0.5
./bin/pushpaka-backend-linux-amd64

# macOS
tar -xzf pushpaka-darwin-1.0.0.5.tar.gz
cd pushpaka-darwin-1.0.0.5
./bin/pushpaka-backend-darwin-amd64

# Windows
Expand-Archive pushpaka-windows-1.0.0.5.zip
cd pushpaka-windows-1.0.0.5
.\bin\pushpaka-backend-windows-amd64.exe
```

### Option 4: Docker Compose
```yaml
services:
  pushpaka:
    image: ghcr.io/your-username/pushpaka:v1.0.0.5
    environment:
      PUSHPAKA_COMPONENT: all
      DB_URL: postgres://user:pass@postgres:5432/pushpaka
    ports:
      - "8080:8080"
    volumes:
      - /deploy:/deploy/pushpaka
      - ./config:/etc/pushpaka
```

---

## 🔍 Verify Artifacts

### Check Checksums
```bash
# Download SHA256SUMS from release
sha256sum -c SHA256SUMS

# Output should show "OK" for each file
pushpaka-backend-linux-amd64: OK
pushpaka-worker-linux-amd64: OK
...
```

### Check Binary Version
```bash
./pushpaka-backend-linux-amd64 --version
# Output: pushpaka v1.0.0.5

./pushpaka-worker-linux-amd64 --version
# Output: pushpaka-worker v1.0.0.5
```

---

## 🎯 Common Workflows

### Workflow 1: Regular Release (Automatic)
```bash
# 1. Make changes
git add .
git commit -m "feat: add cool feature"

# 2. Push to main
git push origin main

# 3. GitHub Actions automatically:
#    - Runs CI tests
#    - Builds all binaries
#    - Creates Docker image
#    - Publishes release
#    ✅ Done! Check Releases tab
```

### Workflow 2: Bump Major/Minor Version
```bash
# 1. Update VERSION file
echo "1.1.0" > VERSION

# 2. Commit
git add VERSION
git commit -m "chore: bump version to 1.1.0"

# 3. Push
git push origin main

# Next release will be: v1.1.0.0
```

### Workflow 3: Release Only Backend & Docker
```bash
# Via GitHub UI:
# 1. Go to Actions
# 2. Select "Manual Release Component"
# 3. Fill in:
#    - Component: backend-only
#    - Version: (leave empty)
# 4. Click "Run workflow"

# ✅ Only backend binaries + docker image released
```

### Workflow 4: Hot Fix Release
```bash
# 1. Create fix on main
git add .
git commit -m "fix: critical bug"
git push origin main

# 2. If you need to release immediately (don't wait for build):
# Go to Actions → Manual Release
# Component: all
# Version: (leave empty)
# Run workflow

# ✅ Happens immediately
```

---

## 🚨 Troubleshooting

### "Build Failed: Docker push failed"
**Solution:**
- Go to Repo Settings → Packages and registries
- Ensure "Inherit access from repository" is enabled
- Check `GITHUB_TOKEN` has package write permissions

### "Version counter not incrementing"
**Solution:**
```bash
# Verify VERSION file exists and is readable
cat VERSION

# Check git tags
git tag -l

# If confused, force a new version
echo "1.0.1" > VERSION
git add VERSION && git commit -m "chore: reset version"
git push
```

### "Frontend build timeout"
**Solution:**
- Check `pnpm-lock.yaml` is committed
- Clear cache: Actions → Manage → Delete all cache
- Re-run workflow

### "Windows binary not building"
**Reason:** Windows ARM64 is intentionally excluded
- Windows only supports amd64
- Check workflow output for actual errors

---

## ⚙️ Advanced Configuration

### Custom Build Hooks (Add to release.yml)
```yaml
- name: Custom Build Step
  run: |
    echo "Custom build step here"
    ./scripts/pre-build.sh
```

### Change Docker Registry
In all workflows, find:
```yaml
REGISTRY: ghcr.io
```

Change to:
```yaml
REGISTRY: docker.io  # or your registry
```

### Cache Retention
In workflows, adjust:
```yaml
retention-days: 5  # Change to desired number
```

### Add More Platforms
In `release.yml`, add to strategy matrix:
```yaml
goarch: [amd64, arm64, 386]  # Add 386 for 32-bit
```

---

## 📋 Checklist

Before your first release:

- [ ] `VERSION` file exists at repo root
- [ ] `.github/workflows/release.yml` is present
- [ ] `.github/workflows/ci.yml` is present
- [ ] GitHub Actions are enabled in repo
- [ ] Container registry access is configured
- [ ] All tests pass (`ci.yml` should succeed)
- [ ] You can see workflow files in Actions tab

---

## 📞 Support

For issues:
1. Check workflow logs: Actions → Your Workflow → Click failed step
2. Review `.github/WORKFLOWS.md` for detailed documentation
3. Run local tests: `.github/scripts/manage-workflows.sh test all`
4. Check git tags: `git tag -l`

---

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Docker Official Images](https://hub.docker.com/_/golang)
- [Semantic Versioning](https://semver.org/)

---

**Last Updated:** 2024
**Version:** 1.0.0
**Status:** ✅ Production Ready
