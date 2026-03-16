'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { projectsApi } from '@/lib/api'
import { Header } from '@/components/layout/Header'
import toast from 'react-hot-toast'
import { Loader2, GitBranch, Terminal, Globe, ChevronDown, Lock, Eye, EyeOff } from 'lucide-react'

const FRAMEWORKS = [
  { value: '', label: 'Auto-detect' },
  { value: 'nextjs', label: 'Next.js' },
  { value: 'react', label: 'React (Vite/CRA)' },
  { value: 'vue', label: 'Vue.js' },
  { value: 'nodejs', label: 'Node.js' },
  { value: 'python', label: 'Python (Flask/FastAPI)' },
  { value: 'go', label: 'Go' },
  { value: 'static', label: 'Static HTML' },
  { value: 'docker', label: 'Docker (custom Dockerfile)' },
]

export default function NewProjectPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [showToken, setShowToken] = useState(false)
  const [form, setForm] = useState({
    name: '',
    repo_url: '',
    branch: 'main',
    build_command: '',
    start_command: '',
    port: 3000,
    framework: '',
    is_private: false,
    git_token: '',
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      await projectsApi.create(form)
      toast.success('Project created!')
      router.push('/dashboard/projects')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      toast.error(error?.response?.data?.error || 'Failed to create project')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="New Project" subtitle="Connect a Git repository to deploy" />
      <div className="p-6 max-w-2xl">
        <div className="card">
          <form onSubmit={handleSubmit} className="space-y-5">
            {/* Name */}
            <div>
              <label className="label">Project Name</label>
              <input
                type="text"
                className="input"
                placeholder="my-awesome-app"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
                minLength={2}
                maxLength={64}
              />
            </div>

            {/* Repository URL */}
            <div>
              <label className="label">
                <Globe size={13} className="inline mr-1.5" />
                Git Repository URL
              </label>
              <input
                type="url"
                className="input"
                placeholder="https://github.com/user/repo"
                value={form.repo_url}
                onChange={(e) => setForm({ ...form, repo_url: e.target.value })}
                required
              />
            </div>

            {/* Private repo toggle */}
            <div className="flex items-center gap-3 p-3 rounded-lg border border-surface-border bg-surface-elevated">
              <Lock size={15} className="text-slate-500 shrink-0" />
              <div className="flex-1">
                <div className="text-sm font-medium text-slate-300">Private Repository</div>
                <div className="text-xs text-slate-500">Enable to provide a Personal Access Token</div>
              </div>
              <button
                type="button"
                onClick={() => setForm({ ...form, is_private: !form.is_private, git_token: form.is_private ? '' : form.git_token })}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${form.is_private ? 'bg-brand-600' : 'bg-slate-700'}`}
              >
                <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${form.is_private ? 'translate-x-6' : 'translate-x-1'}`} />
              </button>
            </div>

            {/* PAT field (only when private) */}
            {form.is_private && (
              <div>
                <label className="label">
                  <Lock size={13} className="inline mr-1.5" />
                  Personal Access Token (PAT)
                </label>
                <div className="relative">
                  <input
                    type={showToken ? 'text' : 'password'}
                    className="input pr-10"
                    placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
                    value={form.git_token}
                    onChange={(e) => setForm({ ...form, git_token: e.target.value })}
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
                  GitHub: Settings → Developer settings → Personal access tokens → Grant <code className="text-slate-400">repo</code> scope.
                  The token is stored securely and never shown again.
                </p>
              </div>
            )}

            {/* Branch */}
            <div>
              <label className="label">
                <GitBranch size={13} className="inline mr-1.5" />
                Branch
              </label>
              <input
                type="text"
                className="input"
                placeholder="main"
                value={form.branch}
                onChange={(e) => setForm({ ...form, branch: e.target.value })}
              />
            </div>

            {/* Framework */}
            <div>
              <label className="label">Framework / Runtime</label>
              <div className="relative">
                <select
                  className="input appearance-none pr-8"
                  value={form.framework}
                  onChange={(e) => setForm({ ...form, framework: e.target.value })}
                >
                  {FRAMEWORKS.map(({ value, label }) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </select>
                <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 pointer-events-none" />
              </div>
            </div>

            {/* Build & Start commands */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="label">
                  <Terminal size={13} className="inline mr-1.5" />
                  Build Command
                </label>
                <input
                  type="text"
                  className="input"
                  placeholder="npm run build"
                  value={form.build_command}
                  onChange={(e) => setForm({ ...form, build_command: e.target.value })}
                />
              </div>
              <div>
                <label className="label">Start Command</label>
                <input
                  type="text"
                  className="input"
                  placeholder="npm start"
                  value={form.start_command}
                  onChange={(e) => setForm({ ...form, start_command: e.target.value })}
                />
              </div>
            </div>

            {/* Port */}
            <div>
              <label className="label">Port</label>
              <input
                type="number"
                className="input"
                placeholder="3000"
                value={form.port}
                onChange={(e) => setForm({ ...form, port: parseInt(e.target.value) || 3000 })}
                min={1}
                max={65535}
              />
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={() => router.back()}
                className="btn-secondary"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                className="btn-primary"
              >
                {loading ? (
                  <>
                    <Loader2 size={15} className="animate-spin" />
                    Creating...
                  </>
                ) : (
                  'Create Project'
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}

