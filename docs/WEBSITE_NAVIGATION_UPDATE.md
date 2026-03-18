# Website Navigation Update Guide

This guide explains how to update the website header to include links to new pages (Helm installation, releases, documentation).

## Current Structure

The website is built with **Astro 5.1 + React 19 + Tailwind CSS**.

### Current Header Component

Location: `website/src/components/Header.astro`

**Current navigation items:**
- Home
- Features  
- Documentation
- Installation
- Releases

### New Pages to Add

1. **Helm Installation** - `website/src/pages/helm-install.astro`
   - Detailed Kubernetes/Helm setup guide
   - Prerequisites, quick start, configuration, troubleshooting
   - Should appear after "Installation" in nav

2. **API Documentation** - `website/src/pages/api.astro` (create)
   - API endpoints, authentication, examples
   - Should appear after "Documentation" in nav

3. **Roadmap** - Link to `ROADMAP.md`
   - Product roadmap and planned features
   - Should appear in footer

## Update Steps

### 1. Update Header Navigation

**File:** `website/src/components/Header.astro`

**Add to navigation array:**

```astro
---
// website/src/components/Header.astro

const navLinks = [
  { label: 'Home', href: '/Pushpaka/' },
  { label: 'Features', href: '/Pushpaka/features' },
  { label: 'Documentation', href: '/Pushpaka/docs' },
  { label: 'Installation', href: '/Pushpaka/install' },
  
  // NEW: Helm deployment guide
  { label: 'Helm', href: '/Pushpaka/helm-install' },
  
  // NEW: API documentation
  { label: 'API', href: '/Pushpaka/api' },
  
  { label: 'Releases', href: '/Pushpaka/releases' },
];
---

<header class="py-4 px-6 bg-gray-900/50 border-b border-gray-800">
  <nav class="flex gap-8">
    {navLinks.map(link => (
      <a href={link.href} class="hover:text-brand-400 transition">
        {link.label}
      </a>
    ))}
  </nav>
</header>
```

### 2. Update Mobile Navigation

If using mobile hamburger menu:

```astro
---
// In the mobile menu section of Header.astro
const mobileNavLinks = [
  { label: 'Home', href: '/Pushpaka/' },
  { label: 'Features', href: '/Pushpaka/features' },
  { label: 'Documentation', href: '/Pushpaka/docs' },
  { label: 'Installation', href: '/Pushpaka/install' },
  { label: 'Helm', href: '/Pushpaka/helm-install' },
  { label: 'API', href: '/Pushpaka/api' },
  { label: 'Releases', href: '/Pushpaka/releases' },
];
---

<!-- Mobile menu (show on small screens) -->
<div class="md:hidden">
  <button id="mobile-menu-btn" class="text-white">☰</button>
  <div id="mobile-menu" class="hidden flex flex-col gap-4 p-4">
    {mobileNavLinks.map(link => (
      <a href={link.href} class="block hover:text-brand-400">
        {link.label}
      </a>
    ))}
  </div>
</div>
```

### 3. Update Footer with Roadmap Link

**File:** `website/src/components/Footer.astro`

```astro
---
// website/src/components/Footer.astro
---

<footer class="py-8 px-6 bg-gray-900 border-t border-gray-800">
  <div class="max-w-6xl mx-auto">
    <!-- ... existing footer content ... -->
    
    <!-- NEW: Additional links section -->
    <div class="mt-8 grid grid-cols-1 md:grid-cols-3 gap-8">
      <div>
        <h3 class="text-brand-400 font-semibold mb-4">Documentation</h3>
        <ul class="space-y-2">
          <li><a href="/Pushpaka/docs" class="hover:text-brand-400">Docs</a></li>
          <li><a href="/Pushpaka/api" class="hover:text-brand-400">API</a></li>
          <li><a href="/Pushpaka/helm-install" class="hover:text-brand-400">Helm</a></li>
        </ul>
      </div>
      
      <div>
        <h3 class="text-brand-400 font-semibold mb-4">Community</h3>
        <ul class="space-y-2">
          <li><a href="https://github.com/vikukumar/Pushpaka" class="hover:text-brand-400">GitHub</a></li>
          <li><a href="https://github.com/vikukumar/Pushpaka/releases" class="hover:text-brand-400">Releases</a></li>
          <li><a href="${process.env.PUBLIC_GITHUB_REPO}/blob/main/ROADMAP.md" class="hover:text-brand-400">Roadmap</a></li>
        </ul>
      </div>
      
      <div>
        <h3 class="text-brand-400 font-semibold mb-4">Support</h3>
        <ul class="space-y-2">
          <li><a href="https://github.com/vikukumar/Pushpaka/issues" class="hover:text-brand-400">Issues</a></li>
          <li><a href="https://github.com/vikukumar/Pushpaka/discussions" class="hover:text-brand-400">Discussions</a></li>
          <li><span class="text-gray-400">💬 Chatbot available on site</span></li>
        </ul>
      </div>
    </div>
  </div>
</footer>
```

