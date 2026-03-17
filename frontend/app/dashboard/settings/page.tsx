'use client'

import { useAuthStore } from '@/lib/auth'
import { Header } from '@/components/layout/Header'
import { User, Shield, Key, Bell } from 'lucide-react'

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

        {/* Notifications */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Bell size={16} className="text-brand-400" />
            <h3 className="text-sm font-semibold text-white">Notifications</h3>
          </div>
          <div className="space-y-3">
            {[
              'Deployment success',
              'Deployment failure',
              'Custom domain verified',
            ].map((label) => (
              <label key={label} className="flex items-center gap-3 cursor-pointer">
                <input type="checkbox" defaultChecked className="w-4 h-4 rounded accent-brand-500" />
                <span className="text-sm text-slate-300">{label}</span>
              </label>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
