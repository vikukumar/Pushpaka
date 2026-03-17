'use client'

import { useState, useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { projectsApi, webhooksApi } from '@/lib/api'
import { Project, WebhookConfig } from '@/types'
import { Header } from '@/components/layout/Header'
import toast from 'react-hot-toast'
import { Trash2, AlertTriangle, Lock, Eye, EyeOff, Save, Cpu, Link2, Plus, X, Copy, Check } from 'lucide-react'

export default function ProjectSettingsPage() {
  const pathname = usePathname()
  const id = pathname.split('/')[3] || ''
  const router = useRouter()
  const queryClient = useQueryClient()
  const [showToken, setShowToken] = useState(false)
  const [newToken, setNewToken] = useState('')
  const [isPrivate, setIsPrivate] = useState<boolean | null>(null)
  const [savingToken, setSavingToken] = useState(false)

  // Resource limits
  const [cpuLimit, setCpuLimit] = useState('')
  const [memoryLimit, setMemoryLimit] = useState('')
  const [restartPolicy, setRestartPolicy] = useState('unless-stopped')
  const [savingLimits, setSavingLimits] = useState(false)

  // Webhooks
  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([])
  const [webhookProvider, setWebhookProvider] = useState('github')
  const [webhookBranch, setWebhookBranch] = useState('')
  const [creatingWebhook, setCreatingWebhook] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  const { data: project } = useQuery<Project>({
    queryKey: ['project', id],
    queryFn: () => projectsApi.get(id).then((r) => r.data),
  })

  // Initialise toggle from fetched project (once only)
  useEffect(() => {
    if (project && isPrivate === null) {
      setIsPrivate(project.is_private ?? false)
      setCpuLimit(project.cpu_limit ?? '')
      setMemoryLimit(project.memory_limit ?? '')
      setRestartPolicy(project.restart_policy || 'unless-stopped')
    }
  }, [project, isPrivate])

  // Load webhooks for this project
  useEffect(() => {
    if (!id) return
    webhooksApi.list().then((r) => {
      const all: WebhookConfig[] = r.data?.data ?? []
      setWebhooks(all.filter((w) => w.project_id === id))
    }).catch(() => {})
  }, [id])

  const handleSaveLimits = async () => {
    setSavingLimits(true)
    try {
      await projectsApi.update(id, {
        cpu_limit: cpuLimit,
        memory_limit: memoryLimit,
        restart_policy: restartPolicy,
      })
      toast.success('Resource limits saved')
      queryClient.invalidateQueries({ queryKey: ['project', id] })
    } catch {
      toast.error('Failed to save resource limits')
    } finally {
      setSavingLimits(false)
    }
  }

  const handleCreateWebhook = async () => {
    setCreatingWebhook(true)
    try {
      const res = await webhooksApi.create({
        project_id: id,
        provider: webhookProvider,
        branch: webhookBranch || undefined,
      })
      setWebhooks((prev) => [...prev, res.data])
      toast.success('Webhook created')
    } catch {
      toast.error('Failed to create webhook')
    } finally {
      setCreatingWebhook(false)
    }
  }

  const handleDeleteWebhook = async (webhookId: string) => {
    try {
      await webhooksApi.delete(webhookId)
      setWebhooks((prev) => prev.filter((w) => w.id !== webhookId))
      toast.success('Webhook deleted')
    } catch {
      toast.error('Failed to delete webhook')
    }
  }

  const copyWebhookUrl = (webhookId: string, url: string) => {
    navigator.clipboard.writeText(url)
    setCopiedId(webhookId)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const effectiveIsPrivate = isPrivate !== null ? isPrivate : (project?.is_private ?? false)

  const handleSaveToken = async () => {
    setSavingToken(true)
    try {
      await projectsApi.update(id, {
        is_private: effectiveIsPrivate,
        git_token: newToken,
      })
      toast.success('Repository access updated')
      setNewToken('')
      queryClient.invalidateQueries({ queryKey: ['project', id] })
    } catch {
      toast.error('Failed to update token')
    } finally {
      setSavingToken(false)
    }
  }

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
                {project?.created_at ? new Date(project.created_at).toLocaleString() : ''}
              </dd>
            </div>
          </dl>
        </div>

        {/* Private repository access */}
        <div className="card">
          <div className="flex items-center gap-2 mb-4">
            <Lock size={16} className="text-brand-400" />
            <h3 className="text-sm font-semibold text-slate-300">Repository Access</h3>
          </div>

          {/* Private toggle */}
          <div className="flex items-center gap-3 p-3 rounded-lg border border-surface-border bg-surface-elevated mb-4">
            <div className="flex-1">
              <div className="text-sm font-medium text-slate-300">Private Repository</div>
              <div className="text-xs text-slate-500">Use a Personal Access Token when cloning</div>
            </div>
            <button
              type="button"
              onClick={() => setIsPrivate(!effectiveIsPrivate)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${effectiveIsPrivate ? 'bg-brand-600' : 'bg-slate-700'}`}
            >
              <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${effectiveIsPrivate ? 'translate-x-6' : 'translate-x-1'}`} />
            </button>
          </div>

          {/* Token input */}
          <div className="space-y-3">
            <div>
              <label className="label text-xs">New Personal Access Token</label>
              <p className="text-xs text-slate-600 mb-2">
                {project?.is_private ? 'A token is already stored. Enter a new one to replace it.' : 'Enter a token to enable private repo access.'}
              </p>
              <div className="relative">
                <input
                  type={showToken ? 'text' : 'password'}
                  className="input pr-10"
                  placeholder={project?.is_private ? ' (stored)' : 'ghp_xxxxxxxxxxxxxxxxxxxx'}
                  value={newToken}
                  onChange={(e) => setNewToken(e.target.value)}
                  autoComplete="off"
                />
                <button
                  type="button"
                  onClick={() => setShowToken(!showToken)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                >
                  {showToken ? <EyeOff size={14} /> : <Eye size={14} />}
                </button>
              </div>
              <p className="text-xs text-slate-600 mt-1">
                GitHub: Settings  Developer settings  Personal access tokens  Grant <code className="text-slate-400">repo</code> scope.
              </p>
            </div>

            <button
              onClick={handleSaveToken}
              disabled={savingToken || (!newToken && isPrivate === null)}
              className="btn-primary text-xs py-1.5"
            >
              {savingToken ? 'Saving...' : (
                <><Save size={13} /> Save Access Settings</>
              )}
            </button>
          </div>
        </div>

        {/* Resource Limits */}
        <div className="card">
          <div className="flex items-center gap-2 mb-4">
            <Cpu size={16} className="text-brand-400" />
            <h3 className="text-sm font-semibold text-slate-300">Resource Limits</h3>
          </div>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="label text-xs">CPU Limit</label>
                <input
                  type="text"
                  className="input w-full text-sm"
                  placeholder="e.g. 0.5  (half a core)"
                  value={cpuLimit}
                  onChange={(e) => setCpuLimit(e.target.value)}
                />
                <p className="text-xs text-slate-600 mt-1">Docker --cpus value. Leave blank for unlimited.</p>
              </div>
              <div>
                <label className="label text-xs">Memory Limit</label>
                <input
                  type="text"
                  className="input w-full text-sm"
                  placeholder="e.g. 512m or 2g"
                  value={memoryLimit}
                  onChange={(e) => setMemoryLimit(e.target.value)}
                />
                <p className="text-xs text-slate-600 mt-1">Docker --memory value. Leave blank for unlimited.</p>
              </div>
            </div>
            <div>
              <label className="label text-xs">Restart Policy</label>
              <select
                className="input w-full text-sm"
                value={restartPolicy}
                onChange={(e) => setRestartPolicy(e.target.value)}
              >
                <option value="unless-stopped">unless-stopped (default)</option>
                <option value="always">always</option>
                <option value="on-failure">on-failure</option>
                <option value="no">no</option>
              </select>
            </div>
            <button
              onClick={handleSaveLimits}
              disabled={savingLimits}
              className="btn-primary text-xs py-1.5"
            >
              {savingLimits ? 'Saving...' : <><Save size={13} /> Save Resource Limits</>}
            </button>
          </div>
        </div>

        {/* Webhooks */}
        <div className="card">
          <div className="flex items-center gap-2 mb-4">
            <Link2 size={16} className="text-brand-400" />
            <h3 className="text-sm font-semibold text-slate-300">Webhooks</h3>
          </div>
          <p className="text-xs text-slate-500 mb-4">
            Configure a webhook URL to give to GitHub or GitLab so pushes trigger auto-deployments.
          </p>

          {/* Existing webhooks */}
          {webhooks.length > 0 && (
            <div className="space-y-2 mb-4">
              {webhooks.map((wh) => (
                <div
                  key={wh.id}
                  className="flex items-center gap-2 p-3 rounded-lg text-xs"
                  style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
                >
                  <span className="text-slate-400 font-medium capitalize w-14">{wh.provider}</span>
                  <span className="text-slate-500">{wh.branch || 'any branch'}</span>
                  <span className="flex-1 font-mono text-slate-600 truncate">{wh.webhook_url}</span>
                  <button
                    onClick={() => copyWebhookUrl(wh.id, wh.webhook_url)}
                    className="btn-secondary py-0.5 px-1.5 text-xs"
                    title="Copy webhook URL"
                  >
                    {copiedId === wh.id ? <Check size={12} className="text-green-400" /> : <Copy size={12} />}
                  </button>
                  <button
                    onClick={() => handleDeleteWebhook(wh.id)}
                    className="btn-danger py-0.5 px-1.5 text-xs"
                  >
                    <X size={12} />
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Create webhook */}
          <div className="flex items-end gap-2">
            <div>
              <label className="label text-xs">Provider</label>
              <select
                className="input text-sm"
                value={webhookProvider}
                onChange={(e) => setWebhookProvider(e.target.value)}
              >
                <option value="github">GitHub</option>
                <option value="gitlab">GitLab</option>
              </select>
            </div>
            <div className="flex-1">
              <label className="label text-xs">Branch filter (optional)</label>
              <input
                type="text"
                className="input w-full text-sm"
                placeholder="main"
                value={webhookBranch}
                onChange={(e) => setWebhookBranch(e.target.value)}
              />
            </div>
            <button
              onClick={handleCreateWebhook}
              disabled={creatingWebhook}
              className="btn-primary text-xs py-1.5 px-3"
            >
              {creatingWebhook ? '...' : <><Plus size={13} /> Add Webhook</>}
            </button>
          </div>
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


// Required for Next.js static export with dynamic segments.
export function generateStaticParams() {
  return []
}

