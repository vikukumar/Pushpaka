'use client'

import { usePathname } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { deploymentsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { StatusBadge } from '@/components/dashboard/StatusBadge'
import { timeAgo } from '@/lib/utils'
import { Deployment } from '@/types'
import { Loader2, GitBranch, GitCommit, ChevronRight } from 'lucide-react'

export default function ProjectDeploymentsPage() {
  const pathname = usePathname()
  const id = pathname.split('/')[3] || ''

  const { data, isLoading } = useQuery({
    queryKey: ['deployments', 'project', id],
    queryFn: () => deploymentsApi.list(50).then((r) => r.data),
    refetchInterval: 5000,
  })

  const deployments: Deployment[] = (data?.data || []).filter(
    (d: Deployment) => d.project_id === id
  )

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Deployments" subtitle="All deployments for this project" />

      <div className="p-6">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 size={24} className="animate-spin text-brand-400" />
          </div>
        ) : deployments.length === 0 ? (
          <div className="card text-center py-12">
            <p className="text-slate-400">No deployments yet for this project.</p>
          </div>
        ) : (
          <div className="card p-0 overflow-hidden">
            <div className="divide-y divide-surface-border">
              {deployments.map((d) => (
                <Link
                  key={d.id}
                  href={`/dashboard/deployments/${d.id}`}
                  className="flex items-center gap-4 px-5 py-4 hover:bg-slate-800/50 transition-colors"
                >
                  <StatusBadge status={d.status} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 text-sm text-slate-300">
                      <span className="flex items-center gap-1.5">
                        <GitBranch size={12} className="text-slate-500" />
                        {d.branch}
                      </span>
                      {d.commit_sha && (
                        <span className="flex items-center gap-1.5 text-slate-500">
                          <GitCommit size={12} />
                          {d.commit_sha.slice(0, 7)}
                        </span>
                      )}
                    </div>
                    {d.error_msg && (
                      <p className="text-xs text-red-400 mt-0.5 truncate">{d.error_msg}</p>
                    )}
                  </div>
                  <div className="text-xs text-slate-500 hidden md:block">{timeAgo(d.created_at)}</div>
                  <div className="text-xs text-slate-600 font-mono hidden md:block">{d.id.slice(0, 8)}</div>
                  <ChevronRight size={14} className="text-slate-600" />
                </Link>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

