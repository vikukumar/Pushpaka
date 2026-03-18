# Pushpaka CI/CD - Quick Reference

## 🎯 At a Glance

| What | Where | Trigger |
|------|-------|---------|
| **Automatic Release** | GitHub Actions | Push to `main` |
| **All Binaries** | GitHub Releases | Auto on push |
| **Docker Image** | GHCR | Auto on push |
| **Manual Release** | Workflow Dispatch | Manual button |
| **Tests Only (No Release)** | CI Workflow | PR/Push any branch |

---

## 🔄 Build Counter Logic

```
Version File: 1.0.0

Build 1 → v1.0.0.0 (first release)
Build 2 → v1.0.0.1 (counter +1)
Build 3 → v1.0.0.2 (counter +1)

Change VERSION to 1.0.1:
Build 4 → v1.0.1.0 (counter resets)
Build 5 → v1.0.1.1 (counter +1)
```

---

## 📂 Key Files

```
Pushpaka/
├── VERSION                          # Version file (e.g., "1.0.0")
├── Dockerfile                       # Updated with build args
├── .github/
│   ├── workflows/
│   │   ├── release.yml             # ⭐ Main workflow (auto-release)
│   │   ├── ci.yml                  # Tests on PR/push
│   │   └── manual-release.yml      # Manual trigger
│   ├── scripts/
│   │   ├── manage-workflows.sh     # Linux/macOS helper
│   │   └── manage-workflows.ps1    # Windows PowerShell helper
│   ├── WORKFLOWS.md                # Detailed documentation
│   └── SETUP.md                    # Setup guide (you are here)
```

---

## ⚡ Quick Commands

### Show Version Info
```bash
# Linux/macOS
./.github/scripts/manage-workflows.sh version

# Windows PowerShell
.\.github\scripts\manage-workflows.ps1 version
```

### Bump Version
```bash
# Linux/macOS
./.github/scripts/manage-workflows.sh bump 1.0.1

# Windows PowerShell
.\.github\scripts\manage-workflows.ps1 bump 1.0.1
```

### Validate Workflows
```bash
# Linux/macOS
./.github/scripts/manage-workflows.sh validate

# Windows PowerShell
.\.github\scripts\manage-workflows.ps1 validate
```

### Test Local Build
```bash
# Linux/macOS
./.github/scripts/manage-workflows.sh test backend
./.github/scripts/manage-workflows.sh test frontend
./.github/scripts/manage-workflows.sh test docker

# Windows PowerShell
.\.github\scripts\manage-workflows.ps1 test backend
.\.github\scripts\manage-workflows.ps1 test frontend
.\.github\scripts\manage-workflows.ps1 test docker
```

---

## 🚀 Release Flows

### Automatic (Recommended)
```
1. Make changes
2. git push origin main
3. ✅ Automatic: tests → build → release → docker → github release
```

### Manual (One Component)
```
1. GitHub → Actions → "Manual Release Component"
2. Select: backend-only / worker-only / frontend-only / docker-only / all
3. ✅ Release that component(s)
```

### Version Bump
```
1. echo "1.0.1" > VERSION
2. git add VERSION && git commit -m "chore: bump to 1.0.1"
3. git push origin main
4. ✅ Next build: v1.0.1.0 (counter resets)
```

---

## 📦 Release Artifacts

```
Binaries (per platform × 2 archs):
├── pushpaka-backend-linux-amd64
├── pushpaka-backend-linux-arm64
├── pushpaka-backend-darwin-amd64
├── pushpaka-backend-darwin-arm64
├── pushpaka-backend-windows-amd64.exe
├── pushpaka-worker-linux-amd64
├── pushpaka-worker-linux-arm64
├── pushpaka-worker-darwin-amd64
├── pushpaka-worker-darwin-arm64
├── pushpaka-worker-windows-amd64.exe
└── pushpaka-frontend-1.0.0.5.tar.gz

Combined Packages (Backend + Worker + Frontend):
├── pushpaka-linux-1.0.0.5.tar.gz
├── pushpaka-darwin-1.0.0.5.tar.gz
└── pushpaka-windows-1.0.0.5.zip

Docker Images:
├── ghcr.io/org/pushpaka:v1.0.0.5
└── ghcr.io/org/pushpaka:latest

Release:
├── GitHub Release with all above
└── SHA256SUMS checksum file
```

---

## 🐳 Docker Quick Start

```bash
# Pull latest
docker pull ghcr.io/org/pushpaka:latest

# Run all components
docker run -e PUSHPAKA_COMPONENT=all \
  ghcr.io/org/pushpaka:latest

# Run API only
docker run -e PUSHPAKA_COMPONENT=api \
  ghcr.io/org/pushpaka:latest

# Run with ports and database
docker run -d \
  -p 8080:8080 \
  -e DB_URL=postgres://user:pass@localhost/pushpaka \
  ghcr.io/org/pushpaka:v1.0.0.5
```

---

## ✅ Checklist: First Time Setup

- [ ] VERSION file exists (cat VERSION → "1.0.0")
- [ ] .github/workflows/ directory has 3 YAML files
- [ ] GitHub Actions enabled in repository settings
- [ ] Make a test commit and push to main
- [ ] Check Actions tab → "Build & Release" running
- [ ] Wait for build to complete (~10-15 minutes)
- [ ] Check Releases tab → first release is there
- [ ] Check Container Registry → image pushed to GHCR

---

## 🆘 Common Issues

| Problem | Solution |
|---------|----------|
| "Docker push failed" | Enable GHCR in repo settings |
| "Build counter not incrementing" | Check VERSION file exists at root |
| "Frontend build timeout" | Clear Actions cache, re-run |
| "Go build failed" | Check Go version in workflow matches your system |
| "Workflow not triggering" | Ensure push is to main, files have changes |

---

## 📖 where to find what

| Question | Answer |
|----------|--------|
| How do I bump the version? | Edit VERSION file, git push |
| How do I release only backend? | Manual workflow dispatch → backend-only |
| Where are my binaries? | GitHub Releases tab → download |
| How do I pull the Docker image? | `docker pull ghcr.io/org/pushpaka:latest` |
| What's my current version? | `cat VERSION` |
| Why is build hanging? | Check Actions tab for logs |
| How do I test builds locally? | `.github/scripts/manage-workflows.sh test all` |

---

## 🔗 Links

- **Status:** GitHub → Actions → "Build & Release"
- **Releases:** GitHub → Releases → latest
- **Docker Image:** GHCR → your-org → pushpaka
- **Version:** File → `VERSION` at repo root
- **Documentation:** `.github/WORKFLOWS.md`
- **Setup Guide:** `.github/SETUP.md`

---

**Quick Help:**
```bash
.github/scripts/manage-workflows.sh help
# or on Windows:
.\.github\scripts\manage-workflows.ps1 help
```

---

**Version:** 1.0.0 | **Status:** ✅ Production Ready
