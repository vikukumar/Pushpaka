'use client'

import { Bell, Moon, Sun } from 'lucide-react'
import { useTheme } from 'next-themes'
import { useEffect, useState } from 'react'

interface HeaderProps {
  title: string
  subtitle?: string
}

export function Header({ title, subtitle }: HeaderProps) {
  const { resolvedTheme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  // Avoid hydration mismatch — only render theme-dependent icon after mount
  useEffect(() => { setMounted(true) }, [])

  return (
    <header
      className="h-16 flex items-center justify-between px-6 sticky top-0 z-20"
      style={{
        background: 'linear-gradient(180deg, rgba(11,17,32,0.97) 0%, rgba(9,15,26,0.93) 100%)',
        borderBottom: '1px solid rgba(99,102,241,0.12)',
        backdropFilter: 'blur(14px)',
        WebkitBackdropFilter: 'blur(14px)',
        boxShadow: '0 1px 0 rgba(99,102,241,0.08), 0 4px 24px -4px rgba(0,0,0,0.5)',
      }}
    >
      {/* Title */}
      <div>
        <h1
          className="text-[17px] font-semibold tracking-tight leading-tight"
          style={{
            background: 'linear-gradient(90deg, #f1f5f9 0%, #94a3b8 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            backgroundClip: 'text',
          }}
        >
          {title}
        </h1>
        {subtitle && (
          <p className="text-[11px] text-slate-600 mt-0.5 tracking-wide">{subtitle}</p>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        {/* Theme toggle — rendered after mount to avoid SSR mismatch */}
        {mounted && (
          <button
            onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
            className="p-2 rounded-lg transition-all duration-200 text-slate-500 hover:text-slate-200"
            style={{
              background: 'rgba(255,255,255,0.03)',
              border: '1px solid rgba(99,102,241,0.12)',
              boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.04)',
            }}
            title={resolvedTheme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {resolvedTheme === 'dark' ? <Sun size={15} /> : <Moon size={15} />}
          </button>
        )}

        {/* Notifications */}
        <button
          className="relative p-2 rounded-lg transition-all duration-200 text-slate-500 hover:text-slate-200"
          style={{
            background: 'rgba(255,255,255,0.03)',
            border: '1px solid rgba(99,102,241,0.12)',
            boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.04)',
          }}
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
