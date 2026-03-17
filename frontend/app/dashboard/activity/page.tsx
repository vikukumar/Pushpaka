'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { auditApi } from '@/lib/api'
import { AuditLog } from '@/types'
import { Header } from '@/components/layout/Header'
import { timeAgo } from '@/lib/utils'
import Link from 'next/link'
import {
  Activity, Plus, Pencil, Trash2, Rocket, RotateCcw,
  Globe, Key, Bell, Settings, ChevronRight, Loader2,
  RefreshCw, Filter
} from 'lucide-react'

// ── Icons & colours keyed on action ──────────────────────────────────────────
type ActionMeta = { icon: React.ReactNode; color: string; label: string }

function getActionMeta(action: string, resource: string): ActionMeta {
  const a = action.toLowerCase()
  const r = resource.toLowerCase()
  if (a === 'create' && r === 'project')
    return { icon: <Plus size={14} />, color: 'text-green-400 bg-green-400/10', label: 'Project created' }
  if (a === 'update' && r === 'project')
    return { icon: <Pencil size={14} />, color: 'text-blue-400 bg-blue-400/10', label: 'Project updated' }
  if (a === 'delete' && r === 'project')
    return { icon: <Trash2 size={14} />, color: 'text-red-400 bg-red-400/10', label: 'Project deleted' }
  if (a === 'deploy')
    return { icon: <Rocket size={14} />, color: 'text-brand-400 bg-brand-400/10', label: 'Deployment triggered' }
  if (a === 'rollback')
    return { icon: <RotateCcw size={14} />, color: 'text-amber-400 bg-amber-400/10', label: 'Rollback triggered' }
  if (a === 'delete' && r === 'deployment')
    return { icon: <Trash2 size={14} />, color: 'text-red-400 bg-red-400/10', label: 'Deployment deleted' }
  if (r === 'domain')
    return { icon: <Globe size={14} />, color: 'text-cyan-400 bg-cyan-400/10', label: `Domain ${a}d` }
  if (r === 'env_var' || r === 'env')
    return { icon: <Key size={14} />, color: 'text-yellow-400 bg-yellow-400/10', label: `Env var ${a}d` }
  if (r === 'notification')
    return { icon: <Bell size={14} />, color: 'text-purple-400 bg-purple-400/10', label: 'Notifications updated' }
  return { icon: <Settings size={14} />, color: 'text-slate-400 bg-slate-400/10', label: `${resource} ${action}` }
}

// ── Resource link resolver ──────────────────────────────────────────────────
function resourceLink(log: AuditLog): string | null {
  const r = log.resource.toLowerCase()
  if (r === 'project' && log.resource_id) return `/dashboard/projects/${log.resource_id}`
  if (r === 'deployment' && log.resource_id) return `/dashboard/deployments/${log.resource_id}`
  return null
}

// ── Metadata renderer ────────────────────────────────────────────────────────
function parseMeta(raw: string): Record<string, string> {
  try { return JSON.parse(raw) ?? {} } catch { return {} }
}

function MetaBadge({ label, value }: { label: string; value: string }) {
  return (
    <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full"
      style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.09)' }}>
      <span className="text-slate-500">{label}:</span>
      <span className="text-slate-300 font-mono truncate max-w-[180px]">{value}</span>
    </span>
  )
}

// ── Grouping helper ──────────────────────────────────────────────────────────
function groupLabel(dateStr: string): string {
  const d = new Date(dateStr)
  const now = new Date()
  const diffDays = Math.floor((now.getTime() - d.getTime()) / 86400000)
  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays < 7) return 'This week'
  if (diffDays < 30) return 'This month'
  return 'Older'
}

const GROUP_ORDER = ['Today', 'Yesterday', 'This week', 'This month', 'Older']

type FilterType = 'all' | 'project' | 'deployment' | 'env' | 'domain'

