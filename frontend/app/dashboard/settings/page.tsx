'use client'

import { useState, useEffect } from 'react'
import { useAuthStore } from '@/lib/auth'
import { Header } from '@/components/layout/Header'
import { User, Shield, Key, Bell, Loader2, CheckCircle, ChevronDown, ChevronUp, Sparkles, Eye, EyeOff } from 'lucide-react'
import { Select } from '@/components/ui/Select'
import { notificationsApi, aiApi } from '@/lib/api'
import toast from 'react-hot-toast'

interface NotifConfig {
  slack_webhook_url: string
  discord_webhook_url: string
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_from: string
  smtp_to: string
  notify_on_success: boolean
  notify_on_failure: boolean
}

function NotificationsSection() {
  const [config, setConfig] = useState<NotifConfig>({
    slack_webhook_url: '',
    discord_webhook_url: '',
    smtp_host: '',
    smtp_port: 587,
    smtp_username: '',
    smtp_password: '',
    smtp_from: '',
    smtp_to: '',
    notify_on_success: true,
    notify_on_failure: true,
  })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [tab, setTab] = useState<'slack' | 'discord' | 'smtp'>('slack')
  const [showSmtpPass, setShowSmtpPass] = useState(false)

  useEffect(() => {
    notificationsApi.getConfig()
      .then((res) => {
        if (res.data) setConfig((prev) => ({ ...prev, ...res.data }))
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      await notificationsApi.updateConfig(config)
      toast.success('Notification settings saved')
    } catch {
      toast.error('Failed to save notification settings')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return (
    <div className="card flex items-center gap-2 text-slate-400 text-sm">
      <Loader2 size={14} className="animate-spin" /> Loading notification settings…
    </div>
  )

  const tabs: { key: 'slack' | 'discord' | 'smtp'; label: string }[] = [
    { key: 'slack',   label: 'Slack' },
    { key: 'discord', label: 'Discord' },
    { key: 'smtp',    label: 'Email (SMTP)' },
  ]

  return (
    <div className="card">
      <div className="flex items-center gap-3 mb-5">
        <Bell size={16} className="text-brand-400" />
        <h3 className="text-sm font-semibold text-white">Notifications</h3>
      </div>

      {/* Channel toggles */}
      <div className="flex gap-1 mb-5 p-1 rounded-lg" style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.07)' }}>
        {tabs.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`flex-1 py-1.5 px-2 rounded-md text-xs font-medium transition-all ${
              tab === key
                ? 'text-white'
                : 'text-slate-500 hover:text-slate-300'
            }`}
            style={tab === key ? { background: 'rgba(99,102,241,0.25)', border: '1px solid rgba(99,102,241,0.35)' } : {}}
          >
            {label}
          </button>
        ))}
      </div>

      {tab === 'slack' && (
        <div className="space-y-3">
          <label className="block text-xs text-slate-400">Slack Incoming Webhook URL</label>
          <input
            type="url"
            className="input w-full text-sm"
            placeholder="https://hooks.slack.com/services/…"
            value={config.slack_webhook_url}
            onChange={(e) => setConfig({ ...config, slack_webhook_url: e.target.value })}
          />
        </div>
      )}

      {tab === 'discord' && (
        <div className="space-y-3">
          <label className="block text-xs text-slate-400">Discord Webhook URL</label>
          <input
            type="url"
            className="input w-full text-sm"
            placeholder="https://discord.com/api/webhooks/…"
            value={config.discord_webhook_url}
            onChange={(e) => setConfig({ ...config, discord_webhook_url: e.target.value })}
          />
        </div>
      )}

      {tab === 'smtp' && (
        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-slate-400 mb-1">Host</label>
              <input
                type="text"
                className="input w-full text-sm"
                placeholder="smtp.example.com"
                value={config.smtp_host}
                onChange={(e) => setConfig({ ...config, smtp_host: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-xs text-slate-400 mb-1">Port</label>
              <input
                type="number"
                className="input w-full text-sm"
                placeholder="587"
                value={config.smtp_port}
                onChange={(e) => setConfig({ ...config, smtp_port: Number(e.target.value) })}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-slate-400 mb-1">Username</label>
              <input
                type="text"
                className="input w-full text-sm"
                placeholder="you@example.com"
                value={config.smtp_username}
                onChange={(e) => setConfig({ ...config, smtp_username: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-xs text-slate-400 mb-1">Password</label>
              <div className="flex gap-1">
                <input
                  type={showSmtpPass ? 'text' : 'password'}
                  className="input w-full text-sm"
                  placeholder="••••••••"
                  value={config.smtp_password}
                  onChange={(e) => setConfig({ ...config, smtp_password: e.target.value })}
                />
                <button
                  type="button"
                  className="btn-secondary text-xs px-2"
                  onClick={() => setShowSmtpPass((v) => !v)}
                >
                  {showSmtpPass ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
                </button>
              </div>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-slate-400 mb-1">From address</label>
              <input
                type="email"
                className="input w-full text-sm"
                placeholder="noreply@example.com"
                value={config.smtp_from}
                onChange={(e) => setConfig({ ...config, smtp_from: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-xs text-slate-400 mb-1">To address</label>
              <input
                type="email"
                className="input w-full text-sm"
                placeholder="alerts@example.com"
                value={config.smtp_to}
                onChange={(e) => setConfig({ ...config, smtp_to: e.target.value })}
              />
            </div>
          </div>
        </div>
      )}

      {/* Event preferences */}
      <div className="mt-5 pt-4 space-y-2" style={{ borderTop: '1px solid rgba(255,255,255,0.07)' }}>
        <p className="text-xs text-slate-500 mb-3">Send notifications on:</p>
        {([
          { key: 'notify_on_success', label: 'Deployment success' },
          { key: 'notify_on_failure', label: 'Deployment failure' },
        ] as { key: keyof NotifConfig; label: string }[]).map(({ key, label }) => (
          <label key={key} className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              className="w-4 h-4 rounded accent-brand-500"
              checked={config[key] as boolean}
              onChange={(e) => setConfig({ ...config, [key]: e.target.checked })}
            />
            <span className="text-sm text-slate-300">{label}</span>
          </label>
        ))}
      </div>

      <button
        onClick={handleSave}
        disabled={saving}
        className="btn mt-5 text-sm flex items-center gap-2"
      >
        {saving ? <Loader2 size={14} className="animate-spin" /> : <CheckCircle size={14} />}
        Save Notification Settings
      </button>
    </div>
  )
}

// ── AI Provider settings section ─────────────────────────────────────────────
const AI_PROVIDERS = [
  { value: 'openai',      label: 'OpenAI' },
  { value: 'openrouter',  label: 'OpenRouter' },
  { value: 'gemini',      label: 'Google Gemini' },
  { value: 'anthropic',   label: 'Anthropic' },
  { value: 'ollama',      label: 'Ollama (local)' },
]

function AIProviderSection() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [showKey, setShowKey] = useState(false)
  const [form, setForm] = useState({
    provider: 'openai',
    api_key: '',
    model: '',
    base_url: '',
    system_prompt: '',
    monitoring_enabled: false,
    monitoring_interval: 300,
    autonomous_agent: false,
  })
  const [maskedKey, setMaskedKey] = useState('')

  useEffect(() => {
    aiApi.getConfig()
      .then((res) => {
        const d = res.data
        if (d) {
          setForm((prev) => ({
            ...prev,
            provider: d.provider || 'openai',
            model: d.model || '',
            base_url: d.base_url || '',
            system_prompt: d.system_prompt || '',
            monitoring_enabled: !!d.monitoring_enabled,
            monitoring_interval: d.monitoring_interval || 300,
            autonomous_agent: !!d.autonomous_agent,
          }))
          setMaskedKey(d.api_key_masked || '')
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      await aiApi.saveConfig(form)
      toast.success('AI settings saved')
      // Re-fetch to get the masked key
      const res = await aiApi.getConfig()
      setMaskedKey(res.data?.api_key_masked || '')
      if (form.api_key) {
        setForm((f) => ({ ...f, api_key: '' })) // clear plaintext after save
      }
    } catch {
      toast.error('Failed to save AI settings')
    } finally {
      setSaving(false)
    }
  }

  const modelPlaceholder: Record<string, string> = {
    openai:     'gpt-4o-mini',
    openrouter: 'openai/gpt-4o-mini',
    gemini:     'gemini-1.5-flash',
    anthropic:  'claude-3-haiku-20240307',
    ollama:     'llama3',
  }

  if (loading) return (
    <div className="card flex items-center gap-2 text-slate-400 text-sm">
      <Loader2 size={14} className="animate-spin" /> Loading AI settings…
    </div>
  )

  return (
    <div className="card">
      <div className="flex items-center gap-3 mb-5">
        <Sparkles size={16} className="text-brand-400" />
        <h3 className="text-sm font-semibold text-white">AI Provider</h3>
        <span className="text-xs text-slate-500 ml-auto">Powers log analysis &amp; chat assistant</span>
      </div>

      <div className="space-y-4">
        {/* Provider */}
        <div>
          <label className="block text-xs text-slate-500 mb-1.5">Provider</label>
          <Select
            value={form.provider}
            onChange={(v) => setForm((f) => ({ ...f, provider: v }))}
            options={AI_PROVIDERS}
          />
        </div>

        {/* API Key */}
        <div>
          <label className="block text-xs text-slate-500 mb-1.5">API Key</label>
          <div className="relative">
            <input
              type={showKey ? 'text' : 'password'}
              className="input pr-10 font-mono text-sm"
              placeholder={maskedKey || 'sk-…  (leave blank to keep existing)'}
              value={form.api_key}
              onChange={(e) => setForm((f) => ({ ...f, api_key: e.target.value }))}
              autoComplete="new-password"
            />
            <button
              type="button"
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
              onClick={() => setShowKey((v) => !v)}
            >
              {showKey ? <EyeOff size={14} /> : <Eye size={14} />}
            </button>
          </div>
          {maskedKey && (
            <p className="text-xs text-slate-600 mt-1">Current: {maskedKey}</p>
          )}
        </div>

        {/* Model */}
        <div>
          <label className="block text-xs text-slate-500 mb-1.5">Model</label>
          <input
            className="input font-mono text-sm"
            placeholder={modelPlaceholder[form.provider] || 'model-name'}
            value={form.model}
            onChange={(e) => setForm((f) => ({ ...f, model: e.target.value }))}
          />
        </div>

        {/* Base URL (optional — for Ollama / OpenRouter / self-hosted) */}
        {['ollama', 'openrouter', 'openai'].includes(form.provider) && (
          <div>
            <label className="block text-xs text-slate-500 mb-1.5">
              Base URL <span className="text-slate-600">(optional override)</span>
            </label>
            <input
              className="input font-mono text-sm"
              placeholder={form.provider === 'ollama' ? 'http://localhost:11434' : 'https://openrouter.ai/api/v1'}
              value={form.base_url}
              onChange={(e) => setForm((f) => ({ ...f, base_url: e.target.value }))}
            />
          </div>
        )}

        {/* System prompt */}
        <div>
          <label className="block text-xs text-slate-500 mb-1.5">
            Custom System Prompt <span className="text-slate-600">(optional)</span>
          </label>
          <textarea
            className="input resize-none text-sm"
            rows={3}
            placeholder="You are a helpful deployment assistant…"
            value={form.system_prompt}
            onChange={(e) => setForm((f) => ({ ...f, system_prompt: e.target.value }))}
          />
        </div>

        {/* Monitoring toggle */}
        <div className="flex items-center gap-3">
          <input
            id="ai-monitor"
            type="checkbox"
            className="accent-brand-500 w-4 h-4"
            checked={form.monitoring_enabled}
            onChange={(e) => setForm((f) => ({ ...f, monitoring_enabled: e.target.checked }))}
          />
          <label htmlFor="ai-monitor" className="text-sm text-slate-300 select-none cursor-pointer">
            Enable background AI monitoring (auto-analyse failed deployments)
          </label>
        </div>

        {/* Autonomous toggle */}
        <div className="flex items-center gap-3">
          <input
            id="ai-autonomous"
            type="checkbox"
            className="accent-brand-500 w-4 h-4"
            checked={form.autonomous_agent}
            onChange={(e) => setForm((f) => ({ ...f, autonomous_agent: e.target.checked }))}
          />
          <label htmlFor="ai-autonomous" className="text-sm text-slate-300 select-none cursor-pointer">
            Enable autonomous agent mode (allow AI to restart, sync, and promote without asking)
          </label>
        </div>
      </div>

      <button
        className="btn-primary mt-5 text-sm"
        onClick={handleSave}
        disabled={saving}
      >
        {saving ? <Loader2 size={14} className="animate-spin" /> : <CheckCircle size={14} />}
        Save AI Settings
      </button>
    </div>
  )
}

export default function SettingsPage() {
  const { user } = useAuthStore()

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Account Settings" subtitle="Manage your account and preferences" />

      <div className="p-6 max-w-2xl space-y-5">
        {/* Profile */}
        <div className="card">
          <div className="flex items-center gap-3 mb-5">
            <User size={16} className="text-brand-400" />
            <h3 className="text-sm font-semibold text-white">Profile</h3>
          </div>
          <dl className="space-y-3">
            {[
              { label: 'Name', value: user?.name },
              { label: 'Email', value: user?.email },
              { label: 'Role', value: user?.role },
              { label: 'User ID', value: user?.id, mono: true },
            ].map(({ label, value, mono }) => (
              <div key={label} className="flex items-center gap-3">
                <dt className="text-slate-500 text-sm w-20">{label}</dt>
                <dd className={`text-sm text-slate-200 ${mono ? 'font-mono text-xs' : ''}`}>{value}</dd>
              </div>
            ))}
          </dl>
        </div>

        {/* Security */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Shield size={16} className="text-brand-400" />
            <h3 className="text-sm font-semibold text-white">Security</h3>
          </div>
          <p className="text-sm text-slate-400">
            Use a strong, unique password. Your JWT session expires after 24 hours.
          </p>
          <button className="btn-secondary mt-3 text-sm">Change Password</button>
        </div>

        {/* API Keys */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Key size={16} className="text-brand-400" />
            <h3 className="text-sm font-semibold text-white">API Key</h3>
          </div>
          <p className="text-sm text-slate-400 mb-3">
            Use your API key to authenticate programmatic requests to the Pushpaka API.
          </p>
          <div className="flex items-center gap-2">
            <input
              type="password"
              className="input font-mono text-xs flex-1"
              value="sk-"
              readOnly
            />
            <button className="btn-secondary text-sm">Reveal</button>
            <button className="btn-secondary text-sm">Regenerate</button>
          </div>
        </div>

        {/* AI Provider — live */}
        <AIProviderSection />

        {/* Notifications — live */}
        <NotificationsSection />
      </div>
    </div>
  )
}
