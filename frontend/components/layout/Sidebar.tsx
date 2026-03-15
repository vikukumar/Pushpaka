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
  ChevronRight,
  LogOut,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/lib/auth'
import { useRouter } from 'next/navigation'

const navItems = [
  { href: '/dashboard', label: 'Overview', icon: LayoutDashboard },
  { href: '/dashboard/projects', label: 'Projects', icon: FolderGit2 },
  { href: '/dashboard/deployments', label: 'Deployments', icon: Rocket },
  { href: '/dashboard/domains', label: 'Domains', icon: Globe },
  { href: '/dashboard/activity', label: 'Activity', icon: Activity },
  { href: '/dashboard/settings', label: 'Settings', icon: Settings },
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
    <aside className="w-64 h-screen bg-surface-elevated border-r border-surface-border flex flex-col fixed left-0 top-0 z-30">
      {/* Logo */}
      <div className="p-5 border-b border-surface-border">
        <Link href="/dashboard" className="flex items-center gap-3">
          <div className="w-8 h-8 bg-brand-600 rounded-lg flex items-center justify-center">
            <svg viewBox="0 0 32 32" className="w-5 h-5" fill="none">
              <rect x="3" y="15" width="26" height="10" rx="5" fill="white" fillOpacity="0.9"/>
              <circle cx="8" cy="15" r="4" fill="white" fillOpacity="0.9"/>
              <circle cx="16" cy="12" r="5" fill="white" fillOpacity="0.9"/>
              <circle cx="24" cy="14" r="4" fill="white" fillOpacity="0.9"/>
              <path d="M3 18 Q0 16 1 20 Q0 23 3 22" fill="#22d3ee" opacity="0.9"/>
              <path d="M29 18 Q32 16 31 20 Q32 23 29 22" fill="#22d3ee" opacity="0.9"/>
              <line x1="16" y1="4" x2="16" y2="9" stroke="rgba(255,255,255,0.7)" strokeWidth="1.5" strokeLinecap="round"/>
              <circle cx="16" cy="3.5" r="1.5" fill="#22d3ee"/>
            </svg>
          </div>
          <div>
            <span className="font-bold text-white">Pushpaka</span>
            <span className="block text-[10px] text-slate-500 leading-none">v1.0.0</span>
          </div>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-3 space-y-0.5 overflow-y-auto">
        {navItems.map(({ href, label, icon: Icon }) => {
          const isActive = pathname === href || pathname.startsWith(href + '/')
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all group',
                isActive
                  ? 'bg-brand-600/15 text-brand-300 border border-brand-500/20'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
              )}
            >
              <Icon size={16} className={cn(isActive ? 'text-brand-400' : 'text-slate-500 group-hover:text-slate-400')} />
              {label}
              {isActive && <ChevronRight size={14} className="ml-auto text-brand-400" />}
            </Link>
          )
        })}
      </nav>

      {/* User section */}
      <div className="p-3 border-t border-surface-border">
        <div className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-slate-800 transition-colors">
          <div className="w-8 h-8 bg-brand-600 rounded-full flex items-center justify-center text-white text-xs font-bold">
            {user?.name?.[0]?.toUpperCase() || 'U'}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-slate-200 truncate">{user?.name}</div>
            <div className="text-xs text-slate-500 truncate">{user?.email}</div>
          </div>
          <button
            onClick={handleLogout}
            className="text-slate-500 hover:text-red-400 transition-colors"
            title="Sign out"
          >
            <LogOut size={15} />
          </button>
        </div>
      </div>
    </aside>
  )
}
