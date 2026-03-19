'use client'

import { Bell, Sun, Moon, Menu } from 'lucide-react'
import React from 'react'
import { useTheme } from '@/lib/theme'
import { useSidebar } from '@/components/providers/AuthProvider'

interface HeaderProps {
  title: string
  subtitle?: string
  actions?: React.ReactNode
}

export function Header({ title, subtitle, actions }: HeaderProps) {
  const { theme, toggle, mounted } = useTheme()
  const isDark = !mounted || theme === 'dark'
  const { toggle: toggleSidebar } = useSidebar()

  return (
    <header
      className="h-14 md:h-16 flex items-center justify-between px-4 md:px-6 sticky top-0 z-20 transition-all duration-300 border-b border-[var(--header-border)] shadow-sm bg-[var(--header-bg)] backdrop-blur-md"
    >
      {/* Hamburger (mobile only) + Title */}
      <div className="flex items-center gap-3">
        <button
          onClick={toggleSidebar}
          className="md:hidden p-2 rounded-lg text-slate-500 hover:text-[var(--text-primary)] hover:bg-[var(--border-subtle)] transition-colors"
          aria-label="Open menu"
        >
          <Menu size={18} />
        </button>
        <div>
          <h1
            className="text-[15px] md:text-[17px] font-semibold tracking-tight leading-tight"
            style={{
              background: 'var(--header-title)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              backgroundClip: 'text',
            }}
          >
            {title}
          </h1>
          {subtitle && (
            <p
              className="text-[11px] mt-0.5 tracking-wide"
              style={{ color: 'var(--text-muted)' }}
            >
              {subtitle}
            </p>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        {actions && <div className="flex items-center gap-2 mr-1">{actions}</div>}
        {/* Animated theme toggle pill */}
        <button
          onClick={toggle}
          aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
          className="relative flex items-center rounded-full transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-brand-500/40"
          style={{
            width: '56px',
            height: '28px',
            padding: '3px',
            background: isDark
              ? 'linear-gradient(135deg, #1a2844 0%, #111c30 100%)'
              : 'linear-gradient(135deg, #e0e7ff 0%, #c7d2fe 100%)',
            border: isDark
              ? '1px solid rgba(129,140,248,0.25)'
              : '1px solid rgba(99,102,241,0.3)',
            boxShadow: isDark
              ? 'inset 0 1px 0 rgba(255,255,255,0.04), 0 0 12px rgba(99,102,241,0.15)'
              : 'inset 0 1px 0 rgba(255,255,255,0.6), 0 2px 8px rgba(99,102,241,0.15)',
          }}
        >
          {/* Sliding thumb */}
          <span
            className="absolute flex items-center justify-center rounded-full transition-all duration-300"
            style={{
              width: '22px',
              height: '22px',
              top: '3px',
              left: isDark ? '3px' : '31px',
              background: isDark
                ? 'linear-gradient(135deg, #4338ca 0%, #6366f1 100%)'
                : 'linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%)',
              boxShadow: isDark
                ? '0 2px 8px rgba(99,102,241,0.6), 0 0 0 1px rgba(255,255,255,0.1)'
                : '0 2px 8px rgba(251,191,36,0.6), 0 0 0 1px rgba(255,255,255,0.5)',
            }}
          >
            {isDark
              ? <Moon size={11} className="text-white" />
              : <Sun size={11} className="text-white" />
            }
          </span>
          {/* Background icons */}
          <Sun
            size={11}
            className="absolute transition-opacity duration-200"
            style={{
              right: '7px',
              top: '50%',
              transform: 'translateY(-50%)',
              color: isDark ? '#475569' : 'transparent',
              opacity: isDark ? 1 : 0,
            }}
          />
          <Moon
            size={11}
            className="absolute transition-opacity duration-200"
            style={{
              left: '7px',
              top: '50%',
              transform: 'translateY(-50%)',
              color: isDark ? 'transparent' : '#6366f1',
              opacity: isDark ? 0 : 1,
            }}
          />
        </button>

        {/* Notifications - hidden on small mobile */}
        <button
          className="relative p-2 rounded-lg transition-all duration-200 hidden sm:flex"
          style={{
            color: 'var(--text-muted)',
            background: isDark ? 'rgba(255,255,255,0.03)' : 'rgba(99,102,241,0.05)',
            border: '1px solid var(--border-subtle)',
          }}
          onMouseEnter={e => (e.currentTarget.style.color = 'var(--text-primary)')}
          onMouseLeave={e => (e.currentTarget.style.color = 'var(--text-muted)')}
        >
          <Bell size={15} />
          <span
            className="absolute top-1.5 right-1.5 w-1.5 h-1.5 rounded-full"
            style={{ background: '#818cf8', boxShadow: '0 0 6px rgba(129,140,248,0.9)' }}
          />
        </button>
      </div>
    </header>
  )
}