### 4. Create API Documentation Page

**File:** `website/src/pages/api.astro`

```astro
---
import Layout from '../layouts/Layout.astro';
import { Code } from 'astro:components';
---

<Layout title="API Documentation - Pushpaka">
  <div class="min-h-screen py-12 px-4">
    <div class="max-w-4xl mx-auto">
      
      <h1 class="text-4xl md:text-5xl font-bold mb-6 bg-gradient-to-r from-brand-400 to-cyan-400 bg-clip-text text-transparent">
        API Documentation
      </h1>

      <p class="text-xl text-gray-300 mb-12">
        Complete reference for Pushpaka REST API endpoints. All requests require authentication.
      </p>

      <!-- Authentication section -->
      <section class="mb-16">
        <h2 class="text-2xl font-bold mb-6 text-white">Authentication</h2>
        
        <p class="text-gray-300 mb-4">
          Two authentication methods are supported:
        </p>

        <div class="bg-gray-900 rounded-lg p-6 mb-6">
          <h3 class="text-lg font-semibold mb-3 text-brand-400">1. JWT Token</h3>
          
          <p class="text-gray-300 mb-4">Get token via login endpoint:</p>
          
          <pre class="bg-gray-800 p-4 rounded overflow-x-auto">
<code>POST /api/v1/auth/login
Content-Type: application/json

{'{'}
  "email": "admin@example.com",
  "password": "password"
{'}'}

Response:
{'{'}
  "token": "eyJhbGc...",
  "expires_in": 3600
{'}'}
</code>
          </pre>

          <p class="text-gray-300 mt-4">
            Use token in subsequent requests:
          </p>

          <pre class="bg-gray-800 p-4 rounded overflow-x-auto">
<code>GET /api/v1/projects
Authorization: Bearer eyJhbGc...</code>
          </pre>
        </div>

        <div class="bg-gray-900 rounded-lg p-6">
          <h3 class="text-lg font-semibold mb-3 text-brand-400">2. API Key</h3>
          
          <p class="text-gray-300 mb-4">Generate API key in dashboard settings:</p>

          <pre class="bg-gray-800 p-4 rounded overflow-x-auto">
<code>GET /api/v1/projects
Authorization: ApiKey sk-xxxxxxxxxxxxxx</code>
          </pre>
        </div>
      </section>

      <!-- Projects endpoints -->
      <section class="mb-16">
        <h2 class="text-2xl font-bold mb-6 text-white">Projects</h2>
        
        <div class="space-y-6">
          {/* List projects */}
          <div class="bg-gray-900 rounded-lg p-6">
            <h3 class="font-mono text-cyan-400 mb-2">GET /api/v1/projects</h3>
            <p class="text-gray-300 mb-4">List all projects for the authenticated user.</p>
            
            <div class="bg-gray-800 p-3 rounded text-sm">
              <p class="text-gray-400 mb-2">Response:</p>
              <pre>
<code>{`{
  "projects": [
    {
      "id": "proj_123",
      "name": "My App",
      "repository": "https://github.com/user/app",
      "status": "active",
      "created_at": "2024-03-17T00:00:00Z"
    }
  ],
  "count": 1
}`}</code>
              </pre>
            </div>
          </div>

          {/* Create project */}
          <div class="bg-gray-900 rounded-lg p-6">
            <h3 class="font-mono text-cyan-400 mb-2">POST /api/v1/projects</h3>
            <p class="text-gray-300 mb-4">Create a new project.</p>
            
            <div class="bg-gray-800 p-3 rounded text-sm">
              <p class="text-gray-400 mb-2">Request:</p>
              <pre>
<code>{`{
  "name": "My New App",
  "repository": "https://github.com/user/new-app",
  "git_token": "ghp_xxxxxxxx" (optional, for private repos)
}`}</code>
              </pre>
            </div>
          </div>
        </div>
      </section>

      {/* Deployments endpoints */}
      <section class="mb-16">
        <h2 class="text-2xl font-bold mb-6 text-white">Deployments</h2>
        
        <div class="space-y-6">
          <div class="bg-gray-900 rounded-lg p-6">
            <h3 class="font-mono text-cyan-400 mb-2">POST /api/v1/projects/{'{id}'}/deployments</h3>
            <p class="text-gray-300 mb-4">Trigger a new deployment.</p>
            
            <div class="bg-gray-800 p-3 rounded text-sm">
              <p class="text-gray-400 mb-2">Request:</p>
              <pre>
<code>{`{
  "branch": "main",
  "commit": "abc123def456" (optional)
}`}</code>
              </pre>
            </div>
          </div>

          <div class="bg-gray-900 rounded-lg p-6">
            <h3 class="font-mono text-cyan-400 mb-2">GET /api/v1/projects/{'{id}'}/deployments/{'{depId}'}/logs/stream</h3>
            <p class="text-gray-300 mb-4">Stream real-time deployment logs via WebSocket.</p>
          </div>
        </div>
      </section>

      {/* Rate limiting */}
      <section class="mb-16">
        <h2 class="text-2xl font-bold mb-6 text-white">Rate Limiting</h2>
        
        <div class="bg-gray-900 rounded-lg p-6">
          <p class="text-gray-300 mb-4">Rate limits are included in response headers:</p>
          
          <pre class="bg-gray-800 p-4 rounded">
<code>X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1647532800</code>
          </pre>

          <p class="text-gray-300 mt-4">
            Default limits:
          </p>
          <ul class="list-disc list-inside text-gray-300 mt-2">
            <li>1000 requests per hour (authenticated users)</li>
            <li>100 requests per hour (public endpoints)</li>
            <li>10 concurrent builds per project</li>
          </ul>
        </div>
      </section>

      {/* Errors */}
      <section>
        <h2 class="text-2xl font-bold mb-6 text-white">Error Handling</h2>
        
        <div class="bg-gray-900 rounded-lg p-6">
          <p class="text-gray-300 mb-4">Errors include a status code and message:</p>
          
          <pre class="bg-gray-800 p-4 rounded">
<code>{`{
  "error": "Unauthorized",
  "code": "401",
  "message": "Invalid authentication token"
}`}</code>
          </pre>

          <p class="text-gray-300 mt-4">Common status codes:</p>
          <ul class="list-disc list-inside text-gray-300 mt-2">
            <li>200: Success</li>
            <li>400: Bad Request</li>
            <li>401: Unauthorized</li>
            <li>404: Not Found</li>
            <li>429: Rate Limited</li>
            <li>500: Internal Server Error</li>
          </ul>
        </div>
      </section>

    </div>
  </div>
</Layout>
```

