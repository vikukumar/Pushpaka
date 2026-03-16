/** @type {import('next').NextConfig} */

// Static export is only used when building for the Go binary embedding.
// Set STATIC_EXPORT=1 to produce a fully static output that can be copied into
// backend/ui/dist and embedded in the Go binary.  During `next dev` and plain
// `pnpm build` the app runs as a standard Next.js server.
const isStaticExport = process.env.STATIC_EXPORT === '1'

const nextConfig = {
  // Only apply static-export settings when explicitly requested.
  ...(isStaticExport ? { output: 'export', trailingSlash: true } : {}),
  images: {
    // unoptimized required for output: export; harmless in dev.
    unoptimized: true,
    remotePatterns: [
      { protocol: 'https', hostname: 'avatars.githubusercontent.com' },
    ],
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    NEXT_PUBLIC_WS_URL: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080',
  },
}

module.exports = nextConfig
