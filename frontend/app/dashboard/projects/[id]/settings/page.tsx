'use client'

import { useParams, useRouter } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { projectsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import toast from 'react-hot-toast'
import { Trash2, AlertTriangle } from 'lucide-react'

export default function ProjectSettingsPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const queryClient = useQueryClient()

  const { data: project } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectsApi.get(id).then((r) => r.data),
  })

  const handleDelete = async () => {
    if (!confirm(`Delete project "${project?.name}"? This cannot be undone.`)) return
    try {
      await projectsApi.delete(id)
      toast.success('Project deleted')
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      router.push('/dashboard/projects')
    } catch {
      toast.error('Failed to delete project')
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Settings" subtitle="Manage project settings" />

      <div className="p-6 max-w-2xl space-y-6">
        {/* Project info */}
        <div className="card">
          <h3 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4">
            Project Information
          </h3>
          <dl className="space-y-2">
            <div className="flex gap-3">
              <dt className="text-slate-500 text-sm w-24">ID</dt>
              <dd className="text-sm text-slate-300 font-mono">{project?.id}</dd>
            </div>
            <div className="flex gap-3">
              <dt className="text-slate-500 text-sm w-24">Name</dt>
              <dd className="text-sm text-slate-300">{project?.name}</dd>
            </div>
            <div className="flex gap-3">
              <dt className="text-slate-500 text-sm w-24">Created</dt>
              <dd className="text-sm text-slate-300">
                {project?.created_at ? new Date(project.created_at).toLocaleString() : '—'}
              </dd>
            </div>
          </dl>
        </div>

        {/* Danger zone */}
        <div className="card border-red-500/20">
          <div className="flex items-start gap-3 mb-4">
            <AlertTriangle size={18} className="text-red-400 shrink-0 mt-0.5" />
            <div>
              <h3 className="text-sm font-semibold text-red-400">Danger Zone</h3>
              <p className="text-xs text-slate-500 mt-0.5">
                These actions are irreversible. Please be certain.
              </p>
            </div>
          </div>

          <div className="flex items-center justify-between p-3 rounded-lg border border-red-500/20 bg-red-500/5">
            <div>
              <div className="text-sm text-slate-300 font-medium">Delete this project</div>
              <div className="text-xs text-slate-500">
                All deployments, logs, and configuration will be permanently removed.
              </div>
            </div>
            <button onClick={handleDelete} className="btn-danger">
              <Trash2 size={14} />
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
