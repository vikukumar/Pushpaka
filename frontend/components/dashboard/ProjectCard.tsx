'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Project, Deployment } from '@/types'
import { StatusBadge } from './StatusBadge'
import { timeAgo, truncate } from '@/lib/utils'
import { projectsApi } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { 
  GitBranch, GitCommit, ExternalLink, Rocket, Trash2, Lock, 
  Code2, Coffee, Terminal, Globe, Cpu, Server, Layers
} from 'lucide-react'
import { useConfirm } from '@/components/ui/Modal'

interface ProjectCardProps {
  project: Project
  latestDeployment?: Deployment
  runningCount?: number
  buildingCount?: number
  failedCount?: number
}

const getFrameworkIcon = (framework?: string) => {
  const fw = framework?.toLowerCase() || ''
  if (fw.includes('react') || fw.includes('next')) return <Layers size={14} className="text-blue-400" />
  if (fw.includes('vue')) return <Layers size={14} className="text-emerald-400" />
  if (fw.includes('node')) return <Server size={14} className="text-green-500" />
  if (fw.includes('python')) return <Terminal size={14} className="text-blue-500" />
  if (fw.includes('go')) return <Cpu size={14} className="text-cyan-500" />
  if (fw.includes('static') || fw.includes('html')) return <Globe size={14} className="text-slate-400" />
  return <Code2 size={14} className="text-slate-500" />
}

export function ProjectCard({ 
  project, 
  latestDeployment, 
  runningCount = 0,
  buildingCount = 0,
  failedCount = 0
}: ProjectCardProps) {
  const queryClient = useQueryClient()
  const [deleting, setDeleting] = useState(false)
  const { confirm, Component: ConfirmModal } = useConfirm()

  const handleDelete = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    
    const ok = await confirm({
      title: 'Delete Project',
      message: `Are you sure you want to delete "${project.name}"? This will permanently remove the project and all its deployments.`,
      confirmText: 'Delete Project',
      type: 'error'
    })
    
    if (!ok) return
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

  const isBuilding = latestDeployment?.status === 'building' || buildingCount > 0

  return (
    <>
    <div className={`card overflow-hidden relative group animate-fade-in shadow-sm transition-all duration-300 hover:translate-y-[-2px] hover:shadow-lg border-l-4 ${
      latestDeployment?.status === 'running' ? 'border-l-emerald-500' : 
      latestDeployment?.status === 'failed' ? 'border-l-red-500' : 
      isBuilding ? 'border-l-amber-500' : 'border-l-slate-700'
    }`}>
      {/* Shine effect overlay */}
      <div className="absolute inset-0 pointer-events-none z-0 overflow-hidden">
        <div className={`absolute -inset-[100%] aspect-square bg-[conic-gradient(from_0deg,transparent_0deg,rgba(99,102,241,0.03)_120deg,transparent_180deg)] animate-[spin_8s_linear_infinite] ${!isBuilding && 'opacity-0 group-hover:opacity-100 transition-opacity duration-700'}`} />
        <div className="absolute inset-0 bg-gradient-to-br from-transparent via-white/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>

      <div className="relative z-10">
        <div className="flex items-start justify-between mb-3">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              {getFrameworkIcon(project.framework)}
              <span className="text-[10px] font-medium text-slate-500 uppercase tracking-tighter">
                {project.framework || 'Detecting...'}
              </span>
            </div>
            <Link
              href={`/dashboard/projects/${project.id}`}
              className="text-base font-semibold text-[var(--text-primary)] hover:text-[var(--brand-primary)] transition-colors inline-flex items-center gap-1.5"
            >
              {project.is_private && <Lock size={12} className="text-slate-500 shrink-0" />}
              {project.name}
            </Link>
          </div>
          <div className="flex items-center gap-2 ml-3">
            {latestDeployment && <StatusBadge status={latestDeployment.status} />}
            <button
              onClick={handleDelete}
              disabled={deleting}
              className="p-1.5 rounded text-slate-600 hover:text-red-400 hover:bg-red-400/10 transition-colors disabled:opacity-50"
            >
              <Trash2 size={14} />
            </button>
          </div>
        </div>

        {/* Commit message - NEW */}
        {project.latest_commit_msg && (
          <div className="flex items-start gap-2 mb-3 bg-slate-900/40 p-2 rounded-md border border-slate-800/50">
            <GitCommit size={12} className="text-slate-500 mt-0.5 shrink-0" />
            <p className="text-[11px] text-slate-400 line-clamp-1 leading-relaxed italic">
              {project.latest_commit_msg}
            </p>
          </div>
        )}

        <div className="flex items-center gap-4 text-xs text-slate-500 mb-4">
          <span className="flex items-center gap-1.5">
            <GitBranch size={12} className="text-brand-400" />
            <span className="font-medium text-slate-400">{project.branch}</span>
          </span>
          {latestDeployment?.commit_sha && (
            <span className="flex items-center gap-1.5 font-mono text-[10px]">
              {truncate(latestDeployment.commit_sha, 7)}
            </span>
          )}
          {latestDeployment && (
            <span className="ml-auto opacity-70">{timeAgo(latestDeployment.created_at)}</span>
          )}
        </div>

        {/* Status Counts - NEW */}
        <div className="flex items-center gap-3 mb-4 py-2 border-y border-slate-800/30">
          {runningCount > 0 && (
            <div className="flex items-center gap-1 px-1.5 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              <span className="text-[9px] font-bold text-emerald-400">{runningCount} Running</span>
            </div>
          )}
          {buildingCount > 0 && (
            <div className="flex items-center gap-1 px-1.5 py-0.5 rounded bg-amber-500/10 border border-amber-500/20">
              <div className="w-1.5 h-1.5 rounded-full bg-amber-500 animate-bounce" />
              <span className="text-[9px] font-bold text-amber-400">{buildingCount} Building</span>
            </div>
          )}
          {failedCount > 0 && (
            <div className="flex items-center gap-1 px-1.5 py-0.5 rounded bg-red-500/10 border border-red-500/20">
              <div className="w-1.5 h-1.5 rounded-full bg-red-500" />
              <span className="text-[9px] font-bold text-red-400">{failedCount} Failed</span>
            </div>
          )}
        </div>

        <div className="flex gap-2">
          <Link
            href={`/dashboard/projects/${project.id}`}
            className="btn-secondary text-xs py-1 px-3"
          >
            View
          </Link>
          <Link
            href={`/dashboard/projects/${project.id}/editor`}
            className="btn-secondary text-xs py-1 px-3 flex items-center gap-1.5"
          >
            <Code2 size={12} />
            Editor
          </Link>
          {latestDeployment?.url && (
            <a
              href={latestDeployment.url}
              target="_blank"
              rel="noopener noreferrer"
              className="btn-secondary text-xs py-1 px-3 group/link"
            >
              <ExternalLink size={12} className="group-hover/link:text-brand-400 transition-colors" />
              Open
            </a>
          )}
          <Link
            href={`/dashboard/projects/${project.id}/deployments`}
            className="flex items-center gap-1 text-[10px] text-slate-500 hover:text-brand-400 transition-colors ml-auto group/hist font-medium"
          >
            <Rocket size={10} className="group-hover/hist:translate-y-[-1px] transition-transform" />
            History
          </Link>
        </div>
      </div>
    </div>
    {ConfirmModal}
    </>
  )
}
