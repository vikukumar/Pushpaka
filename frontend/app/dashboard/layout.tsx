import { AuthProvider } from '@/components/providers/AuthProvider'

// Force all dashboard pages to be server-rendered on every request.
// Removed automatically by scripts/patch-layout.js when building the static
// export for Go binary embedding (STATIC_EXPORT=1), which doesn't support SSR.
export const dynamic = 'force-dynamic'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return <AuthProvider>{children}</AuthProvider>
}