export default function ActivityPage() {
  const [limit, setLimit] = useState(50)
  const [filter, setFilter] = useState<FilterType>('all')

  const { data, isLoading, isFetching, refetch } = useQuery<{ data: AuditLog[] } | null>({
    queryKey: ['activity', limit],
    queryFn: () => auditApi.list(limit, 0).then((r) => r.data ?? null),
    refetchInterval: 30000,
  })

  const logs = ((data?.data) ?? []).filter((l) => {
    if (filter === 'all') return true
    if (filter === 'project') return l.resource.toLowerCase() === 'project'
    if (filter === 'deployment') return l.resource.toLowerCase() === 'deployment'
    if (filter === 'env') return l.resource.toLowerCase().includes('env')
    if (filter === 'domain') return l.resource.toLowerCase() === 'domain'
    return true
  })

  // Group logs by time period
  const groups: Record<string, AuditLog[]> = {}
  for (const log of logs) {
    const label = groupLabel(log.created_at)
    if (!groups[label]) groups[label] = []
    groups[label].push(log)
  }

  const filterButtons: { key: FilterType; label: string }[] = [
    { key: 'all', label: 'All' },
    { key: 'project', label: 'Projects' },
    { key: 'deployment', label: 'Deployments' },
    { key: 'env', label: 'Env Vars' },
    { key: 'domain', label: 'Domains' },
  ]

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Activity" subtitle="Complete audit trail of all platform events" />

      <div className="p-6 max-w-3xl">
        {/* Toolbar */}
        <div className="flex items-center gap-3 mb-6 flex-wrap">
          {/* Filter chips */}
          <div className="flex items-center gap-1.5 p-1 rounded-lg flex-wrap"
            style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.07)' }}>
            <Filter size={12} className="text-slate-500 ml-1.5" />
            {filterButtons.map(({ key, label }) => (
              <button
                key={key}
                onClick={() => setFilter(key)}
                className={`px-3 py-1 rounded-md text-xs font-medium transition-all ${
                  filter === key ? 'text-white' : 'text-slate-500 hover:text-slate-300'
                }`}
                style={filter === key ? { background: 'rgba(99,102,241,0.25)', border: '1px solid rgba(99,102,241,0.35)' } : {}}
              >
                {label}
              </button>
            ))}
          </div>

          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="btn-secondary text-xs py-1.5 ml-auto flex items-center gap-1.5"
          >
            <RefreshCw size={12} className={isFetching ? 'animate-spin' : ''} />
            Refresh
          </button>
        </div>

        {/* Feed */}
        {isLoading ? (
          <div className="flex justify-center py-16">
            <Loader2 size={24} className="animate-spin text-brand-400" />
          </div>
        ) : logs.length === 0 ? (
          <div className="card text-center py-16">
            <Activity size={40} className="mx-auto text-slate-700 mb-4" />
            <p className="text-slate-400 text-sm">No activity events yet.</p>
            <p className="text-slate-600 text-xs mt-1">Events are recorded when you create projects, trigger deployments, and more.</p>
          </div>
        ) : (
          <div className="space-y-8">
            {GROUP_ORDER.filter(g => groups[g]?.length).map((groupName) => (
              <div key={groupName}>
                {/* Group label */}
                <div className="flex items-center gap-3 mb-3">
                  <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">{groupName}</span>
                  <div className="flex-1 h-px bg-surface-border" />
                </div>

                {/* Events */}
                <div className="relative">
                  {/* Vertical timeline line */}
                  <div className="absolute left-5 top-0 bottom-0 w-px bg-surface-border" />

                  <div className="space-y-0.5">
                    {groups[groupName].map((log) => {
                      const meta = getActionMeta(log.action, log.resource)
                      const extra = parseMeta(log.metadata)
                      const link = resourceLink(log)

                      return (
                        <div key={log.id} className="flex gap-4 items-start group">
                          {/* Icon bubble */}
                          <div className={`relative z-10 flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${meta.color}`}>
                            {meta.icon}
                          </div>

                          {/* Content */}
                          <div className="flex-1 min-w-0 py-2">
                            <div className="flex items-center gap-2 flex-wrap">
                              <span className="text-sm font-medium text-white">{meta.label}</span>
                              {link && (
                                <Link
                                  href={link}
                                  className="text-xs text-brand-400 hover:text-brand-300 flex items-center gap-0.5"
                                >
                                  View <ChevronRight size={11} />
                                </Link>
                              )}
                              <span className="text-xs text-slate-600 ml-auto">{timeAgo(log.created_at)}</span>
                            </div>

                            {/* Metadata badges */}
                            {Object.keys(extra).length > 0 && (
                              <div className="flex flex-wrap gap-1.5 mt-1.5">
                                {Object.entries(extra).map(([k, v]) => (
                                  <MetaBadge key={k} label={k} value={String(v)} />
                                ))}
                              </div>
                            )}

                            {/* Resource ID (small) */}
                            <div className="text-xs text-slate-700 mt-0.5 font-mono">
                              {log.resource}/{log.resource_id?.slice(0, 8)}
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              </div>
            ))}

            {/* Load more */}
            {logs.length >= limit && (
              <div className="text-center pt-2">
                <button
                  onClick={() => setLimit((l) => l + 50)}
                  className="btn-secondary text-xs"
                >
                  Load more
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
