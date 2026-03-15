'use client'

import { Bell, Moon, Sun, Search } from 'lucide-react'
import { useTheme } from 'next-themes'

interface HeaderProps {
  title: string
  subtitle?: string
}

export function Header({ title, subtitle }: HeaderProps) {
  const { theme, setTheme } = useTheme()

  return (
    <header className="h-16 border-b border-surface-border flex items-center justify-between px-6 bg-surface/80 backdrop-blur-sm sticky top-0 z-20">
      <div>
        <h1 className="text-lg font-semibold text-white">{title}</h1>
        {subtitle && <p className="text-xs text-slate-500">{subtitle}</p>}
      </div>

      <div className="flex items-center gap-3">
        {/* Search */}
        <div className="relative hidden md:block">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Search..."
            className="pl-8 pr-4 py-1.5 text-sm bg-surface border border-surface-border rounded-lg text-slate-300 placeholder:text-slate-600 focus:outline-none focus:ring-1 focus:ring-brand-500 w-44"
          />
        </div>

        {/* Theme toggle */}
        <button
          onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          className="p-2 text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
        >
          {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
        </button>

        {/* Notifications */}
        <button className="p-2 text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors relative">
          <Bell size={16} />
          <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 bg-brand-500 rounded-full" />
        </button>
      </div>
    </header>
  )
}
