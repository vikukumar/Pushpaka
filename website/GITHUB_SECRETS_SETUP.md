# GitHub Secrets Setup Guide

This guide explains how to configure GitHub Secrets for Pushpaka's AI chatbot and deployment features.

## Overview

GitHub Secrets are encrypted environment variables stored securely in your repository. They're used to pass sensitive information to your GitHub Actions workflows and deployed applications.

## Required Secrets

### 1. OpenRouter API Key (Chatbot)

The chatbot feature uses OpenRouter to provide AI-powered support 24/7.

**Steps:**

1. **Get API Key from OpenRouter**
   - Visit [https://openrouter.ai](https://openrouter.ai)
   - Sign up or log in
   - Go to **API Keys** → **Create Key**
   - Copy the API key (format: `sk-xxx...`)

2. **Add to GitHub Secrets**
   - Go to your GitHub repository
   - Click **Settings** → **Secrets and variables** → **Actions**
   - Click **New repository secret**
   - Name: `OPENROUTER_API_KEY`
   - Value: Paste the API key from OpenRouter
   - Click **Add secret**

3. **Verify**
   - The secret appears in the list (value is masked)
   - GitHub Actions workflows can now access it via `${{ secrets.OPENROUTER_API_KEY }}`
   - Website automatically enables chatbot feature

**Environment Variables (for reference):**
```bash
# Set in your deployment OR GitHub Secrets
OPENROUTER_API_KEY=sk-your-api-key-here
OPENROUTER_MODEL=openai/gpt-4-turbo
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
```

---

### 2. Docker Registry (Optional)

For private Docker registries or automated Docker Hub pushes.

**DockerHub:**

```bash
# Username
DOCKER_USERNAME=youruser

# Access Token (not password!)
DOCKER_PASSWORD=dckr_pat_xxxxx
```

**GitHub Container Registry (GHCR):**

```bash
# Already available via: ${{ secrets.GITHUB_TOKEN }}
# No configuration needed!
```

---

### 3. Deployment Secrets (Optional)

For automated deploys to cloud providers.

**DigitalOcean / AWS / Azure:**

```bash
# Store cloud provider credentials
CLOUD_PROVIDER=digitalocean
CLOUD_API_TOKEN=dop_v1_xxxxx
```

---

## How to Add a Secret

### Through GitHub UI

1. **Navigate to Secrets**
   ```
   GitHub Repository → Settings → Secrets and variables → Actions
   ```

2. **Create New Secret**
   - Click **New repository secret**
   - **Name:** `OPENROUTER_API_KEY` (all caps, underscores for spaces)
   - **Value:** Paste your secret value
   - Click **Add secret**

3. **Verify**
   - Secret appears in list but value is hidden
   - Status shows as "Available" (green checkmark)

### Through GitHub CLI

```bash
# Login to GitHub CLI
gh auth login

# Add secret
gh secret set OPENROUTER_API_KEY --body "sk-your-key"

# List secrets
gh secret list

# Delete secret
gh secret delete OPENROUTER_API_KEY
```

---

## Using Secrets in GitHub Actions

### In Workflows

```yaml
name: Deploy with Chatbot

on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Access secret in environment
      - name: Set up environment
        env:
          OPENROUTER_KEY: ${{ secrets.OPENROUTER_API_KEY }}
        run: |
          echo "Chatbot enabled"
          echo "API Key: [MASKED]"

      # Or pass as GitHub Actions input
      - name: Build and test
        run: |
          npm run build
          npm test
        env:
          OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
```

### In Helm Deployment

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pushpaka-secrets
  namespace: pushpaka
type: Opaque
stringData:
  OPENROUTER_API_KEY: "{{ secrets.OPENROUTER_API_KEY }}"
  DATABASE_URL: "postgresql://user:pass@postgresql:5432/pushpaka"
  REDIS_URL: "redis://:password@redis:6379"
```

---

## Best Practices

### Security

✅ **DO:**
- ✅ Use strong, unique API keys (40+ characters)
- ✅ Rotate keys periodically (quarterly or monthly)
- ✅ Use separate keys for dev/prod environments
- ✅ Store keys in GitHub Secrets, never in code
- ✅ Use environment-specific secret names (e.g., `PROD_API_KEY`)

❌ **DON'T:**
- ❌ Commit secrets to version control
- ❌ Log or print secret values
- ❌ Use test/demo keys in production
- ❌ Share secrets via email or chat
- ❌ Use overly permissive API key scopes

### Organization

```bash
# Naming Convention (env_provider_resource)
PROD_OPENROUTER_API_KEY
DEV_OPENROUTER_API_KEY
PROD_DOCKER_PASSWORD
PROD_DATABASE_URL

# Grouping by environment
# Secrets tab shows:
# - PROD_* secrets (production)
# - DEV_* secrets (development)
# - SHARED_* secrets (all environments)
```

---

## Troubleshooting

### Secret Not Accessible in Workflow

**Problem:** Workflow can't access secret

**Solution:**
1. Verify secret name matches exactly (case-sensitive)
2. Check syntax: `${{ secrets.SECRET_NAME }}`
3. Verify workflow has read permissions
4. Re-authenticate with `gh auth login`

```yaml
# ✅ Correct
env:
  MY_SECRET: ${{ secrets.OPENROUTER_API_KEY }}

# ❌ Wrong
env:
  MY_SECRET: ${{ secrets.openrouter_api_key }}  # Case mismatch!
```

### Chatbot Not Working

**Problem:** Chatbot returns errors

**Steps:**
1. Verify `OPENROUTER_API_KEY` is set in secrets
2. Check API key is valid (not expired/revoked)
3. Verify quota on OpenRouter account
4. Check browser console for errors
5. Enable debug logging

```bash
# Verify in deployment
echo "Key Status: ${OPENROUTER_API_KEY:+FOUND}${OPENROUTER_API_KEY:-NOT_FOUND}"
```

### Secret in Logs

**Problem:** Secret accidentally printed in logs

**GitHub automatically masks:**
- Values matching registered secrets
- Known secret patterns
- API key formats

**But be extra careful:**
```yaml
# ❌ AVOID - Uses secrets in script
run: echo "Secret is ${{ secrets.OPENROUTER_API_KEY }}"

# ✅ BETTER - Pass only via environment
env:
  KEY: ${{ secrets.OPENROUTER_API_KEY }}
run: echo "Secret loaded"  # Don't print it!
```

---

## Advanced: Branch-Specific Secrets

Some providers allow branch-specific secret overrides:

```bash
# Environment secrets (GitHub Enterprise)
PROD_OPENROUTER_API_KEY  (only for main branch)
DEV_OPENROUTER_API_KEY   (for develop branch)
```

Configure in:
```
Settings → Environments → [environment name] → Secrets
```

---

## Integration Points

### Deployed Website (Astro)

The website automatically uses secrets:

```javascript
// In /api/chat endpoint
const apiKey = process.env.OPENROUTER_API_KEY;
const model = process.env.OPENROUTER_MODEL || "openai/gpt-4-turbo";

if (!apiKey) {
  return { error: "Chatbot not configured" };
}

// Call OpenRouter
const response = await fetch("https://openrouter.ai/api/v1/chat/completions", {
  method: "POST",
  headers: {
    "Authorization": `Bearer ${apiKey}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    model: model,
    messages: [
      { role: "system", content: systemPrompt },
      { role: "user", content: message },
    ],
  }),
});
```

### Helm Deployment

Create `helm-values-secrets.yaml`:

```yaml
env:
  OPENROUTER_API_KEY:
    secretRef:
      name: pushpaka-secrets
      key: OPENROUTER_API_KEY
  DATABASE_URL:
    secretRef:
      name: pushpaka-secrets
      key: DATABASE_URL
```

Deploy with:
```bash
helm install pushpaka pushpaka/pushpaka \
  -f helm-values-secrets.yaml
```

---

## Verification Checklist

After setup, verify:

- [ ] Secret is listed in `Settings → Secrets and variables → Actions`
- [ ] Secret value is masked (shows as `●●●●●●●●`)
- [ ] Workflow can access secret without errors
- [ ] Chatbot responds to messages on website
- [ ] API logs show successful OpenRouter calls
- [ ] No secret values printed in GitHub Actions logs

---

## Support

**If you encounter issues:**

1. Check [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
2. Verify API key is valid on provider's website
3. Review GitHub Actions logs for error messages
4. Contact support: vikukumar@example.com

---

**Last Updated:** March 17, 2026  
**Version:** v1.0.0
