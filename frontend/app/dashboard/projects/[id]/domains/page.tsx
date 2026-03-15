'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { domainsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { Domain } from '@/types'
import toast from 'react-hot-toast'
import { Plus, Trash2, CheckCircle2, XCircle, Loader2, Globe } from 'lucide-react'

export default function ProjectDomainsPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [domain, setDomain] = useState('')
  const [adding, setAdding] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['domains', 'project', id],
    queryFn: () => domainsApi.list().then((r) => r.data),
  })

  const domains: Domain[] = (data?.data || []).filter((d: Domain) => d.project_id === id)

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await domainsApi.add({ project_id: id, domain })
      toast.success('Domain added')
      queryClient.invalidateQueries({ queryKey: ['domains'] })
      setDomain('')
      setAdding(false)
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      toast.error(error?.response?.data?.error || 'Failed to add domain')
    }
  }

  const handleDelete = async (domainId: string) => {
    try {
      await domainsApi.delete(domainId)
      toast.success('Domain removed')
      queryClient.invalidateQueries({ queryKey: ['domains'] })
    } catch {
      toast.error('Failed to remove domain')
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Custom Domains" subtitle="Configure custom domains for this project" />

      <div className="p-6 max-w-2xl space-y-4">
        <div className="flex justify-between items-center">
          <p className="text-sm text-slate-400">
            Add a CNAME record pointing to <code className="text-brand-300 text-xs px-1 py-0.5 bg-brand-500/10 rounded">api.pushpaka.app</code>
          </p>
          <button className="btn-primary text-sm" onClick={() => setAdding(!adding)}>
            <Plus size={14} />
            Add Domain
          </button>
        </div>

        {adding && (
          <div className="card border-brand-500/30">
            <form onSubmit={handleAdd} className="space-y-3">
              <h3 className="text-sm font-semibold text-white">Add Custom Domain</h3>
              <div>
                <label className="label">Domain</label>
                <input
                  type="text"
                  className="input"
                  placeholder="app.yourdomain.com"
                  value={domain}
                  onChange={(e) => setDomain(e.target.value)}
                  required
                />
              </div>
              <div className="flex gap-2">
                <button type="submit" className="btn-primary text-sm">Add Domain</button>
                <button type="button" className="btn-secondary text-sm" onClick={() => setAdding(false)}>Cancel</button>
              </div>
            </form>
          </div>
        )}

        {isLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 size={20} className="animate-spin text-brand-400" />
          </div>
        ) : domains.length === 0 ? (
          <div className="card text-center py-12">
            <Globe size={32} className="mx-auto text-slate-700 mb-3" />
            <p className="text-slate-400 text-sm">No custom domains configured</p>
          </div>
        ) : (
          <div className="card p-0 overflow-hidden">
            <div className="divide-y divide-surface-border">
              {domains.map((d) => (
                <div key={d.id} className="flex items-center gap-4 px-5 py-3.5">
                  <Globe size={15} className="text-slate-600 shrink-0" />
                  <span className="text-sm text-slate-200 flex-1 font-mono">{d.domain}</span>
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
                    className="text-slate-600 hover:text-red-400 transition-colors"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
