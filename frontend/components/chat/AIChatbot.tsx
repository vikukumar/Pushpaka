'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { aiApi } from '@/lib/api'
import { Sparkles, X, Send, Loader2, Bot, User, ChevronDown, Minimize2 } from 'lucide-react'

interface Message {
  role: 'user' | 'assistant'
  content: string
  error?: boolean
}

interface AIChatbotProps {
  deploymentId?: string
}

const WELCOME = `Hi! I'm **Pushpaka Assistant** — your AI co-pilot for cloud deployments. ✨

I can help you with:
- 🔍 Diagnosing build failures and errors
- 🐳 Docker container configuration
- 🌿 Branch/webhook setup
- ⚙️ Environment variables & secrets
- 📊 Performance and log analysis

Ask me anything about your deployments!`

function renderMarkdown(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/`([^`]+)`/g, '<code class="bg-white/10 px-1 py-0.5 rounded text-xs font-mono text-cyan-300">$1</code>')
    .replace(/```([\s\S]*?)```/g, '<pre class="bg-black/40 rounded-lg p-3 text-xs font-mono text-green-300 overflow-x-auto my-2">$1</pre>')
    .replace(/^- (.+)$/gm, '<li class="ml-3 list-disc">$1</li>')
    .replace(/^(#{1,3}) (.+)$/gm, '<div class="font-semibold text-white mt-2">$2</div>')
    .replace(/\n\n/g, '<div class="h-2"></div>')
    .replace(/\n/g, '<br/>')
    .replace(/🔍|🐳|🌿|⚙️|📊|✨/g, (m) => `<span>${m}</span>`)
}

export default function AIChatbot({ deploymentId }: AIChatbotProps) {
  const [open, setOpen] = useState(false)
  const [minimized, setMinimized] = useState(false)
  const [messages, setMessages] = useState<Message[]>([
    { role: 'assistant', content: WELCOME },
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [pulse, setPulse] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (open && !minimized) {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
      setTimeout(() => inputRef.current?.focus(), 100)
    }
  }, [messages, open, minimized])

  // Subtle pulse on the button when closed to draw attention
  useEffect(() => {
    if (!open) {
      const t = setInterval(() => setPulse((v) => !v), 4000)
      return () => clearInterval(t)
    }
  }, [open])

  const send = useCallback(async () => {
    const msg = input.trim()
    if (!msg || loading) return

    setInput('')
    setMessages((prev) => [...prev, { role: 'user', content: msg }])
    setLoading(true)

    try {
      const res = await aiApi.chat(msg, deploymentId)
      const reply = res.data?.reply || 'No response from AI.'
      setMessages((prev) => [...prev, { role: 'assistant', content: reply }])
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      const errMsg = err?.response?.data?.error || 'AI service unavailable. Ensure AI_API_KEY is configured.'
      setMessages((prev) => [...prev, {
        role: 'assistant',
        content: errMsg,
        error: true,
      }])
    } finally {
      setLoading(false)
    }
  }, [input, loading, deploymentId])

  const handleKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      send()
    }
  }

  const quickActions = deploymentId
    ? ['Why did my build fail?', 'Show key errors', 'How to fix this?']
    : ['How do I set up webhooks?', 'How to configure resource limits?', 'Best practices for Dockerfile?']

  return (
    <>
      {/* Floating toggle button */}
      <button
        onClick={() => { setOpen((v) => !v); setMinimized(false) }}
        className="fixed bottom-6 right-6 z-50 w-14 h-14 rounded-full flex items-center justify-center shadow-2xl transition-all duration-300 hover:scale-110 active:scale-95"
        style={{
          background: open
            ? 'linear-gradient(135deg, #4f46e5, #7c3aed)'
            : 'linear-gradient(135deg, #4338ca, #6366f1, #818cf8)',
          boxShadow: open
            ? '0 8px 32px rgba(99,102,241,0.6), 0 0 0 1px rgba(255,255,255,0.1)'
            : pulse
            ? '0 8px 32px rgba(99,102,241,0.8), 0 0 40px rgba(99,102,241,0.4)'
            : '0 8px 32px rgba(99,102,241,0.5)',
        }}
        aria-label={open ? 'Close AI Assistant' : 'Open AI Assistant'}
      >
        {open
          ? <X size={22} className="text-white" />
          : <Sparkles size={22} className="text-white" />}

        {/* Unread indicator */}
        {!open && (
          <span
            className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-cyan-400 border-2 border-slate-900 text-[9px] font-bold flex items-center justify-center text-slate-900"
          >
            AI
          </span>
        )}
      </button>

      {/* Chat window */}
      {open && (
        <div
          className="fixed bottom-24 right-4 sm:right-6 z-50 flex flex-col rounded-2xl overflow-hidden shadow-2xl transition-all duration-300"
          style={{
            width: 'min(calc(100vw - 32px), 380px)',
            height: minimized ? '56px' : 'min(calc(100vh - 140px), 560px)',
            background: 'linear-gradient(175deg, #0e1c32 0%, #0b1524 100%)',
            border: '1px solid rgba(99,102,241,0.3)',
            boxShadow: '0 24px 80px rgba(0,0,0,0.7), 0 0 0 1px rgba(99,102,241,0.1)',
          }}
        >
          {/* Header */}
          <div
            className="flex items-center gap-3 px-4 py-3 shrink-0 cursor-pointer"
            style={{
              background: 'linear-gradient(90deg, rgba(99,102,241,0.18) 0%, rgba(124,58,237,0.12) 100%)',
              borderBottom: minimized ? 'none' : '1px solid rgba(99,102,241,0.2)',
            }}
            onClick={() => setMinimized((v) => !v)}
          >
            <div
              className="w-8 h-8 rounded-full flex items-center justify-center shrink-0"
              style={{
                background: 'linear-gradient(135deg, #4338ca, #6366f1)',
                boxShadow: '0 0 12px rgba(99,102,241,0.6)',
              }}
            >
              <Sparkles size={14} className="text-white" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="text-sm font-semibold text-white">Pushpaka Assistant</div>
              <div className="text-[10px] text-cyan-400 flex items-center gap-1">
                <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />
                AI-powered · Always on
              </div>
            </div>
            <div className="flex items-center gap-1">
              {minimized
                ? <ChevronDown size={14} className="text-slate-400 rotate-180" />
                : <Minimize2 size={14} className="text-slate-400" />}
            </div>
          </div>

          {!minimized && (
            <>
              {/* Messages */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {messages.map((msg, i) => (
                  <div
                    key={i}
                    className={`flex gap-2.5 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}
                  >
                    {/* Avatar */}
                    <div
                      className="w-7 h-7 rounded-full flex items-center justify-center shrink-0 mt-0.5"
                      style={{
                        background: msg.role === 'user'
                          ? 'linear-gradient(135deg, #0ea5e9, #06b6d4)'
                          : 'linear-gradient(135deg, #4338ca, #6366f1)',
                      }}
                    >
                      {msg.role === 'user'
                        ? <User size={12} className="text-white" />
                        : <Bot size={12} className="text-white" />}
                    </div>

                    {/* Bubble */}
                    <div
                      className={`max-w-[85%] rounded-2xl px-3 py-2 text-sm leading-relaxed ${
                        msg.role === 'user'
                          ? 'rounded-tr-sm'
                          : 'rounded-tl-sm'
                      } ${msg.error ? 'border border-red-500/30' : ''}`}
                      style={{
                        background: msg.role === 'user'
                          ? 'linear-gradient(135deg, rgba(14,165,233,0.25), rgba(6,182,212,0.2))'
                          : msg.error
                          ? 'rgba(239,68,68,0.1)'
                          : 'rgba(255,255,255,0.05)',
                        border: msg.role === 'user'
                          ? '1px solid rgba(14,165,233,0.3)'
                          : msg.error
                          ? '1px solid rgba(239,68,68,0.3)'
                          : '1px solid rgba(255,255,255,0.08)',
                        color: msg.error ? '#fca5a5' : 'var(--text-primary)',
                      }}
                    >
                      <div
                        dangerouslySetInnerHTML={{ __html: renderMarkdown(msg.content) }}
                      />
                    </div>
                  </div>
                ))}

                {loading && (
                  <div className="flex gap-2.5">
                    <div
                      className="w-7 h-7 rounded-full flex items-center justify-center shrink-0"
                      style={{ background: 'linear-gradient(135deg, #4338ca, #6366f1)' }}
                    >
                      <Bot size={12} className="text-white" />
                    </div>
                    <div
                      className="rounded-2xl rounded-tl-sm px-4 py-3 flex items-center gap-2"
                      style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.08)' }}
                    >
                      <Loader2 size={12} className="animate-spin text-brand-400" />
                      <span className="text-xs text-slate-400">Thinking…</span>
                    </div>
                  </div>
                )}
                <div ref={bottomRef} />
              </div>

              {/* Quick actions */}
              {messages.length === 1 && !loading && (
                <div className="px-4 pb-2 flex flex-wrap gap-1.5">
                  {quickActions.map((q) => (
                    <button
                      key={q}
                      onClick={() => { setInput(q); setTimeout(() => inputRef.current?.focus(), 50) }}
                      className="text-xs px-2.5 py-1 rounded-full transition-colors hover:text-white"
                      style={{
                        background: 'rgba(99,102,241,0.12)',
                        border: '1px solid rgba(99,102,241,0.25)',
                        color: '#a5b4fc',
                      }}
                    >
                      {q}
                    </button>
                  ))}
                </div>
              )}

              {/* Input */}
              <div
                className="shrink-0 p-3"
                style={{ borderTop: '1px solid rgba(255,255,255,0.07)' }}
              >
                <div
                  className="flex items-center gap-2 rounded-xl px-3 py-2"
                  style={{
                    background: 'rgba(255,255,255,0.04)',
                    border: '1px solid rgba(99,102,241,0.25)',
                  }}
                >
                  <input
                    ref={inputRef}
                    type="text"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={handleKey}
                    placeholder="Ask anything about your deployment…"
                    className="flex-1 bg-transparent text-sm text-white placeholder-slate-600 outline-none"
                    disabled={loading}
                  />
                  <button
                    onClick={send}
                    disabled={loading || !input.trim()}
                    className="w-7 h-7 rounded-lg flex items-center justify-center transition-all disabled:opacity-40 disabled:cursor-not-allowed hover:scale-110"
                    style={{
                      background: 'linear-gradient(135deg, #4338ca, #6366f1)',
                      boxShadow: '0 2px 8px rgba(99,102,241,0.4)',
                    }}
                  >
                    <Send size={12} className="text-white" />
                  </button>
                </div>
                <p className="text-[10px] text-slate-700 text-center mt-1.5">
                  Powered by AI · Set AI_API_KEY to enable
                </p>
              </div>
            </>
          )}
        </div>
      )}
    </>
  )
}
