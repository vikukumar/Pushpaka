'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Project, Deployment } from '@/types'
import { StatusBadge } from './StatusBadge'
import { timeAgo, truncate } from '@/lib/utils'
import { projectsApi } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { GitBranch, GitCommit, ExternalLink, Rocket, Trash2, Lock } from 'lucide-react'

interface ProjectCardProps {
  project: Project
  latestDeployment?: Deployment
  runningCount?: number
}

export function ProjectCard({ project, latestDeployment, runningCount = 0 }: ProjectCardProps) {
  const queryClient = useQueryClient()
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!confirm(`Delete project "${project.name}"? This cannot be undone.`)) return
    setDeleting(true)
    try {
      await projectsApi.delete(project.id)
      toast.success('Project deleted')
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    } catch {
      toast.error('Failed to delete project')
      setDeleting(false)
    }
  }

  return (
    <div className="card hover:border-brand-500/30 transition-all group animate-fade-in shadow-sm">
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1 min-w-0">
          <Link
            href={`/dashboard/projects/${project.id}`}
            className="text-base font-semibold text-[var(--text-primary)] hover:text-[var(--brand-primary)] transition-colors inline-flex items-center gap-1.5"
          >
            {project.is_private && <Lock size={12} className="text-slate-500 shrink-0" />}
            {project.name}
            {runningCount > 0 && (
              <span className="ml-2 px-1.5 py-0.5 text-[9px] font-bold bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 rounded-md animate-pulse-subtle">
                {runningCount} Running
              </span>
            )}
          </Link>
          <p className="text-xs text-slate-500 mt-0.5 truncate">{project.repo_url}</p>
        </div>
        <div className="flex items-center gap-2 ml-3">
          {latestDeployment && <StatusBadge status={latestDeployment.status} />}
          <button
            onClick={handleDelete}
            disabled={deleting}
            title="Delete project"
            className="p-1.5 rounded text-slate-600 hover:text-red-400 hover:bg-red-400/10 transition-colors disabled:opacity-50"
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>

      <div className="flex items-center gap-4 text-xs text-slate-500 mb-4">
        <span className="flex items-center gap-1.5">
          <GitBranch size={12} />
          {project.branch}
        </span>
        {latestDeployment?.commit_sha && (
          <span className="flex items-center gap-1.5">
            <GitCommit size={12} />
            {truncate(latestDeployment.commit_sha, 7)}
          </span>
        )}
        {latestDeployment && (
          <span className="ml-auto">{timeAgo(latestDeployment.created_at)}</span>
        )}
      </div>

      <div className="flex gap-2">
        <Link
          href={`/dashboard/projects/${project.id}`}
          className="btn-secondary text-xs py-1.5"
        >
          View
        </Link>
        {latestDeployment?.url && (
          <a
            href={latestDeployment.url}
            target="_blank"
            rel="noopener noreferrer"
            className="btn-secondary text-xs py-1.5"
          >
            <ExternalLink size={12} />
            Open
          </a>
        )}
        <Link
          href={`/dashboard/projects/${project.id}/deployments`}
          className="flex items-center gap-1.5 text-xs text-slate-500 hover:text-[var(--brand-primary)] transition-colors ml-auto"
        >
          <Rocket size={12} />
          {latestDeployment ? 'History' : 'No deployments'}
        </Link>
      </div>
    </div>
  )
}
