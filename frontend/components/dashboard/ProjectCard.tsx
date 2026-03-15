'use client'

import Link from 'next/link'
import { Project, Deployment } from '@/types'
import { StatusBadge } from './StatusBadge'
import { timeAgo, truncate } from '@/lib/utils'
import { GitBranch, GitCommit, ExternalLink, Rocket } from 'lucide-react'

interface ProjectCardProps {
  project: Project
  latestDeployment?: Deployment
}

export function ProjectCard({ project, latestDeployment }: ProjectCardProps) {
  return (
    <div className="card hover:border-brand-500/30 transition-all group">
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1 min-w-0">
          <Link
            href={`/dashboard/projects/${project.id}`}
            className="text-base font-semibold text-white hover:text-brand-300 transition-colors"
          >
            {project.name}
          </Link>
          <p className="text-xs text-slate-500 mt-0.5 truncate">{project.repo_url}</p>
        </div>
        <div className="flex items-center gap-2 ml-3">
          {latestDeployment && <StatusBadge status={latestDeployment.status} />}
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
          className="flex items-center gap-1.5 text-xs text-slate-500 hover:text-slate-300 transition-colors ml-auto"
        >
          <Rocket size={12} />
          {latestDeployment ? 'History' : 'No deployments'}
        </Link>
      </div>
    </div>
  )
}
