'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { auditApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { AuditLog } from '@/types'
import { Shield, Loader2, ChevronLeft, ChevronRight, Info } from 'lucide-react'
import { timeAgo } from '@/lib/utils'

const PAGE_SIZE = 50

export default function AuditPage() {
  const [offset, setOffset] = useState(0)

  const { data, isLoading } = useQuery({
    queryKey: ['audit', offset],
    queryFn: () => auditApi.list(PAGE_SIZE, offset).then((r) => r.data),
  })

  const logs: AuditLog[] = data?.data ?? []
  const hasMore = logs.length === PAGE_SIZE

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Audit Log" subtitle="Security and activity trail for your account" />

      <div className="p-6">
        <div className="card overflow-hidden p-0">
          {/* Table header */}
          <div
            className="grid grid-cols-[1fr_1fr_1fr_1fr_2fr] gap-3 px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider"
            style={{ borderBottom: '1px solid var(--surface-border)', background: 'rgba(255,255,255,0.02)' }}
          >
            <span>Time</span>
            <span>Action</span>
            <span>Resource</span>
            <span>IP</span>
            <span>User Agent</span>
          </div>

          {isLoading ? (
            <div className="flex justify-center items-center h-40">
              <Loader2 size={20} className="animate-spin text-brand-400" />
            </div>
          ) : logs.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-40 gap-2 text-slate-500">
              <Shield size={24} />
              <p className="text-sm">No audit entries yet</p>
            </div>
          ) : (
            <div className="divide-y divide-surface-border">
              {logs.map((log) => (
                <div
                  key={log.id}
                  className="grid grid-cols-[1fr_1fr_1fr_1fr_2fr] gap-3 px-4 py-3 text-sm hover:bg-white/[0.02] transition-colors"
                >
                  <span className="text-slate-400 text-xs whitespace-nowrap" title={log.created_at}>
                    {timeAgo(log.created_at)}
                  </span>
                  <span className="text-slate-200 font-mono text-xs">{log.action}</span>
                  <span className="text-slate-400 text-xs">
                    {log.resource}
                    {log.resource_id && (
                      <span className="ml-1 text-slate-600 font-mono">{log.resource_id.slice(0, 8)}</span>
                    )}
                  </span>
                  <span className="text-slate-500 text-xs font-mono">{log.ip_addr || '—'}</span>
                  <span
                    className="text-slate-600 text-xs truncate"
                    title={log.user_agent}
                  >
                    {log.user_agent || '—'}
                  </span>
                </div>
              ))}
            </div>
          )}

          {/* Pagination */}
          {!isLoading && (logs.length > 0 || offset > 0) && (
            <div
              className="flex items-center justify-between px-4 py-3 text-xs text-slate-500"
              style={{ borderTop: '1px solid var(--surface-border)' }}
            >
              <span className="flex items-center gap-1.5">
                <Info size={12} />
                Showing {offset + 1}–{offset + logs.length}
              </span>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                  disabled={offset === 0}
                  className="btn-secondary py-1 px-2 disabled:opacity-40"
                >
                  <ChevronLeft size={14} />
                </button>
                <button
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                  disabled={!hasMore}
                  className="btn-secondary py-1 px-2 disabled:opacity-40"
                >
                  <ChevronRight size={14} />
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
