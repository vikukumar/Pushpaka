'use client'

import { useState, useRef, useEffect } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { aiApi } from '@/lib/api'
import { PageHeader } from '@/components/ui/PageHeader'
import { BotMessageSquare, Send, Loader2, RefreshCw, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import Link from 'next/link'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  ts: Date
}

export default function AIChatPage() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content:
        "Hello! I'm your Pushpaka support agent. I can help you troubleshoot deployments, analyze logs, answer questions about your projects, and guide you through platform features. How can I assist you today?",
      ts: new Date(),
    },
  ])
  const [input, setInput] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('')
  const [showPromptEditor, setShowPromptEditor] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  const { data: configData } = useQuery({
    queryKey: ['ai-config'],
    queryFn: () => aiApi.getConfig(),
  })

  useEffect(() => {
    if (configData?.data?.system_prompt) {
      setSystemPrompt(configData.data.system_prompt)
    }
  }, [configData])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const { mutate: sendMessage, isPending } = useMutation({
    mutationFn: (text: string) => aiApi.chat(text),
    onSuccess: (res) => {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          role: 'assistant',
          content: res.data?.reply ?? res.data?.message ?? 'No response received.',
          ts: new Date(),
        },
      ])
    },
    onError: () => {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          role: 'assistant',
          content: 'Sorry, I encountered an error. Please check your AI configuration in Settings.',
          ts: new Date(),
        },
      ])
    },
  })

  const handleSend = () => {
    const text = input.trim()
    if (!text || isPending) return
    setMessages((prev) => [
      ...prev,
      { id: Date.now().toString(), role: 'user', content: text, ts: new Date() },
    ])
    setInput('')
    sendMessage(text)
  }

  return (
    <div className="flex flex-col h-screen md:h-[calc(100vh-0px)]">
      <PageHeader
        title="Support Agent"
        description="AI-powered assistant for your platform"
        icon={<BotMessageSquare className="text-brand-400" size={22} />}
        actions={
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowPromptEditor((v) => !v)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 transition-colors"
              style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.08)' }}
            >
              <Settings size={13} />
              System Prompt
            </button>
            <button
              onClick={() => setMessages(messages.slice(0, 1))}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 transition-colors"
              style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.08)' }}
            >
              <RefreshCw size={13} />
              Clear
            </button>
            <Link
              href="/dashboard/settings#ai"
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
              style={{ background: 'rgba(99,102,241,0.15)', border: '1px solid rgba(99,102,241,0.3)', color: '#a5b4fc' }}
            >
              AI Settings
            </Link>
          </div>
        }
      />

      <div className="flex flex-col flex-1 min-h-0 px-4 md:px-6 pb-4 gap-3">
        {/* System prompt editor */}
        {showPromptEditor && (
          <div
            className="rounded-xl p-4"
            style={{ background: 'rgba(99,102,241,0.06)', border: '1px solid rgba(99,102,241,0.2)' }}
          >
            <p className="text-xs font-semibold text-brand-400 mb-2 uppercase tracking-wider">Custom System Prompt</p>
            <textarea
              value={systemPrompt}
              onChange={(e) => setSystemPrompt(e.target.value)}
              rows={4}
              className="w-full rounded-lg px-3 py-2 text-sm text-slate-300 resize-none focus:outline-none focus:ring-2 focus:ring-brand-500/50"
              style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.08)' }}
              placeholder="You are a helpful DevOps assistant for Pushpaka PaaS..."
            />
            <p className="text-[11px] text-slate-600 mt-1">Save your system prompt in <Link href="/dashboard/settings#ai" className="text-brand-400 hover:underline">AI Settings</Link> to persist it.</p>
          </div>
        )}

        {/* Messages */}
        <div className="flex-1 min-h-0 overflow-y-auto space-y-4 pr-1">
          {messages.map((msg) => (
            <div key={msg.id} className={cn('flex gap-3', msg.role === 'user' ? 'flex-row-reverse' : 'flex-row')}>
              <div
                className="w-8 h-8 rounded-full shrink-0 flex items-center justify-center text-xs font-bold"
                style={
                  msg.role === 'assistant'
                    ? { background: 'linear-gradient(135deg,#4338ca,#6366f1)', color: '#fff' }
                    : { background: 'rgba(99,102,241,0.25)', color: '#a5b4fc' }
                }
              >
                {msg.role === 'assistant' ? 'AI' : 'U'}
              </div>
              <div
                className={cn(
                  'max-w-[75%] rounded-2xl px-4 py-3 text-sm leading-relaxed',
                  msg.role === 'user' ? 'rounded-tr-sm' : 'rounded-tl-sm'
                )}
                style={
                  msg.role === 'assistant'
                    ? { background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.07)', color: '#cbd5e1' }
                    : { background: 'linear-gradient(135deg,rgba(99,102,241,0.2),rgba(79,70,229,0.15))', border: '1px solid rgba(99,102,241,0.25)', color: '#e2e8f0' }
                }
              >
                <p className="whitespace-pre-wrap">{msg.content}</p>
                <p className="text-[10px] mt-1.5 opacity-40">
                  {msg.ts.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </p>
              </div>
            </div>
          ))}
          {isPending && (
            <div className="flex gap-3">
              <div className="w-8 h-8 rounded-full shrink-0 flex items-center justify-center text-xs font-bold" style={{ background: 'linear-gradient(135deg,#4338ca,#6366f1)', color: '#fff' }}>AI</div>
              <div className="rounded-2xl rounded-tl-sm px-4 py-3" style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.07)' }}>
                <Loader2 size={16} className="text-brand-400 animate-spin" />
              </div>
            </div>
          )}
          <div ref={bottomRef} />
        </div>

        {/* Input bar */}
        <div
          className="flex items-end gap-2 rounded-2xl px-4 py-3"
          style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.08)' }}
        >
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() } }}
            rows={1}
            placeholder="Ask anything about your platform, deployments, or logs…"
            className="flex-1 bg-transparent text-sm text-slate-200 placeholder-slate-600 resize-none focus:outline-none min-h-[24px] max-h-32"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || isPending}
            className={cn(
              'flex-shrink-0 w-9 h-9 rounded-xl flex items-center justify-center transition-all',
              input.trim() && !isPending
                ? 'bg-brand-500 hover:bg-brand-600 text-white shadow-lg shadow-brand-900/40'
                : 'text-slate-700 cursor-not-allowed'
            )}
            style={!input.trim() || isPending ? { background: 'rgba(255,255,255,0.05)' } : undefined}
          >
            {isPending ? <Loader2 size={15} className="animate-spin" /> : <Send size={15} />}
          </button>
        </div>
      </div>
    </div>
  )
}
