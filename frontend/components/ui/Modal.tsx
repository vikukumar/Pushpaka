'use client'

import { useState, useEffect, useCallback, ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { X, AlertCircle, HelpCircle, Info, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  children: ReactNode
  footer?: ReactNode
  type?: 'info' | 'warn' | 'error' | 'confirm'
  maxWidth?: string
}

export function Modal({
  isOpen,
  onClose,
  title,
  children,
  footer,
  type = 'info',
  maxWidth = 'max-w-md'
}: ModalProps) {
  const [mounted, setMounted] = useState(false)
  const [shouldRender, setShouldRender] = useState(isOpen)

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    if (isOpen) {
      setShouldRender(true)
      document.body.style.overflow = 'hidden'
    } else {
      const timer = setTimeout(() => {
        setShouldRender(false)
        document.body.style.overflow = ''
      }, 300)
      return () => clearTimeout(timer)
    }
  }, [isOpen])

  if (!mounted || !shouldRender) return null

  const icons = {
    info: <Info className="text-blue-400" size={20} />,
    warn: <AlertCircle className="text-amber-400" size={20} />,
    error: <AlertCircle className="text-red-400" size={20} />,
    confirm: <HelpCircle className="text-brand-400" size={20} />,
  }

  return createPortal(
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6">
      {/* Backdrop */}
      <div
        className={cn(
          "absolute inset-0 bg-slate-950/60 backdrop-blur-sm transition-opacity duration-300 ease-out",
          isOpen ? "opacity-100" : "opacity-0"
        )}
        onClick={onClose}
      />

      {/* Modal Container */}
      <div
        className={cn(
          "relative w-full bg-slate-900 border border-white/10 rounded-xl shadow-2xl overflow-hidden transition-all duration-300 cubic-bezier(0.16, 1, 0.3, 1)",
          maxWidth,
          isOpen ? "opacity-100 scale-100 translate-y-0" : "opacity-0 scale-95 translate-y-8"
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-800 bg-slate-900/50">
          <div className="flex items-center gap-3">
            {icons[type]}
            <h3 className="text-sm font-semibold text-white tracking-tight">{title}</h3>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Content */}
        <div className="px-6 py-5 text-sm text-slate-300 leading-relaxed">
          {children}
        </div>

        {/* Footer */}
        {footer && (
          <div className="px-6 py-4 bg-slate-900/80 border-t border-slate-800 flex justify-end gap-3">
            {footer}
          </div>
        )}
      </div>
    </div>,
    document.body
  )
}

// ─── Promise-based Confirmation ──────────────────────────────────────────────

interface ConfirmOptions {
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  type?: 'info' | 'warn' | 'error' | 'confirm'
}

export function useConfirm() {
  const [state, setState] = useState<ConfirmOptions | null>(null)
  const [resolver, setResolver] = useState<((v: boolean) => void) | null>(null)

  const confirm = useCallback((options: ConfirmOptions) => {
    setState(options)
    return new Promise<boolean>((resolve) => {
      setResolver(() => resolve)
    })
  }, [])

  const handleClose = useCallback((value: boolean) => {
    resolver?.(value)
    setState(null)
  }, [resolver])

  const Component = state ? (
    <Modal
      isOpen={!!state}
      onClose={() => handleClose(false)}
      title={state.title}
      type={state.type || 'confirm'}
      footer={
        <>
          <button
            onClick={() => handleClose(false)}
            className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-white transition-colors"
          >
            {state.cancelText || 'Cancel'}
          </button>
          <button
            onClick={() => handleClose(true)}
            className={cn(
              "px-4 py-2 text-xs font-semibold rounded-lg shadow-lg transition-all transform active:scale-95",
              state.type === 'error' ? "bg-red-600 hover:bg-red-500 text-white" :
              state.type === 'warn' ? "bg-amber-600 hover:bg-amber-500 text-white" :
              "bg-brand-600 hover:bg-brand-500 text-white shadow-brand-900/20"
            )}
          >
            {state.confirmText || 'Confirm'}
          </button>
        </>
      }
    >
      {state.message}
    </Modal>
  ) : null

  return { confirm, Component }
}

// ─── Promise-based Prompt ────────────────────────────────────────────────────

interface PromptOptions {
  title: string
  message: string
  defaultValue?: string
  placeholder?: string
  confirmText?: string
  cancelText?: string
}

export function usePrompt() {
  const [state, setState] = useState<PromptOptions | null>(null)
  const [value, setValue] = useState('')
  const [resolver, setResolver] = useState<((v: string | null) => void) | null>(null)

  const prompt = useCallback((options: PromptOptions) => {
    setState(options)
    setValue(options.defaultValue || '')
    return new Promise<string | null>((resolve) => {
      setResolver(() => resolve)
    })
  }, [])

  const handleClose = useCallback((v: string | null) => {
    resolver?.(v)
    setState(null)
  }, [resolver])

  const Component = state ? (
    <Modal
      isOpen={!!state}
      onClose={() => handleClose(null)}
      title={state.title}
      type="info"
      footer={
        <>
          <button
            onClick={() => handleClose(null)}
            className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-white transition-colors"
          >
            {state.cancelText || 'Cancel'}
          </button>
          <button
            onClick={() => handleClose(value)}
            className="px-4 py-2 text-xs font-semibold rounded-lg bg-brand-600 hover:bg-brand-500 text-white shadow-lg shadow-brand-500/20 transition-all transform active:scale-95 group"
          >
            <span className="flex items-center gap-2">
              {state.confirmText || 'Continue'}
              <ChevronRight size={14} className="group-hover:translate-x-0.5 transition-transform" />
            </span>
          </button>
        </>
      }
    >
      <div className="space-y-3">
        <p>{state.message}</p>
        <input
          autoFocus
          className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-brand-500/50 transition-all"
          placeholder={state.placeholder}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleClose(value)}
        />
      </div>
    </Modal>
  ) : null

  return { prompt, Component }
}
