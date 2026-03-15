'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { projectsApi, deploymentsApi } from '@/lib/api'
import { Project, Deployment } from '@/types'
import { Header } from '@/components/layout/Header'
import { ProjectCard } from '@/components/dashboard/ProjectCard'
import { StatusBadge } from '@/components/dashboard/StatusBadge'
import { timeAgo } from '@/lib/utils'
import { Plus, Rocket, FolderGit2, Activity, TrendingUp } from 'lucide-react'

export default function DashboardPage() {
  const { data: projectsData } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectsApi.list().then((r) => r.data),
  })

  const { data: deploymentsData } = useQuery({
    queryKey: ['deployments'],
    queryFn: () => deploymentsApi.list(10).then((r) => r.data),
  })

  const projects: Project[] = projectsData?.data || []
  const deployments: Deployment[] = deploymentsData?.data || []

  const stats = {
    totalProjects: projects.length,
    activeDeployments: deployments.filter((d) => d.status === 'running').length,
    buildingDeployments: deployments.filter((d) => d.status === 'building').length,
    failedDeployments: deployments.filter((d) => d.status === 'failed').length,
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Overview" subtitle="Your deployment platform dashboard" />

      <div className="p-6 space-y-6">
        {/* Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Projects', value: stats.totalProjects, icon: FolderGit2, color: 'text-brand-400' },
            { label: 'Running', value: stats.activeDeployments, icon: Activity, color: 'text-emerald-400' },
            { label: 'Building', value: stats.buildingDeployments, icon: TrendingUp, color: 'text-amber-400' },
            { label: 'Failed', value: stats.failedDeployments, icon: Rocket, color: 'text-red-400' },
          ].map(({ label, value, icon: Icon, color }) => (
            <div key={label} className="card">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm text-slate-500">{label}</span>
                <Icon size={18} className={color} />
              </div>
              <div className="text-3xl font-bold text-white">{value}</div>
            </div>
          ))}
        </div>

        {/* Projects & recent deployments side by side */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Recent Projects */}
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-base font-semibold text-white">Projects</h2>
              <Link href="/dashboard/projects/new" className="btn-primary text-xs py-1.5">
                <Plus size={14} />
                New project
              </Link>
            </div>

            {projects.length === 0 ? (
              <div className="card text-center py-12">
                <FolderGit2 size={32} className="mx-auto text-slate-600 mb-3" />
                <p className="text-slate-400 text-sm">No projects yet</p>
                <Link href="/dashboard/projects/new" className="btn-primary mt-4 inline-flex">
                  <Plus size={14} />
                  Create your first project
                </Link>
              </div>
            ) : (
              <div className="space-y-3">
                {projects.slice(0, 4).map((project) => {
                  const latestDeployment = deployments.find(
                    (d) => d.project_id === project.id
                  )
                  return (
                    <ProjectCard
                      key={project.id}
                      project={project}
                      latestDeployment={latestDeployment}
                    />
                  )
                })}
              </div>
            )}
          </div>

          {/* Recent Deployments */}
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-base font-semibold text-white">Recent Deployments</h2>
              <Link href="/dashboard/deployments" className="text-xs text-slate-500 hover:text-slate-300">
                View all
              </Link>
            </div>

            {deployments.length === 0 ? (
              <div className="card text-center py-12">
                <Rocket size={32} className="mx-auto text-slate-600 mb-3" />
                <p className="text-slate-400 text-sm">No deployments yet</p>
              </div>
            ) : (
              <div className="card p-0 overflow-hidden">
                <div className="divide-y divide-surface-border">
                  {deployments.slice(0, 8).map((d) => {
                    const project = projects.find((p) => p.id === d.project_id)
                    return (
                      <Link
                        key={d.id}
                        href={`/dashboard/deployments/${d.id}`}
                        className="flex items-center gap-3 px-4 py-3 hover:bg-slate-800/50 transition-colors"
                      >
                        <StatusBadge status={d.status} />
                        <div className="flex-1 min-w-0">
                          <div className="text-sm text-slate-200 font-medium truncate">
                            {project?.name || 'Unknown project'}
                          </div>
                          <div className="text-xs text-slate-500">{d.branch} · {timeAgo(d.created_at)}</div>
                        </div>
                        <div className="text-xs text-slate-600 font-mono">{d.id.slice(0, 8)}</div>
                      </Link>
                    )
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