### 5. Update Main Navigation Configuration

**File:** `website/src/layouts/Layout.astro`

If using a centralized navigation configuration:

```astro
---
// website/src/data/navigation.ts

export const navigation = [
  {
    label: 'Docs',
    href: '/Pushpaka/docs',
    icon: '📖',
  },
  {
    label: 'Installation',
    href: '/Pushpaka/install',
    icon: '⚙️',
  },
  {
    label: 'Helm',
    href: '/Pushpaka/helm-install',
    icon: '☸️',
    new: true, // Badge to show new page
  },
  {
    label: 'API',
    href: '/Pushpaka/api',
    icon: '🔌',
  },
  {
    label: 'Releases',
    href: '/Pushpaka/releases',
    icon: '📦',
  },
];

export const footerLinks = [
  {
    section: 'Documentation',
    links: [
      { label: 'Guides', href: '/Pushpaka/docs' },
      { label: 'API Reference', href: '/Pushpaka/api' },
      { label: 'Helm Charts', href: '/Pushpaka/helm-install' },
    ],
  },
  {
    section: 'Resources',
    links: [
      { label: 'Roadmap', href: 'https://github.com/vikukumar/Pushpaka/blob/main/ROADMAP.md' },
      { label: 'Releases', href: 'https://github.com/vikukumar/Pushpaka/releases' },
      { label: 'GitHub', href: 'https://github.com/vikukumar/Pushpaka' },
    ],
  },
];
---
```

## Testing

After updates, verify:

```bash
# Navigate to website directory
cd website/src

# Build website
npm run build

# Start dev server
npm run dev

# Check:
# ✅ Header shows all new links (Helm, API)
# ✅ Links navigate correctly
# ✅ Mobile menu includes new pages
# ✅ Footer shows roadmap link
# ✅ New pages load without errors
# ✅ Active link highlighting works
```

## Integration Checklist

- [ ] Update Header.astro with new nav links
- [ ] Add mobile menu items
- [ ] Update Footer.astro with additional links
- [ ] Create api.astro page with documentation
- [ ] Test all navigation links locally
- [ ] Verify responsive design on mobile
- [ ] Test URL paths match Astro configuration
- [ ] Build passes without errors
- [ ] Push to GitHub (triggers GitHub Pages deploy)
- [ ] Verify live website shows all new pages

## Deployment

After making changes:

```bash
# Commit changes
git add website/src/components/Header.astro
git add website/src/components/Footer.astro
git add website/src/pages/api.astro
git commit -m "feat: Add Helm, API, and Roadmap nav links"

# Push to GitHub
git push origin main

# GitHub Actions automatically:
# 1. Builds website
# 2. Deploys to GitHub Pages
# 3. Updates https://vikukumar.github.io/Pushpaka/
```

---

**Pages Now Available:**
- Home: `/Pushpaka/`
- Features: `/Pushpaka/features`
- Documentation: `/Pushpaka/docs`
- Installation: `/Pushpaka/install`
- **[NEW] Helm Installation:** `/Pushpaka/helm-install`
- **[NEW] API Documentation:** `/Pushpaka/api`
- Releases: `/Pushpaka/releases`

**Last Updated:** March 17, 2026
