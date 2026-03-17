'use client'

import { useState, useEffect } from 'react'
import { useAuthStore } from '@/lib/auth'
import { Header } from '@/components/layout/Header'
import { User, Shield, Key, Bell, Loader2, CheckCircle, ChevronDown, ChevronUp } from 'lucide-react'
import { notificationsApi } from '@/lib/api'
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

        {/* Notifications — live */}
        <NotificationsSection />
      </div>
    </div>
  )
}
