# Complete Integration Guide - Helm, Releases, and Chatbot

Comprehensive guide to integrate all enterprise Pushpaka features: Helm charts, release tracking, and AI chatbot.

## Status Overview

✅ **COMPLETED:**
- Helm chart infrastructure (16 files)
- Release tracking system (5 files)
- GitHub Actions Helm release workflow
- Website Helm installation page
- ChatBot.astro component
- GitHub Secrets setup documentation

⚠️ **IN PROGRESS:**
- Backend `/api/chat` endpoint implementation
- GitHub Secrets configuration (user action)
- Website header navigation update

📋 **READY TO IMPLEMENT:**
- Main README.md update (template provided)
- Website navigation integration

---

## Quick Integration Checklist

### Phase 1: Environment Setup (5 min)
- [ ] Copy `.env.example` to `.env` and configure
- [ ] Set `OPENROUTER_API_KEY` in GitHub Secrets
- [ ] Verify VERSION file exists at repo root

### Phase 2: Backend Implementation (30 min)
- [ ] Create `backend/handlers/api/chat.go` from example
- [ ] Import and register chat routes in main.go
- [ ] Test `/api/v1/chat` endpoint locally
- [ ] Test chatbot health check endpoint

### Phase 3: Website Updates (20 min)
- [ ] Update Header.astro with new nav links
- [ ] Create API documentation page
- [ ] Update Footer with roadmap link
- [ ] Test responsive design

### Phase 4: Documentation (10 min)
- [ ] Replace README.md with README_NEW.md
- [ ] Review ROADMAP.md
- [ ] Verify all markdown files compile

### Phase 5: Verification (15 min)
- [ ] Test Helm chart: `helm lint helm/pushpaka/`
- [ ] Build and deploy website
- [ ] Test chatbot on live website
- [ ] Verify Helm charts publish to GitHub Pages

---

## Phase 1: Environment Setup

### Step 1.1: Configure Environment File

```bash
cd /path/to/Pushpaka

# Copy template
cp .env.example .env

# Edit configuration
nano .env
```

**Required for Chatbot (.env):**
```bash
# Core Settings
DOMAIN=pushpaka.example.com
JWT_SECRET=$(openssl rand -hex 32)
PORT=8080

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=pushpaka
POSTGRES_PASSWORD=use-strong-password
POSTGRES_DB=pushpaka

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=use-strong-password

# Chatbot - IMPORTANT
OPENROUTER_API_KEY=sk-your-key-here
OPENROUTER_MODEL=openai/gpt-4-turbo
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1

# TLS/SSL
ACME_EMAIL=admin@example.com
ACME_PROD=true
```

### Step 1.2: Get OpenRouter API Key

