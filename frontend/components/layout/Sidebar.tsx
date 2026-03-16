'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  FolderGit2,
  Rocket,
  Globe,
  Settings,
  Activity,
  LogOut,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/lib/auth'
import { useRouter } from 'next/navigation'

const navItems = [
  { href: '/dashboard',             label: 'Overview',     icon: LayoutDashboard },
  { href: '/dashboard/projects',    label: 'Projects',     icon: FolderGit2 },
  { href: '/dashboard/deployments', label: 'Deployments',  icon: Rocket },
  { href: '/dashboard/domains',     label: 'Domains',      icon: Globe },
  { href: '/dashboard/activity',    label: 'Activity',     icon: Activity },
  { href: '/dashboard/settings',    label: 'Settings',     icon: Settings },
]

export function Sidebar() {
  const pathname = usePathname()
  const { user, clearAuth } = useAuthStore()
  const router = useRouter()

  const handleLogout = () => {
    clearAuth()
    router.push('/login')
  }

  return (
    <aside
      className="w-64 h-screen flex flex-col fixed left-0 top-0 z-30 overflow-hidden"
      style={{
        background: 'linear-gradient(175deg, #0e1c32 0%, #091520 60%, #0c1b2e 100%)',
        borderRight: '1px solid rgba(99,102,241,0.12)',
        boxShadow: '4px 0 32px rgba(0,0,0,0.5), 1px 0 0 rgba(99,102,241,0.05)',
      }}
    >
      {/* Ambient top-left glow orb */}
      <div
        className="absolute -top-16 -left-16 w-56 h-56 rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(99,102,241,0.09) 0%, transparent 70%)' }}
      />

      {/* â”€â”€â”€ Logo â”€â”€â”€ */}
      <div
        className="px-5 py-4 relative shrink-0"
        style={{ borderBottom: '1px solid rgba(99,102,241,0.1)' }}
      >
        <Link href="/dashboard" className="flex items-center gap-3 group">
          <div
            className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0"
            style={{
              background: 'linear-gradient(135deg, #3730a3 0%, #4f46e5 55%, #7c3aed 100%)',
              boxShadow: '0 4px 16px rgba(99,102,241,0.45), inset 0 1px 0 rgba(255,255,255,0.2)',
            }}
          >
            <svg viewBox="0 0 32 32" className="w-5 h-5" fill="none">
              <rect x="3" y="15" width="26" height="10" rx="5" fill="white" fillOpacity="0.95"/>
              <circle cx="8"  cy="15" r="4" fill="white" fillOpacity="0.95"/>
              <circle cx="16" cy="12" r="5" fill="white" fillOpacity="0.95"/>
              <circle cx="24" cy="14" r="4" fill="white" fillOpacity="0.95"/>
              <path d="M3 18 Q0 16 1 20 Q0 23 3 22"   fill="#22d3ee"/>
              <path d="M29 18 Q32 16 31 20 Q32 23 29 22" fill="#22d3ee"/>
              <line x1="16" y1="4" x2="16" y2="9"
                stroke="rgba(255,255,255,0.6)" strokeWidth="1.5" strokeLinecap="round"/>
              <circle cx="16" cy="3.5" r="1.5" fill="#22d3ee"/>
            </svg>
          </div>
          <div className="min-w-0">
            <div
              className="font-bold tracking-tight text-[15px] leading-none"
              style={{
                background: 'linear-gradient(90deg, #a5b4fc 0%, #818cf8 50%, #c4b5fd 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
              }}
            >
              Pushpaka
            </div>
            <div className="text-[9px] text-slate-600 mt-1 tracking-[0.18em] uppercase font-semibold">
              Platform
            </div>
          </div>
        </Link>
      </div>

      {/* â”€â”€â”€ Navigation â”€â”€â”€ */}
      <nav className="flex-1 p-3 space-y-0.5 overflow-y-auto">
        {navItems.map(({ href, label, icon: Icon }) => {
          // /dashboard exact match to avoid marking Overview active on sub-routes
          const isActive =
            href === '/dashboard'
              ? pathname === href
              : pathname === href || pathname.startsWith(href + '/')
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                'relative flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium',
                'transition-all duration-200 group overflow-hidden',
                isActive ? 'text-white' : 'text-slate-500 hover:text-slate-300'
              )}
              style={
                isActive
                  ? {
                      background:
                        'linear-gradient(90deg, rgba(99,102,241,0.22) 0%, rgba(99,102,241,0.07) 55%, transparent 100%)',
                      boxShadow:
                        'inset 3px 0 0 #818cf8, inset 0 1px 0 rgba(255,255,255,0.04)',
                    }
                  : undefined
              }
            >
              {/* Hover fill */}
              {!isActive && (
                <span className="absolute inset-0 opacity-0 group-hover:opacity-100 rounded-lg transition-opacity duration-200"
                  style={{ background: 'rgba(255,255,255,0.025)' }}
                />
              )}

              <Icon
                size={15}
                className={cn(
                  'shrink-0 transition-colors',
                  isActive ? 'text-brand-400' : 'text-slate-600 group-hover:text-slate-400'
                )}
                style={isActive ? { filter: 'drop-shadow(0 0 5px rgba(129,140,248,0.75))' } : undefined}
              />
              <span className="truncate flex-1">{label}</span>

              {/* Active indicator dot */}
              {isActive && (
                <span
                  className="w-1.5 h-1.5 rounded-full bg-brand-400 shrink-0"
                  style={{ boxShadow: '0 0 8px rgba(129,140,248,0.9)' }}
                />
              )}
            </Link>
          )
        })}
      </nav>

      {/* â”€â”€â”€ User section â”€â”€â”€ */}
      <div
        className="p-3 relative shrink-0"
        style={{ borderTop: '1px solid rgba(99,102,241,0.1)' }}
      >
        <div
          className="flex items-center gap-3 px-3 py-2.5 rounded-lg cursor-default"
          style={{ background: 'rgba(255,255,255,0.025)', border: '1px solid rgba(99,102,241,0.07)' }}
        >
          <div
            className="w-8 h-8 rounded-lg flex items-center justify-center text-white text-xs font-bold shrink-0"
            style={{
              background: 'linear-gradient(135deg, #4338ca, #6366f1)',
              boxShadow: '0 2px 8px rgba(99,102,241,0.4)',
            }}
          >
            {user?.name?.[0]?.toUpperCase() || 'U'}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-slate-200 truncate">{user?.name}</div>
            <div className="text-[11px] text-slate-600 truncate">{user?.email}</div>
          </div>
          <button
            onClick={handleLogout}
            className="text-slate-600 hover:text-red-400 transition-colors p-1 rounded shrink-0"
            title="Sign out"
          >
            <LogOut size={14} />
          </button>
        </div>
      </div>
    </aside>
  )
}

