'use client'

import { useState, useRef, useEffect, useId } from 'react'
import { ChevronDown, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface SelectOption {
  value: string
  label: string
}

interface SelectProps {
  value: string
  onChange: (value: string) => void
  options: SelectOption[]
  placeholder?: string
  className?: string
  disabled?: boolean
}

export function Select({ value, onChange, options, placeholder, className, disabled }: SelectProps) {
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const id = useId()

  const selected = options.find((o) => o.value === value)

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  // Close on Escape
  useEffect(() => {
    if (!open) return
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [open])

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <button
        type="button"
        id={id}
        disabled={disabled}
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'input flex items-center justify-between gap-2 text-left cursor-pointer select-none',
          disabled && 'opacity-50 cursor-not-allowed',
          open && 'border-brand-500/60 shadow-[inset_0_2px_4px_rgba(0,0,0,0.1),0_0_0_3px_rgba(99,102,241,0.12)]',
        )}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        <span className={selected ? 'text-[var(--input-color)]' : 'text-[var(--input-placeholder)]'}>
          {selected ? selected.label : (placeholder ?? 'Select…')}
        </span>
        <ChevronDown
          size={14}
          className={cn(
            'shrink-0 text-slate-500 transition-transform duration-150',
            open && 'rotate-180',
          )}
        />
      </button>

      {open && (
        <div
          role="listbox"
          aria-labelledby={id}
          className="absolute z-50 mt-1 w-full rounded-lg overflow-hidden py-0.5"
          style={{
            background: 'var(--select-dropdown-bg, #0f172a)',
            border: '1px solid rgba(99,102,241,0.3)',
            boxShadow: '0 8px 32px rgba(0,0,0,0.65), 0 2px 8px rgba(0,0,0,0.4)',
          }}
        >
          {options.map((opt) => {
            const isSelected = opt.value === value
            return (
              <button
                key={opt.value}
                role="option"
                aria-selected={isSelected}
                type="button"
                onClick={() => {
                  onChange(opt.value)
                  setOpen(false)
                }}
                style={{
                  background: isSelected ? 'rgba(99,102,241,0.18)' : 'transparent',
                }}
                className={cn(
                  'w-full flex items-center justify-between gap-2 px-3 py-2 text-sm text-left transition-colors',
                  isSelected
                    ? 'text-brand-300 hover:bg-[rgba(99,102,241,0.25)]'
                    : 'text-slate-300 hover:bg-[rgba(255,255,255,0.07)] hover:text-white',
                )}
              >
                <span>{opt.label}</span>
                {isSelected && <Check size={13} className="text-brand-400 shrink-0" />}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
