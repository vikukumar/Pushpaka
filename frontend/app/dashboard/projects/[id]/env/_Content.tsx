'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { envApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import { EnvVar } from '@/types'
import toast from 'react-hot-toast'
import { Plus, Trash2, Eye, EyeOff, Loader2, Key } from 'lucide-react'

export default function ProjectEnvPage() {
  const pathname = usePathname()
  const id = pathname.split('/')[3] || ''
  const queryClient = useQueryClient()
  const [adding, setAdding] = useState(false)
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')
  const [showValue, setShowValue] = useState(false)
  const [saveLoading, setSaveLoading] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['env', id],
    queryFn: () => envApi.list(id).then((r) => r.data),
  })

  const envVars: EnvVar[] = data?.data || []

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newKey.trim()) return
    setSaveLoading(true)
    try {
      await envApi.set({ project_id: id, key: newKey.trim(), value: newValue })
      toast.success('Variable saved')
      queryClient.invalidateQueries({ queryKey: ['env', id] })
      setNewKey('')
      setNewValue('')
      setAdding(false)
    } catch {
      toast.error('Failed to save variable')
    } finally {
      setSaveLoading(false)
    }
  }

  const handleDelete = async (key: string) => {
    try {
      await envApi.delete({ project_id: id, key })
      toast.success('Variable deleted')
      queryClient.invalidateQueries({ queryKey: ['env', id] })
    } catch {
      toast.error('Failed to delete variable')
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Environment Variables" subtitle="Manage project environment variables" />

      <div className="p-6 max-w-3xl space-y-4">
        <div className="flex justify-between items-center">
          <p className="text-sm text-slate-400">
            Variables are injected at runtime. Values are not shown after saving.
          </p>
          <button className="btn-primary text-sm" onClick={() => setAdding(!adding)}>
            <Plus size={14} />
            Add Variable
          </button>
        </div>

        {/* Add form */}
        {adding && (
          <div className="card border-brand-500/30">
            <form onSubmit={handleAdd} className="space-y-4">
              <h3 className="text-sm font-semibold text-white">New Variable</h3>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="label">Key</label>
                  <input
                    type="text"
                    className="input font-mono"
                    placeholder="DATABASE_URL"
                    value={newKey}
                    onChange={(e) => setNewKey(e.target.value.toUpperCase().replace(/\s/g, '_'))}
                    required
                  />
                </div>
                <div>
                  <label className="label">Value</label>
                  <div className="relative">
                    <input
                      type={showValue ? 'text' : 'password'}
                      className="input font-mono pr-10"
                      placeholder="Value"
                      value={newValue}
                      onChange={(e) => setNewValue(e.target.value)}
                    />
                    <button
                      type="button"
                      onClick={() => setShowValue(!showValue)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500"
                    >
                      {showValue ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                  </div>
                </div>
              </div>
              <div className="flex gap-2">
                <button type="submit" disabled={saveLoading} className="btn-primary text-sm">
                  {saveLoading ? <Loader2 size={14} className="animate-spin" /> : 'Save'}
                </button>
                <button
                  type="button"
                  className="btn-secondary text-sm"
                  onClick={() => { setAdding(false); setNewKey(''); setNewValue('') }}
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Variables list */}
        {isLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 size={20} className="animate-spin text-brand-400" />
          </div>
        ) : envVars.length === 0 ? (
          <div className="card text-center py-12">
            <Key size={32} className="mx-auto text-slate-700 mb-3" />
            <p className="text-slate-400 text-sm">No environment variables set</p>
          </div>
        ) : (
          <div className="card p-0 overflow-hidden">
            <div className="divide-y divide-surface-border">
              {envVars.map((v) => (
                <div key={v.id} className="flex items-center gap-4 px-5 py-3">
                  <Key size={14} className="text-slate-600 shrink-0" />
                  <span className="font-mono text-sm text-slate-200 flex-1">{v.key}</span>
                  <span className="text-xs text-slate-600 font-mono">
                    {v.has_value ? '••••••••' : '(empty)'}
                  </span>
                  <button
                    onClick={() => handleDelete(v.key)}
                    className="text-slate-600 hover:text-red-400 transition-colors ml-2"
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
