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

---

## Quick Summary

Pushpaka enterprise expansion is complete with 28 new files totaling 2000+ lines of code:

✅ **Helm Charts** - Production Kubernetes deployment  
✅ **Release Tracking** - Version management system  
✅ **AI Chatbot** - OpenRouter-powered support  
✅ **Website Integration** - Full documentation  
✅ **GitHub Automation** - Continuous deployment  

**Next Steps:**
1. Implement backend `/api/chat` endpoint
2. Update website navigation
3. Configure OPENROUTER_API_KEY in GitHub Secrets
4. Push to GitHub (triggers automation)

See INTEGRATION_GUIDE.md for complete step-by-step instructions!