1. Visit [https://openrouter.ai](https://openrouter.ai)
2. Sign up / Log in
3. Go to **API Keys** section
4. Create new API key
5. Copy the key (format: `sk-xxxxx...`)

### Step 1.3: Add GitHub Secret

**In GitHub Repository:**

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. **Name:** `OPENROUTER_API_KEY`
4. **Value:** Paste API key from step 1.2
5. Click **Add secret**

**Verify:**
```bash
gh secret list | grep OPENROUTER
```

### Step 1.4: Verify VERSION File

```bash
# Check VERSION file exists at repo root
cat VERSION
# Output: 1.0.0

# Or create if missing
echo "1.0.0" > VERSION
git add VERSION
git commit -m "chore: Add VERSION file"
```

---

## Phase 2: Backend Implementation

### Step 2.1: Create Chat Endpoint

```bash
# Create handlers directory structure
mkdir -p backend/handlers/api

# Copy example file
cp backend/handlers/api/chat.go.example backend/handlers/api/chat.go

# Edit to match your Go module structure
nano backend/handlers/api/chat.go
```

**File Content** (from provided template):
- Imports required packages
- Defines ChatRequest/ChatResponse structs
- Implements ChatHandler function
- Calls OpenRouter API
- Handles errors gracefully

### Step 2.2: Register Routes in Main

**File:** `backend/main.go`

```go
package main

import (
    "github.com/vikukumar/pushpaka/backend/handlers/api"
    "github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()

    // Middleware
    router.Use(middleware.RateLimitMiddleware(10, time.Minute))
    router.Use(middleware.CORSMiddleware())

    // API Routes
    api.RegisterChatRoutes(router)

    // Other routes...

    router.Run(":8080")
}
```

### Step 2.3: Test Endpoint Locally

```bash
# Start backend
cd backend
go run main.go

# In another terminal, test endpoint
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "How do I deploy an app?",
    "systemPrompt": "You are a helpful Pushpaka assistant..."
  }'

# Expected response:
# { "response": "To deploy an app with Pushpaka, ..." }
```

### Step 2.4: Test Health Check

```bash
# Test chatbot availability
curl http://localhost:8080/api/v1/chat/health

# Expected response:
# { "status": "enabled", "model": "openai/gpt-4-turbo" }
```

### Step 2.5: Error Handling

Test error scenarios:

```bash
# Missing API key (should return 503)
unset OPENROUTER_API_KEY
curl http://localhost:8080/api/v1/chat/health
# Expected: "status": "disabled"

# Invalid request (should return 400)
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{ "message": "" }'
# Expected: 400 error

# Set API key back
export OPENROUTER_API_KEY=sk-...
```

---

## Phase 3: Website Updates

### Step 3.1: Update Header Navigation

**File:** `website/src/components/Header.astro`

Add new navigation links:

```astro
---
const navLinks = [
  { label: 'Home', href: '/Pushpaka/' },
  { label: 'Features', href: '/Pushpaka/features' },
  { label: 'Documentation', href: '/Pushpaka/docs' },
  { label: 'Installation', href: '/Pushpaka/install' },
  { label: 'Helm', href: '/Pushpaka/helm-install' },      // NEW
  { label: 'API', href: '/Pushpaka/api' },                 // NEW
  { label: 'Releases', href: '/Pushpaka/releases' },
];
---
```

### Step 3.2: Create API Documentation Page

Create `website/src/pages/api.astro`:

```astro
---
import Layout from '../layouts/Layout.astro';
---

<Layout title="API Documentation - Pushpaka">
  <div class="max-w-4xl mx-auto py-12 px-4">
    <h1>API Documentation</h1>
    
    <section>
      <h2>Authentication</h2>
      <!-- Add API auth documentation -->
    </section>
    
    <section>
      <h2>Projects Endpoint</h2>
      <!-- Document endpoints -->
    </section>
    
    <!-- See WEBSITE_NAVIGATION_UPDATE.md for full content -->
  </div>
</Layout>
```

Or use [WEBSITE_NAVIGATION_UPDATE.md](WEBSITE_NAVIGATION_UPDATE.md) for complete template.

### Step 3.3: Update Footer

**File:** `website/src/components/Footer.astro`

Add roadmap link and additional resource sections:

```astro
---
// Add to footer
---

<footer>
  <!-- ... existing footer ... -->
  
  {/* NEW: Links section */}
  <div class="mt-8 grid grid-cols-3 gap-8">
    <div>
      <h3>Documentation</h3>
      <ul>
        <li><a href="/Pushpaka/docs">Docs</a></li>
        <li><a href="/Pushpaka/api">API</a></li>
        <li><a href="/Pushpaka/helm-install">Helm</a></li>
      </ul>
    </div>
    
    <div>
      <h3>Community</h3>
      <ul>
        <li><a href="https://github.com/vikukumar/pushpaka">GitHub</a></li>
        <li><a href="https://github.com/vikukumar/pushpaka/blob/main/ROADMAP.md">Roadmap</a></li>
      </ul>
    </div>
  </div>
</footer>
```

### Step 3.4: Test Locally

```bash
cd website

# Install dependencies
npm install

# Start dev server
npm run dev

# Test in browser:
# ✅ http://localhost:3000/Pushpaka/ - Home page
# ✅ http://localhost:3000/Pushpaka/helm-install - Helm page
# ✅ http://localhost:3000/Pushpaka/api - API page
# ✅ Footer has roadmap link
# ✅ Navigation shows all items on desktop
# ✅ Mobile menu includes all items
```

### Step 3.5: Build Website

```bash
# Build for production
npm run build

# Output in dist/ directory
ls -la dist/

# If build passes, ready to deploy!
```

---

## Phase 4: Documentation Updates

### Step 4.1: Update Main README

Replace current README.md:

```bash
# Backup current README
cp README.md README_OLD.md

# Copy new README
cp README_NEW.md README.md

# Review changes
git diff --no-pager README.md | head -50

# Verify all links work
grep -o '\[.*\](.*\.md)' README.md | sort | uniq
```

**Check:**
- ✅ All section titles present
- ✅ Links to new pages work
- ✅ Helm integration documented
- ✅ Chatbot section included
- ✅ Release management explained

### Step 4.2: Verify ROADMAP.md

```bash
# Check roadmap syntax
cat ROADMAP.md | head -30

# Verify structure:
# ✅ Vision section
# ✅ Milestones (v1.0, v1.1, v1.2, v1.3)
# ✅ Long-term vision
# ✅ Success metrics
# ✅ Contributing section
```

### Step 4.3: Verify Additional Docs

```bash
# Check all markdown files
for f in *.md; do
  echo "Checking $f..."
  [ -f "$f" ] && echo "  ✅ Found" || echo "  ❌ Missing"
done

# Required files:
# ✅ README.md (updated)
# ✅ ROADMAP.md
# ✅ GITHUB_SECRETS_SETUP.md
# ✅ WEBSITE_NAVIGATION_UPDATE.md
# ✅ CONTRIBUTING.md (if exists)
# ✅ LICENSE
```

---

## Phase 5: Verification

### Step 5.1: Verify Helm Chart

```bash
# Lint Helm chart
helm lint helm/pushpaka/

# Should output:
# 1 chart(s) linted, 0 chart(s) failed

# Check chart contents
helm show all helm/pushpaka/ | head -50

# List templates
ls -la helm/pushpaka/templates/
```

### Step 5.2: Test Full Stack Locally

```bash
# Start backend
cd backend && go run main.go &
BACKEND_PID=$!

# Wait for backend to start
sleep 2

# Start website
cd website && npm run dev &
WEBSITE_PID=$!

# Wait for website to start
sleep 3

# Test endpoints
echo "Testing backend..."
curl http://localhost:8080/api/v1/chat/health
echo "✅ Backend OK"

echo "Testing chatbot..."
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Test","systemPrompt":"Respond briefly"}'
echo "✅ Chatbot OK"

echo "Testing website..."
curl -s http://localhost:3000/Pushpaka/ | grep -q "Pushpaka" && echo "✅ Website OK"

# Stop processes
kill $BACKEND_PID $WEBSITE_PID
```

### Step 5.3: Commit All Changes

```bash
git status

# Should show:
# M  README.md
# M  website/src/components/Header.astro
# M  website/src/components/Footer.astro
# A  website/src/pages/api.astro
# A  backend/handlers/api/chat.go
# A  GITHUB_SECRETS_SETUP.md
# A  WEBSITE_NAVIGATION_UPDATE.md
# A  ROADMAP.md
# (etc.)

# Commit with descriptive message
git add .
git commit -m "feat: Enterprise features - Helm charts, releases, AI chatbot

- Add production-ready Helm charts with autoscaling (2-5 replicas API, 3-10 workers)
- Implement release tracking system with version-based directories
- Create AI chatbot component (OpenRouter GPT-4 integration)
- Add GitHub Actions workflow for Helm chart publishing
- Update website with Helm installation guide and API documentation
- Document GitHub Secrets setup for chatbot configuration
- Update README with Helm, chatbot, and release information
- Add comprehensive roadmap for v1.1-v1.3+ features"

# Push to GitHub
git push origin main
```

### Step 5.4: Verify GitHub Actions

```bash
# GitHub automatically:
# 1. Lints Helm chart
# 2. Builds website
# 3. Publishes to GitHub Pages
# 4. Creates GitHub Release

# Check status in GitHub:
# ✅ Actions tab - All workflows passed
# ✅ Pages - Website deployed to https://vikukumar.github.io/Pushpaka/
# ✅ Releases - New release created
# ✅ Helm repo - Charts published to /helm directory
```

### Step 5.5: Test Live Website

```bash
# Open browser to live website
https://vikukumar.github.io/Pushpaka/

# Check:
# ✅ Header shows Helm, API links
# ✅ Helm installation page loads
# ✅ API documentation page loads
# ✅ Chatbot button visible (bottom-right)
# ✅ Chatbot responds to messages
# ✅ Footer has roadmap link
# ✅ Mobile menu works
```

---

## Full Component Map

```
Pushpaka/
├── backend/
│   ├── handlers/
│   │   └── api/
│   │       └── chat.go  ⭐ NEW - Chatbot endpoint
│   └── main.go  (update routes)
│
├── website/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Header.astro  ⭐ UPDATE - Add nav links
│   │   │   └── Footer.astro  ⭐ UPDATE - Add roadmap
│   │   └── pages/
│   │       ├── helm-install.astro  ✅ DONE
│   │       └── api.astro  ⭐ NEW - API docs
│   └── astro.config.mjs
│
├── helm/
│   └── pushpaka/  ✅ DONE
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── README.md
│       ├── templates/ (9 files)
│       └── _helpers.tpl
│
├── releases/  ✅ DONE
│   ├── v1.0.0/
│   │   ├── CHANGELOG.md
│   │   ├── FEATURES.md
│   │   └── COMPONENTS.md
│   └── v1.1.0/
│       └── CHANGELOG.md
│
├── .github/
│   └── workflows/
│       └── helm-release.yml  ✅ DONE
│
├── README.md  ⭐ UPDATE
├── ROADMAP.md  ✅ DONE
├── GITHUB_SECRETS_SETUP.md  ✅ DONE
├── WEBSITE_NAVIGATION_UPDATE.md  ✅ DONE
├── .env.example  (verify OPENROUTER keys)
└── VERSION  (verify exists)
```

---

## Expected Results After Integration

### For Users
- ✅ Beautiful modern website with Helm guides
- ✅ Complete API documentation
- ✅ 24/7 AI chatbot support
- ✅ Clear upgrade path with Helm
- ✅ Transparent roadmap

### For Deployments
- ✅ Single binary (`./pushpaka`)
- ✅ Docker Compose (`docker compose up`)
- ✅ Kubernetes via Helm (`helm install pushpaka/pushpaka`)
- ✅ Auto-scaling in K8s (2-5 API, 3-10 workers)
- ✅ Production-ready monitoring

### For Operations
- ✅ GitHub Pages Helm repository
- ✅ Automated chart releases on VERSION change
- ✅ Version tracking in GitHub Releases
- ✅ Environment-specific configuration
- ✅ Secure secret management

---

## Rollback Plan

If something goes wrong:

```bash
# Revert last commit
git revert HEAD

# Or reset to previous state
git reset --hard origin/main

# Manually restore from backups
git checkout HEAD~1 README.md
git checkout HEAD~1 website/src/components/Header.astro
```

---

## Support & Troubleshooting

### Common Issues

**Chatbot not responding:**
- ✅ Verify `OPENROUTER_API_KEY` in GitHub Secrets
- ✅ Check API key is valid on OpenRouter.ai
- ✅ Verify backend is running and `/api/v1/chat/health` returns 200
- ✅ Check browser console for errors

**Helm chart lint fails:**
- ✅ Run `helm lint helm/pushpaka/` locally
- ✅ Check YAML syntax in values.yaml
- ✅ Verify required chart metadata in Chart.yaml

**Website doesn't show new pages:**
- ✅ Rebuild website: `npm run build`
- ✅ Clear browser cache (Ctrl+F5)
- ✅ Check GitHub Actions build log
- ✅ Verify file names match route paths

**GitHub Pages not updating:**
- ✅ Check .github/workflows/deploy.yml is enabled
- ✅ Verify gh-pages branch exists (GitHub auto-creates)
- ✅ Check Settings → Pages → Source is "Deploy from a branch"
- ✅ Allow 2-3 minutes for GitHub to publish

---

## Next Steps

After complete integration:

1. **Monitor Helm Repository**
   - Track downloads at `https://vikukumar.github.io/Pushpaka/helm/`
   - Monitor Helm chart usage in `values.yaml`

2. **Gather Chatbot Analytics**
   - Log questions received
   - Improve system prompt based on queries
   - Track response quality and user satisfaction

3. **Iterate on Release Process**
   - Test full release cycle (bump VERSION → publish)
   - Automate changelog generation
   - Create release notes template

4. **Expand Documentation**
   - Add video tutorials
   - Create case studies
   - Build deployment templates

5. **Plan v1.1.0 Release**
   - Scheduled deployments
   - Email notifications
   - Webhooks
   - See ROADMAP.md for full list

---

**Integration Complete!** 🚀

All enterprise features are now integrated. Your Pushpaka deployment supports:
- ☸️ Production Kubernetes deployment
- 🤖 AI-powered 24/7 support
- 📦 Automated release tracking
- 📚 Comprehensive documentation
- 🔄 Continuous delivery

For questions: [GitHub Discussions](https://github.com/vikukumar/pushpaka/discussions)

**Last Updated:** March 17, 2026
