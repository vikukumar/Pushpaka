'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { deploymentsApi, projectsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { StatusBadge } from '@/components/dashboard/StatusBadge'
import { Deployment, Project } from '@/types'
import { timeAgo } from '@/lib/utils'
import { GitBranch, Loader2, ChevronRight } from 'lucide-react'

export default function AllDeploymentsPage() {
  const { data: deploymentsData, isLoading } = useQuery({
    queryKey: ['deployments', 'all'],
    queryFn: () => deploymentsApi.list(50).then((r) => r.data),
    refetchInterval: 10000,
  })

  const { data: projectsData } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectsApi.list().then((r) => r.data),
  })

  const deployments: Deployment[] = deploymentsData?.data || []
  const projects: Project[] = projectsData?.data || []

  const getProject = (id: string) => projects.find((p) => p.id === id)

  return (
    <div className="flex flex-col min-h-screen">
      <Header
        title="Deployments"
        subtitle={`${deployments.length} deployment${deployments.length !== 1 ? 's' : ''}`}
      />

      <div className="p-6">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 size={24} className="animate-spin text-brand-400" />
          </div>
        ) : deployments.length === 0 ? (
          <div className="card text-center py-12">
            <p className="text-slate-400">No deployments yet. Create a project and trigger your first deploy.</p>
          </div>
        ) : (
          <div className="card p-0 overflow-hidden">
            <div className="divide-y divide-surface-border">
              {deployments.map((d) => {
                const project = getProject(d.project_id)
                return (
                  <Link
                    key={d.id}
                    href={`/dashboard/deployments/${d.id}`}
                    className="flex items-center gap-4 px-5 py-3.5 hover:bg-slate-800/50 transition-colors group"
                  >
                    <StatusBadge status={d.status} />

                    <div className="flex-1 min-w-0">
                      <div className="text-sm text-white font-medium truncate">
                        {project?.name || d.project_id.slice(0, 8)}
                      </div>
                      <div className="flex items-center gap-2 text-xs text-slate-500 mt-0.5">
                        <GitBranch size={10} />
                        {d.branch}
                        {d.error_msg && (
                          <span className="text-red-400 truncate">* {d.error_msg}</span>
                        )}
                      </div>
                    </div>

                    <div className="text-xs text-slate-500">{timeAgo(d.created_at)}</div>
                    <div className="text-xs text-slate-600 font-mono hidden md:block">
                      {d.id.slice(0, 8)}
                    </div>
                    <ChevronRight size={14} className="text-slate-600 group-hover:text-slate-400" />
                  </Link>
                )
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
