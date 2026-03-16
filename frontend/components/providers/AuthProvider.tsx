'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/auth'
import { Sidebar } from '@/components/layout/Sidebar'

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, _hasHydrated } = useAuthStore()
  const router = useRouter()

  useEffect(() => {
    // Only redirect after the auth store has finished rehydrating from
    // localStorage.  Without this guard, the effect fires on the first
    // render (isAuthenticated = false, default) before Zustand persist has
    // had a chance to restore the stored session, causing a spurious logout.
    if (_hasHydrated && !isAuthenticated) {
      router.replace('/login')
    }
  }, [_hasHydrated, isAuthenticated, router])

  // Show nothing while waiting for localStorage rehydration (one-frame flash
  // at most for synchronous storage) and while unauthenticated.
  if (!_hasHydrated || !isAuthenticated) {
    return null
  }

  return (
    <div className="flex min-h-screen bg-surface">
      <Sidebar />
      <main className="flex-1 ml-64 min-h-screen">
        {children}
      </main>
    </div>
  )
}
