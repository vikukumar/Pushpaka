'use client'

import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { deploymentsApi, logsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { StatusBadge } from '@/components/dashboard/StatusBadge'
import { LogViewer } from '@/components/dashboard/LogViewer'
import { Deployment, DeploymentLog } from '@/types'
import { timeAgo, formatDate } from '@/lib/utils'
import { ExternalLink, GitBranch, GitCommit, Clock, Loader2, RotateCcw } from 'lucide-react'
import toast from 'react-hot-toast'
import { useQueryClient } from '@tanstack/react-query'

export default function DeploymentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()

  const { data: deployment, isLoading: deployLoading } = useQuery<Deployment>({
    queryKey: ['deployment', id],
    queryFn: () => deploymentsApi.get(id).then((r) => r.data),
    refetchInterval: (d) =>
      d?.state.data?.status === 'building' || d?.state.data?.status === 'queued' ? 3000 : false,
  })

  const { data: logsData } = useQuery({
    queryKey: ['logs', id],
    queryFn: () => logsApi.get(id).then((r) => r.data),
    refetchInterval: (d) => {
      return deployment?.status === 'building' || deployment?.status === 'queued' ? 2000 : false
    },
  })

  const logs: DeploymentLog[] = logsData?.data || []
  const isLive = deployment?.status === 'building' || deployment?.status === 'queued'

  const handleRollback = async () => {
    try {
      await deploymentsApi.rollback(id)
      toast.success('Rollback deployment triggered!')
      queryClient.invalidateQueries({ queryKey: ['deployments'] })
    } catch {
      toast.error('Rollback failed')
    }
  }

  if (deployLoading) {
    return (
      <div className="flex flex-col min-h-screen">
        <Header title="Deployment" />
        <div className="flex justify-center items-center h-64">
          <Loader2 size={24} className="animate-spin text-brand-400" />
        </div>
      </div>
    )
  }

  if (!deployment) {
    return (
      <div className="flex flex-col min-h-screen">
        <Header title="Deployment not found" />
        <div className="p-6"><p className="text-slate-400">Deployment not found.</p></div>
      </div>
    )
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header
        title={`Deployment ${deployment.id.slice(0, 8)}`}
        subtitle={`Project ${deployment.project_id.slice(0, 8)}`}
      />

      <div className="p-6 space-y-5">
        {/* Header card */}
        <div className="card">
          <div className="flex items-start justify-between flex-wrap gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-3">
                <StatusBadge status={deployment.status} />
                <span className="text-xs text-slate-500 font-mono">{deployment.id}</span>
              </div>

              <div className="flex flex-wrap items-center gap-4 text-sm text-slate-400">
                <span className="flex items-center gap-1.5">
                  <GitBranch size={13} />
                  {deployment.branch}
                </span>
                {deployment.commit_sha && (
                  <span className="flex items-center gap-1.5 font-mono">
                    <GitCommit size={13} />
                    {deployment.commit_sha.slice(0, 7)}
                  </span>
                )}
                <span className="flex items-center gap-1.5">
                  <Clock size={13} />
                  {timeAgo(deployment.created_at)}
                </span>
              </div>

              {deployment.commit_msg && (
                <p className="text-sm text-slate-300 italic">&ldquo;{deployment.commit_msg}&rdquo;</p>
              )}

              {deployment.error_msg && (
                <p className="text-sm text-red-400 bg-red-500/10 rounded px-2 py-1">
                  {deployment.error_msg}
                </p>
              )}
            </div>

            <div className="flex items-center gap-2">
              {deployment.url && (
                <a
                  href={deployment.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn-secondary text-sm"
                >
                  <ExternalLink size={14} />
                  Open
                </a>
              )}
              <button className="btn-secondary text-sm" onClick={handleRollback}>
                <RotateCcw size={14} />
                Rollback
              </button>
            </div>
          </div>

          {/* Timeline */}
          <div className="mt-4 grid grid-cols-3 gap-4 pt-4 border-t border-surface-border">
            {[
              { label: 'Created', value: formatDate(deployment.created_at) },
              { label: 'Started',  value: formatDate(deployment.started_at) },
              { label: 'Finished', value: formatDate(deployment.finished_at) },
            ].map(({ label, value }) => (
              <div key={label}>
                <div className="text-xs text-slate-500 mb-0.5">{label}</div>
                <div className="text-xs text-slate-300">{value}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Logs */}
        <LogViewer
          logs={logs}
          isStreaming={isLive}
          deploymentId={id}
        />
      </div>
    </div>
  )
}
