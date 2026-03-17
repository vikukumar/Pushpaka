'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { alertsApi } from '@/lib/api'
import { PageHeader } from '@/components/ui/PageHeader'
import {
  AlertTriangle,
  CheckCircle2,
  RefreshCw,
  Info,
  Loader2,
  Filter,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface Alert {
  id: string
  message: string
  severity: 'critical' | 'warning' | 'info'
  resolved: boolean
  created_at: string
  resolved_at?: string
}

function severityBadge(s: string) {
  if (s === 'critical')
    return { bg: 'rgba(239,68,68,0.12)', border: 'rgba(239,68,68,0.3)', text: '#f87171', label: 'Critical' }
  if (s === 'warning')
    return { bg: 'rgba(245,158,11,0.12)', border: 'rgba(245,158,11,0.3)', text: '#fbbf24', label: 'Warning' }
  return { bg: 'rgba(99,102,241,0.12)', border: 'rgba(99,102,241,0.3)', text: '#818cf8', label: 'Info' }
}

export default function AIMonitoringPage() {
  const qc = useQueryClient()
  const [onlyUnresolved, setOnlyUnresolved] = useState(false)

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['ai-alerts', onlyUnresolved],
    queryFn: () => alertsApi.list(onlyUnresolved),
    refetchInterval: 30_000,
  })

  const { mutate: resolve, isPending: resolving } = useMutation({
    mutationFn: (id: string) => alertsApi.resolve(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ai-alerts'] }),
  })

  const alerts: Alert[] = data?.data ?? []
  const unresolvedCount = alerts.filter((a) => !a.resolved).length

  return (
    <div className="flex flex-col min-h-screen">
      <PageHeader
        title="AI Monitoring"
        description="Real-time platform alerts and anomaly detection"
        icon={<AlertTriangle className="text-amber-400" size={22} />}
        actions={
          <div className="flex items-center gap-2">
            <button
              onClick={() => setOnlyUnresolved((v) => !v)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors',
                onlyUnresolved ? 'text-amber-300' : 'text-slate-400 hover:text-slate-200'
              )}
              style={{
                background: onlyUnresolved ? 'rgba(245,158,11,0.12)' : 'rgba(255,255,255,0.05)',
                border: `1px solid ${onlyUnresolved ? 'rgba(245,158,11,0.3)' : 'rgba(255,255,255,0.08)'}`,
              }}
            >
              <Filter size={12} />
              {onlyUnresolved ? 'Unresolved only' : 'All alerts'}
            </button>
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
        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[
            { label: 'Total Alerts', value: alerts.length, color: '#818cf8' },
            { label: 'Unresolved', value: unresolvedCount, color: '#f87171' },
            { label: 'Resolved', value: alerts.length - unresolvedCount, color: '#34d399' },
            {
              label: 'Critical',
              value: alerts.filter((a) => a.severity === 'critical' && !a.resolved).length,
              color: '#fb923c',
            },
          ].map((stat) => (
            <div
              key={stat.label}
              className="rounded-xl p-4"
              style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
            >
              <p className="text-xs text-slate-600 mb-1">{stat.label}</p>
              <p className="text-2xl font-bold" style={{ color: stat.color }}>
                {isLoading ? '—' : stat.value}
              </p>
            </div>
          ))}
        </div>

        {/* Alert list */}
        <div
          className="rounded-xl overflow-hidden"
          style={{ border: '1px solid rgba(255,255,255,0.07)' }}
        >
          <div
            className="px-4 py-3 flex items-center justify-between"
            style={{ background: 'rgba(255,255,255,0.03)', borderBottom: '1px solid rgba(255,255,255,0.06)' }}
          >
            <p className="text-sm font-semibold text-slate-300">Alert Feed</p>
            {unresolvedCount > 0 && (
              <span
                className="text-xs font-semibold px-2 py-0.5 rounded-full"
                style={{ background: 'rgba(239,68,68,0.15)', color: '#f87171', border: '1px solid rgba(239,68,68,0.3)' }}
              >
                {unresolvedCount} active
              </span>
            )}
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 size={20} className="animate-spin text-brand-400" />
            </div>
          ) : alerts.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <CheckCircle2 size={36} className="text-emerald-500 mb-3" />
              <p className="text-sm font-medium text-slate-400">All clear!</p>
              <p className="text-xs text-slate-600 mt-1">No alerts to display. The platform is running smoothly.</p>
            </div>
          ) : (
            <div className="divide-y" style={{ borderColor: 'rgba(255,255,255,0.05)' }}>
              {alerts.map((alert) => {
                const badge = severityBadge(alert.severity)
                return (
                  <div
                    key={alert.id}
                    className={cn('flex items-start gap-3 px-4 py-3.5 transition-colors')}
                    style={{ opacity: alert.resolved ? 0.5 : 1 }}
                  >
                    <Info size={15} className="mt-0.5 shrink-0" style={{ color: badge.text }} />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-slate-300 leading-snug">{alert.message}</p>
                      <p className="text-[11px] text-slate-600 mt-0.5">
                        {new Date(alert.created_at).toLocaleString()}
                        {alert.resolved && alert.resolved_at && (
                          <span className="ml-2 text-emerald-600">
                            Resolved {new Date(alert.resolved_at).toLocaleString()}
                          </span>
                        )}
                      </p>
                    </div>
                    <span
                      className="shrink-0 text-[10px] font-semibold px-2 py-0.5 rounded-full"
                      style={{ background: badge.bg, border: `1px solid ${badge.border}`, color: badge.text }}
                    >
                      {badge.label}
                    </span>
                    {!alert.resolved && (
                      <button
                        onClick={() => resolve(alert.id)}
                        disabled={resolving}
                        className="shrink-0 text-xs font-medium px-2.5 py-1 rounded-lg transition-colors"
                        style={{
                          background: 'rgba(52,211,153,0.1)',
                          border: '1px solid rgba(52,211,153,0.25)',
                          color: '#34d399',
                        }}
                      >
                        {resolving ? <Loader2 size={11} className="animate-spin" /> : 'Resolve'}
                      </button>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
