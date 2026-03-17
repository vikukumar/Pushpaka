'use client'

import { useQuery } from '@tanstack/react-query'
import { systemApi } from '@/lib/api'
import { SystemInfo } from '@/types'
import { Container, GitBranch, Server, Cpu, RefreshCw, Zap } from 'lucide-react'

function StatusDot({ ok, pulse = false }: { ok: boolean; pulse?: boolean }) {
  const color = ok ? '#4ade80' : '#f87171'
  const glow  = ok ? 'rgba(74,222,128,0.6)' : 'rgba(248,113,113,0.6)'
  return (
    <span
      className={pulse && ok ? 'animate-glow-pulse' : ''}
      style={{
        display: 'inline-block',
        width: 8, height: 8,
        borderRadius: '50%',
        background: color,
        boxShadow: `0 0 8px ${glow}`,
        flexShrink: 0,
      }}
    />
  )
}

function Row({
  icon: Icon, iconColor, label, detail, ok, extra,
}: {
  icon: React.ElementType
  iconColor: string
  label: string
  detail?: string
  ok: boolean
  extra?: React.ReactNode
}) {
  return (
    <div
      className="flex items-center gap-3 p-3 rounded-xl"
      style={{
        background: 'rgba(255,255,255,0.025)',
        border: '1px solid rgba(99,102,241,0.09)',
        boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.03)',
      }}
    >
      {/* Icon halo */}
      <div
        className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0"
        style={{
          background: `${iconColor}12`,
          border: `1px solid ${iconColor}30`,
          boxShadow: `0 0 12px -4px ${iconColor}40`,
        }}
      >
        <Icon size={15} style={{ color: iconColor }} />
      </div>

      {/* Text */}
      <div className="flex-1 min-w-0">
        <p className="text-xs font-semibold text-slate-200 truncate">{label}</p>
        {detail && (
          <p className="text-[10px] text-slate-600 mt-0.5 truncate">{detail}</p>
        )}
      </div>

      {/* Status + extra */}
      <div className="flex items-center gap-2 shrink-0">
        {extra}
        <StatusDot ok={ok} pulse={ok} />
      </div>
    </div>
  )
}

export function SystemStatus() {
  const { data, isLoading, isError, refetch, isFetching } = useQuery<SystemInfo>({
    queryKey: ['system'],
    queryFn: () => systemApi.get().then((r) => r.data),
    refetchInterval: 10_000,
  })

  if (isLoading) {
    return (
      <div className="card space-y-3 animate-pulse">
        <div className="h-4 rounded" style={{ background: 'rgba(99,102,241,0.15)', width: '40%' }} />
        {[...Array(4)].map((_, i) => (
          <div key={i} className="h-14 rounded-xl" style={{ background: 'rgba(255,255,255,0.03)' }} />
        ))}
      </div>
    )
  }

  if (isError || !data) {
    return (
      <div className="card">
        <p className="text-sm text-red-400">Failed to load system status</p>
      </div>
    )
  }

  const { docker, git, workers, runtime } = data

  return (
    <div className="card">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <h2
          className="text-sm font-semibold flex items-center gap-2"
          style={{
            background: 'linear-gradient(90deg, #a5b4fc, #67e8f9)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            backgroundClip: 'text',
          }}
        >
          <Server size={14} style={{ color: '#818cf8' }} />
          System Status
        </h2>
        <button
          onClick={() => refetch()}
          className="p-1.5 rounded-lg transition-colors text-slate-600 hover:text-slate-300"
          style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(99,102,241,0.1)' }}
          title="Refresh"
        >
          <RefreshCw size={12} className={isFetching ? 'animate-spin-slow' : ''} />
        </button>
      </div>

      <div className="space-y-2.5">
        {/* Docker */}
        <Row
          icon={Container}
          iconColor={docker.available ? '#4ade80' : '#f87171'}
          label="Docker"
          detail={docker.available ? docker.host || 'Connected' : 'Not found " direct deploy mode'}
          ok={docker.available}
        />

        {/* Git */}
        <Row
          icon={GitBranch}
          iconColor={git.available ? '#34d399' : '#f87171'}
          label="Git"
          detail={git.version || (git.available ? 'Available' : 'Not found')}
          ok={git.available}
        />

        {/* Workers */}
        <Row
          icon={Cpu}
          iconColor="#818cf8"
          label="Build Workers"
          detail={
            !workers.tracked
              ? 'External workers via Redis (untracked)'
              : `${workers.total} workers \u00b7 ${workers.active_jobs} active \u00b7 ${workers.idle} idle`
          }
          ok={!workers.tracked ? true : workers.total > 0}
          extra={
            <span
              className="text-[10px] font-mono px-1.5 py-0.5 rounded"
              style={{
                background: 'rgba(99,102,241,0.12)',
                color: '#818cf8',
                border: '1px solid rgba(99,102,241,0.2)',
              }}
            >
              {workers.queue_mode}
            </span>
          }
        />

        {/* Active jobs bar  only shown for in-process (tracked) workers */}
        {workers.tracked && workers.total > 0 && (
          <div
            className="px-3 py-2.5 rounded-xl"
            style={{
              background: 'rgba(255,255,255,0.015)',
              border: '1px solid rgba(99,102,241,0.07)',
            }}
          >
            <div className="flex items-center justify-between mb-2">
              <span className="text-[10px] text-slate-600 uppercase tracking-widest font-semibold">
                Worker Load
              </span>
              <span className="text-[10px] text-slate-500">
                {workers.active_jobs}/{workers.total} busy
              </span>
            </div>
            <div className="h-1.5 rounded-full" style={{ background: 'rgba(99,102,241,0.12)' }}>
              <div
                className="h-1.5 rounded-full transition-all duration-700"
                style={{
                  width: workers.total > 0 ? `${(workers.active_jobs / workers.total) * 100}%` : '0%',
                  background: 'linear-gradient(90deg, #6366f1, #22d3ee)',
                  boxShadow: '0 0 8px rgba(99,102,241,0.5)',
                  minWidth: workers.active_jobs > 0 ? '8px' : '0',
                }}
              />
            </div>
          </div>
        )}

        {/* Runtime */}
        <Row
          icon={Zap}
          iconColor="#22d3ee"
          label="Runtime"
          detail={`${runtime.os} / ${runtime.arch}${runtime.in_container ? ' * container' : ''}`}
          ok
        />
      </div>
    </div>
  )
}


