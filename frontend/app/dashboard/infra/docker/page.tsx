'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { containerApi } from '@/lib/api'
import { PageHeader } from '@/components/ui/PageHeader'
import {
  Container,
  Play,
  Square,
  RotateCcw,
  Terminal,
  RefreshCw,
  Loader2,
  CheckCircle2,
  XCircle,
  Clock,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface DockerContainer {
  id: string
  name: string
  image: string
  status: string
  state: 'running' | 'stopped' | 'exited' | 'paused' | 'restarting'
  ports: string
  created: string
}

function StatusDot({ state }: { state: string }) {
  if (state === 'running')
    return <CheckCircle2 size={13} className="text-emerald-400 shrink-0" />
  if (state === 'restarting')
    return <RotateCcw size={13} className="text-amber-400 animate-spin shrink-0" />
  return <XCircle size={13} className="text-red-400 shrink-0" />
}

export default function DockerPage() {
  const qc = useQueryClient()
  const [logs, setLogs] = useState<{ id: string; content: string } | null>(null)

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['containers'],
    queryFn: () => containerApi.list(),
    refetchInterval: 15_000,
  })

  const { mutate: start, isPending: starting } = useMutation({
    mutationFn: (id: string) => containerApi.start(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['containers'] }),
  })

  const { mutate: stop, isPending: stopping } = useMutation({
    mutationFn: (id: string) => containerApi.stop(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['containers'] }),
  })

  const { mutate: restart, isPending: restarting } = useMutation({
    mutationFn: (id: string) => containerApi.restart(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['containers'] }),
  })

  const { mutate: fetchLogs, isPending: loadingLogs } = useMutation({
    mutationFn: (id: string) => containerApi.logs(id, 200),
    onSuccess: (res, id) => setLogs({ id, content: res.data?.logs ?? '' }),
  })

  const containers: DockerContainer[] = data?.data ?? []
  const running = containers.filter((c) => c.state === 'running').length

  return (
    <div className="flex flex-col min-h-screen">
      <PageHeader
        title="Docker Containers"
        description="Manage and monitor running containers"
        icon={<Container className="text-sky-400" size={22} />}
        actions={
          <div className="flex items-center gap-2">
            <div
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
              style={{ background: 'rgba(52,211,153,0.1)', border: '1px solid rgba(52,211,153,0.25)', color: '#34d399' }}
            >
              <CheckCircle2 size={11} />
              {running} running
            </div>
            <button
              onClick={() => refetch()}
              disabled={isFetching}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 transition-colors"
              style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.08)' }}
            >
              <RefreshCw size={12} className={isFetching ? 'animate-spin' : ''} />
              Refresh
            </button>
          </div>
        }
      />

      <div className="flex-1 p-4 md:p-6 space-y-4">
        {isLoading ? (
          <div className="flex justify-center py-16">
            <Loader2 size={20} className="animate-spin text-brand-400" />
          </div>
        ) : containers.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Container size={40} className="text-slate-700 mb-3" />
            <p className="text-sm font-medium text-slate-400">No containers found</p>
            <p className="text-xs text-slate-600 mt-1">Docker daemon may not be running or no containers exist.</p>
          </div>
        ) : (
          <div
            className="rounded-xl overflow-hidden"
            style={{ border: '1px solid rgba(255,255,255,0.07)' }}
          >
            <table className="w-full text-sm">
              <thead>
                <tr style={{ background: 'rgba(255,255,255,0.03)', borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
                  {['Status', 'Name', 'Image', 'Ports', 'Created', 'Actions'].map((h) => (
                    <th key={h} className="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y" style={{ borderColor: 'rgba(255,255,255,0.04)' }}>
                {containers.map((c) => (
                  <tr key={c.id} className="hover:bg-white/[0.015] transition-colors">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1.5">
                        <StatusDot state={c.state} />
                        <span className={cn(
                          'text-xs capitalize',
                          c.state === 'running' ? 'text-emerald-400' : 'text-slate-500'
                        )}>
                          {c.state}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-slate-300">{c.name}</td>
                    <td className="px-4 py-3 font-mono text-xs text-slate-500 max-w-[180px] truncate">{c.image}</td>
                    <td className="px-4 py-3 font-mono text-xs text-slate-600">{c.ports || '—'}</td>
                    <td className="px-4 py-3 text-xs text-slate-600">
                      <div className="flex items-center gap-1">
                        <Clock size={11} />
                        {c.created}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1">
                        {c.state !== 'running' ? (
                          <button
                            onClick={() => start(c.id)}
                            disabled={starting}
                            title="Start"
                            className="p-1.5 rounded-lg text-emerald-600 hover:text-emerald-400 hover:bg-emerald-500/10 transition-colors"
                          >
                            {starting ? <Loader2 size={13} className="animate-spin" /> : <Play size={13} />}
                          </button>
                        ) : (
                          <button
                            onClick={() => stop(c.id)}
                            disabled={stopping}
                            title="Stop"
                            className="p-1.5 rounded-lg text-red-600 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                          >
                            {stopping ? <Loader2 size={13} className="animate-spin" /> : <Square size={13} />}
                          </button>
                        )}
                        <button
                          onClick={() => restart(c.id)}
                          disabled={restarting}
                          title="Restart"
                          className="p-1.5 rounded-lg text-amber-600 hover:text-amber-400 hover:bg-amber-500/10 transition-colors"
                        >
                          {restarting ? <Loader2 size={13} className="animate-spin" /> : <RotateCcw size={13} />}
                        </button>
                        <button
                          onClick={() => fetchLogs(c.id)}
                          disabled={loadingLogs}
                          title="View logs"
                          className="p-1.5 rounded-lg text-sky-600 hover:text-sky-400 hover:bg-sky-500/10 transition-colors"
                        >
                          {loadingLogs ? <Loader2 size={13} className="animate-spin" /> : <Terminal size={13} />}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Log viewer */}
        {logs && (
          <div
            className="rounded-xl overflow-hidden"
            style={{ border: '1px solid rgba(255,255,255,0.07)' }}
          >
            <div
              className="flex items-center justify-between px-4 py-3"
              style={{ background: 'rgba(255,255,255,0.03)', borderBottom: '1px solid rgba(255,255,255,0.06)' }}
            >
              <p className="text-xs font-semibold text-slate-400 font-mono">
                Logs — {logs.id.slice(0, 12)}
              </p>
              <button
                onClick={() => setLogs(null)}
                className="text-xs text-slate-600 hover:text-slate-400 transition-colors"
              >
                Close ✕
              </button>
            </div>
            <pre
              className="overflow-x-auto p-4 text-xs text-slate-400 font-mono leading-relaxed max-h-80 overflow-y-auto"
              style={{ background: 'rgba(0,0,0,0.4)' }}
            >
              {logs.content || 'No log output.'}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
