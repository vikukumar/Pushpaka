'use client'

import { useEffect, useState, createContext, useContext } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import Link from 'next/link'
import { useAuthStore } from '@/lib/auth'
import { Sidebar } from '@/components/layout/Sidebar'
import { LayoutDashboard, FolderGit2, Rocket, Settings } from 'lucide-react'

interface SidebarCtx {
  open: boolean
  toggle: () => void
  close: () => void
}

export const SidebarContext = createContext<SidebarCtx>({
  open: false,
  toggle: () => { },
  close: () => { },
})

export function useSidebar() {
  return useContext(SidebarContext)
}

const mobileNavItems = [
  { href: '/dashboard', icon: LayoutDashboard, label: 'Home' },
  { href: '/dashboard/projects', icon: FolderGit2, label: 'Projects' },
  { href: '/dashboard/deployments', icon: Rocket, label: 'Deploy' },
  { href: '/dashboard/settings', icon: Settings, label: 'Settings' },
]

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, token, clearAuth, _hasHydrated } = useAuthStore()
  const router = useRouter()
  const pathname = usePathname()
  const [sidebarOpen, setSidebarOpen] = useState(false)

  // Validate token expiration
  useEffect(() => {
    if (_hasHydrated && isAuthenticated && token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]))
        const isExpired = payload.exp * 1000 < Date.now()
        if (isExpired) {
          console.warn('[Auth] Token expired, clearing session...')
          clearAuth()
          router.replace('/login')
        }
      } catch (e) {
        console.error('[Auth] Malformed token, clearing session...')
        clearAuth()
        router.replace('/login')
      }
    }
  }, [_hasHydrated, isAuthenticated, token, clearAuth, router])

  useEffect(() => {
    if (_hasHydrated && !isAuthenticated) {
      router.replace('/login')
    }
  }, [_hasHydrated, isAuthenticated, router])

  if (!_hasHydrated || !isAuthenticated) {
    return null
  }

  return (
    <SidebarContext.Provider value={{
      open: sidebarOpen,
      toggle: () => setSidebarOpen((v) => !v),
      close: () => setSidebarOpen(false),
    }}>
      <div className="flex min-h-screen bg-surface">
        {/* Mobile overlay backdrop */}
        {sidebarOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm md:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}

        <Sidebar />

        <main className="flex-1 min-h-screen md:ml-64 flex flex-col">
          <div className="flex-1 pb-16 md:pb-0">{children}</div>
          {/* Footer */}
          <footer
            className="hidden md:flex items-center justify-center gap-1.5 py-3 text-[11px] select-none"
            style={{ borderTop: '1px solid rgba(255,255,255,0.05)', color: 'var(--text-muted, #475569)' }}
          >
            <span>Made with</span>
            <span title="India">🇮🇳</span>
            <span>and</span>
            <span title="love">❤️</span>
            <span>by</span>
            <a
              href="https://github.com/vikukumar"
              target="_blank"
              rel="noopener noreferrer"
              className="font-semibold hover:text-brand-400 transition-colors"
              style={{ color: '#818cf8' }}
            >
              Vikash Kumar
            </a>
            <span className="mx-1 opacity-40">·</span>
            <span>© {new Date().getFullYear()} Pushpaka</span>
          </footer>
        </main>

        {/* Mobile bottom navigation */}
        <nav
          className="md:hidden fixed bottom-0 left-0 right-0 z-30 flex items-center justify-around px-2 py-2"
          style={{
            background: 'linear-gradient(0deg, rgba(11,17,32,0.98) 0%, rgba(14,22,38,0.97) 100%)',
            borderTop: '1px solid rgba(99,102,241,0.2)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
          }}
        >
          {mobileNavItems.map(({ href, icon: Icon, label }) => {
            const isActive = href === '/dashboard'
              ? pathname === href
              : pathname.startsWith(href)
            return (
              <Link
                key={href}
                href={href}
                className="flex flex-col items-center gap-0.5 px-3 py-1 rounded-xl transition-all min-w-0"
                style={{
                  color: isActive ? '#818cf8' : '#475569',
                  background: isActive ? 'rgba(99,102,241,0.1)' : 'transparent',
                }}
              >
                <Icon size={20} />
                <span className="text-[10px] font-medium">{label}</span>
              </Link>
            )
          })}
        </nav>
      </div>
    </SidebarContext.Provider>
  )
}
