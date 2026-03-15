'use client'

import { useEffect, useRef, useState } from 'react'
import { DeploymentLog } from '@/types'
import { cn } from '@/lib/utils'
import { Terminal, Download, XCircle } from 'lucide-react'

interface LogViewerProps {
  logs: DeploymentLog[]
  isStreaming?: boolean
  deploymentId?: string
}

const levelColors: Record<string, string> = {
  info:  'text-slate-300',
  error: 'text-red-400',
  warn:  'text-amber-400',
  debug: 'text-slate-500',
}

const streamColors: Record<string, string> = {
  system: 'text-brand-400',
  stdout: 'text-slate-300',
  stderr: 'text-red-400',
}

export function LogViewer({ logs, isStreaming, deploymentId }: LogViewerProps) {
  const bottomRef = useRef<HTMLDivElement>(null)
  const [autoScroll, setAutoScroll] = useState(true)
  const wsRef = useRef<WebSocket | null>(null)
  const [streamLogs, setStreamLogs] = useState<DeploymentLog[]>(logs)

  useEffect(() => {
    setStreamLogs(logs)
  }, [logs])

  // WebSocket streaming
  useEffect(() => {
    if (!isStreaming || !deploymentId) return

    const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'
    const token = typeof window !== 'undefined' ? localStorage.getItem('pushpaka_token') : ''
    const ws = new WebSocket(`${WS_URL}/api/v1/logs/${deploymentId}/stream?token=${token}`)
    wsRef.current = ws

    ws.onmessage = (event) => {
      try {
        const log: DeploymentLog = JSON.parse(event.data)
        setStreamLogs((prev) => [...prev, log])
      } catch {/* ignore parse errors */}
    }

    return () => {
      ws.close()
    }
  }, [isStreaming, deploymentId])

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [streamLogs, autoScroll])

  const handleDownload = () => {
    const text = streamLogs.map((l) => `[${l.created_at}] [${l.level.toUpperCase()}] ${l.message}`).join('\n')
    const blob = new Blob([text], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `deployment-${deploymentId?.slice(0, 8)}-logs.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="bg-[#0a0e1a] rounded-xl border border-surface-border overflow-hidden">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2.5 border-b border-surface-border bg-surface-elevated">
        <div className="flex items-center gap-2 text-sm text-slate-400">
          <Terminal size={14} />
          <span>Deployment Logs</span>
          {isStreaming && (
            <span className="flex items-center gap-1.5 text-xs text-emerald-400">
              <span className="w-1.5 h-1.5 bg-emerald-400 rounded-full animate-pulse" />
              Live
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setAutoScroll(!autoScroll)}
            className={cn(
              'text-xs px-2 py-1 rounded transition-colors',
              autoScroll ? 'text-brand-400 bg-brand-500/10' : 'text-slate-500 hover:text-slate-300'
            )}
          >
            Auto-scroll {autoScroll ? 'on' : 'off'}
          </button>
          <button
            onClick={handleDownload}
            className="p-1.5 text-slate-500 hover:text-slate-300 transition-colors"
          >
            <Download size={13} />
          </button>
        </div>
      </div>

      {/* Log content */}
      <div className="h-96 overflow-y-auto p-4 font-mono text-xs space-y-0.5">
        {streamLogs.length === 0 ? (
          <div className="text-slate-600 flex items-center gap-2">
            <XCircle size={14} />
            No logs available
          </div>
        ) : (
          streamLogs.map((log) => (
            <div key={log.id} className="flex gap-3 hover:bg-white/3 rounded px-1">
              <span className="text-slate-600 select-none shrink-0">
                {new Date(log.created_at).toLocaleTimeString()}
              </span>
              <span className={cn('shrink-0', streamColors[log.stream] ?? 'text-slate-400')}>
                [{log.stream}]
              </span>
              <span className={cn(levelColors[log.level] ?? 'text-slate-300', 'break-all')}>
                {log.message}
              </span>
            </div>
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
