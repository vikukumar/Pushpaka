'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import Link from 'next/link'
import { projectsApi, deploymentsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { StatusBadge } from '@/components/dashboard/StatusBadge'
import { timeAgo } from '@/lib/utils'
import toast from 'react-hot-toast'
import {
  Rocket, GitBranch, Globe, Key, Settings,
  ExternalLink, RefreshCw, Loader2, ChevronRight, GitCommit, Code2
} from 'lucide-react'

export default function ProjectDetailPage() {
  const pathname = usePathname()
  // useParams() returns the build-time placeholder '_' in the static export
  // because serverProvidedParams is baked into the pre-rendered RSC payload.
  // Read the real UUID from position 3 of the pathname instead.
  const id = pathname.split('/')[3] || ''
  const queryClient = useQueryClient()
  const [deployLoading, setDeployLoading] = useState(false)

  const { data: projectData, isLoading } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectsApi.get(id).then((r) => r.data),
  })

  const { data: deploymentsData } = useQuery({
    queryKey: ['deployments', 'project', id],
    queryFn: () => deploymentsApi.list(5, 0, id).then((r) => r.data),
    refetchInterval: 5000, // Live updates every 5s
  })

  const project = projectData
  const deployments = deploymentsData?.data || []

  const handleDeploy = async () => {
    setDeployLoading(true)
    try {
      await deploymentsApi.trigger({ project_id: id })
      toast.success('Deployment triggered!')
      queryClient.invalidateQueries({ queryKey: ['deployments'] })
    } catch {
      toast.error('Failed to trigger deployment')
    } finally {
      setDeployLoading(false)
    }
  }

  const handleSync = async () => {
    const loadingToast = toast.loading('Checking for changes...')
    try {
      const res = await projectsApi.sync(id)
      if (res.data?.code === 'UP_TO_DATE') {
        toast.dismiss(loadingToast)
        toast.success('Project is already up to date!')
      } else {
        toast.dismiss(loadingToast)
        toast.success('New changes detected! Deployment triggered.')
        queryClient.invalidateQueries({ queryKey: ['deployments'] })
      }
    } catch (err: any) {
      toast.dismiss(loadingToast)
      toast.error(err.response?.data?.error || 'Failed to sync with repository')
    }
  }

  if (isLoading) {
    return (
      <div className="flex flex-col min-h-screen">
        <Header title="Project" />
        <div className="p-6 flex items-center justify-center h-64">
          <Loader2 size={24} className="animate-spin text-brand-400" />
        </div>
      </div>
    )
  }

  if (!project) {
    return (
      <div className="flex flex-col min-h-screen">
        <Header title="Project not found" />
        <div className="p-6">
          <p className="text-slate-400">Project not found or you don&apos;t have access.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col min-h-screen animate-fade-in">
      <Header
        title={project.name}
        subtitle={project.repo_url}
      />

      <div className="p-6 space-y-6">
        {/* Actions row */}
        <div className="flex items-center gap-3 flex-wrap">
          <button
            onClick={handleDeploy}
            disabled={deployLoading}
            className="btn-primary"
          >
            {deployLoading ? (
              <Loader2 size={15} className="animate-spin" />
            ) : (
              <Rocket size={15} />
            )}
            Deploy Now
          </button>

          <button
            onClick={handleSync}
            className="btn-secondary"
          >
            <RefreshCw size={14} />
            Sync
          </button>

          <Link href={`/dashboard/projects/${id}/deployments`} className="btn-secondary">
            <RefreshCw size={14} />
            Deployments
          </Link>
          <Link href={`/dashboard/projects/${id}/env`} className="btn-secondary">
            <Key size={14} />
            Env Vars
          </Link>
          <Link href={`/dashboard/projects/${id}/domains`} className="btn-secondary">
            <Globe size={14} />
            Domains
          </Link>
          <Link href={`/dashboard/projects/${id}/editor`} className="btn-secondary">
            <Code2 size={14} />
            Editor
          </Link>
          <Link href={`/dashboard/projects/${id}/settings`} className="btn-secondary ml-auto">
            <Settings size={14} />
            Settings
          </Link>
        </div>

        {/* Project info */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="card">
            <h3 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4">
              Project Details
            </h3>
            <dl className="space-y-3">
              {[
                { label: 'Repository', value: project.repo_url, mono: true },
                { label: 'Branch', value: project.branch },
                { label: 'Latest Commit', value: project.latest_commit_sha ? `${project.latest_commit_sha.slice(0, 7)}: ${project.latest_commit_msg}` : 'Checking remote...', mono: true },
                { label: 'Last Synced', value: project.latest_commit_at ? timeAgo(project.latest_commit_at) : 'Never' },
                { label: 'Framework', value: project.framework || 'Auto-detect' },
                { label: 'Port', value: project.port?.toString() || '3000' },
                { label: 'Build Command', value: project.build_command || '', mono: true },
                { label: 'Start Command', value: project.start_command || '', mono: true },
              ].map(({ label, value, mono }) => (
                <div key={label} className="flex items-start gap-3">
                  <dt className="text-slate-500 text-sm w-36 shrink-0">{label}</dt>
                  <dd className={`text-sm text-slate-200 break-all ${mono ? 'font-mono' : ''}`}>{value}</dd>
                </div>
              ))}
            </dl>
          </div>

          {/* Latest deployments */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-slate-400 uppercase tracking-wider">
                Recent Deployments
              </h3>
              <Link
                href={`/dashboard/projects/${id}/deployments`}
                className="text-xs text-slate-500 hover:text-brand-400 flex items-center gap-1"
              >
                View all <ChevronRight size={12} />
              </Link>
            </div>

            {deployments.length === 0 ? (
              <div className="text-center py-8">
                <Rocket size={28} className="mx-auto text-slate-700 mb-2" />
                <p className="text-slate-500 text-sm">No deployments yet</p>
              </div>
            ) : (
              <div className="space-y-2">
                {deployments.map((d: {
                  id: string
                  status: 'queued' | 'building' | 'running' | 'failed' | 'stopped'
                  branch: string
                  commit_sha: string
                  url: string
                  created_at: string
                }) => (
                  <Link
                    key={d.id}
                    href={`/dashboard/deployments/${d.id}`}
                    className="flex items-center gap-3 p-2.5 rounded-lg hover:bg-slate-800 transition-colors"
                  >
                    <StatusBadge status={d.status} />
                    <div className="flex-1 min-w-0 text-xs">
                      <div className="text-slate-300 flex items-center gap-1.5">
                        <GitBranch size={10} /> {d.branch}
                        {d.commit_sha && (
                          <>
                            <GitCommit size={10} /> {d.commit_sha.slice(0, 7)}
                          </>
                        )}
                      </div>
                      <div className="text-slate-600">{timeAgo(d.created_at)}</div>
                    </div>
                    {d.url && (
                      <a
                        href={d.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-slate-600 hover:text-brand-400"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <ExternalLink size={12} />
                      </a>
                    )}
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}


// Required for Next.js static export with dynamic segments.
export function generateStaticParams() {
  return []
}
