'use client'

import { useQuery } from '@tanstack/react-query'
import { useQueryClient } from '@tanstack/react-query'
import { domainsApi, projectsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { Domain, Project } from '@/types'
import { Globe, CheckCircle2, XCircle, Trash2, Loader2 } from 'lucide-react'
import toast from 'react-hot-toast'

export default function DomainsPage() {
  const queryClient = useQueryClient()

  const { data: domainsData, isLoading } = useQuery({
    queryKey: ['domains'],
    queryFn: () => domainsApi.list().then((r) => r.data),
  })

  const { data: projectsData } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectsApi.list().then((r) => r.data),
  })

  const domains: Domain[] = domainsData?.data || []
  const projects: Project[] = projectsData?.data || []

  const getProject = (id: string) => projects.find((p) => p.id === id)

  const handleDelete = async (id: string) => {
    try {
      await domainsApi.delete(id)
      toast.success('Domain removed')
      queryClient.invalidateQueries({ queryKey: ['domains'] })
    } catch {
      toast.error('Failed to remove domain')
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Domains" subtitle="All custom domains across your projects" />

      <div className="p-6">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 size={24} className="animate-spin text-brand-400" />
          </div>
        ) : domains.length === 0 ? (
          <div className="card text-center py-12">
            <Globe size={40} className="mx-auto text-slate-700 mb-3" />
            <p className="text-slate-400 text-sm">No custom domains configured</p>
            <p className="text-slate-600 text-xs mt-1">
              Add custom domains from a project&apos;s Domains tab
            </p>
          </div>
        ) : (
          <div className="card p-0 overflow-hidden">
            <div className="divide-y divide-surface-border">
              {domains.map((d) => {
                const project = getProject(d.project_id)
                return (
                  <div key={d.id} className="flex items-center gap-4 px-5 py-3.5">
                    <Globe size={15} className="text-slate-600 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-mono text-slate-200">{d.domain}</div>
                      <div className="text-xs text-slate-500">{project?.name || 'Unknown project'}</div>
                    </div>
                    {d.verified ? (
                      <span className="flex items-center gap-1.5 text-xs text-emerald-400">
                        <CheckCircle2 size={12} /> Verified
                      </span>
                    ) : (
                      <span className="flex items-center gap-1.5 text-xs text-amber-400">
                        <XCircle size={12} /> Pending DNS
                      </span>
                    )}
                    <button
                      onClick={() => handleDelete(d.id)}
                      className="text-slate-600 hover:text-red-400 transition-colors ml-2"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                )
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
