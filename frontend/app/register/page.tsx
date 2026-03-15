'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { authApi } from '@/lib/api'
import { useAuthStore } from '@/lib/auth'
import { AuthResponse } from '@/types'
import toast from 'react-hot-toast'
import { Loader2 } from 'lucide-react'

export default function RegisterPage() {
  const router = useRouter()
  const { setAuth } = useAuthStore()
  const [form, setForm] = useState({ email: '', name: '', password: '' })
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (form.password.length < 8) {
      toast.error('Password must be at least 8 characters')
      return
    }
    setLoading(true)
    try {
      const { data } = await authApi.register(form)
      const auth: AuthResponse = data
      setAuth(auth.user, auth.token)
      toast.success('Account created!')
      router.push('/dashboard')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      toast.error(error?.response?.data?.error || 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-surface flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center gap-3 mb-4">
            <div className="w-10 h-10 bg-brand-600 rounded-xl flex items-center justify-center">
              <svg viewBox="0 0 32 32" className="w-6 h-6" fill="none">
                <rect x="3" y="15" width="26" height="10" rx="5" fill="white" fillOpacity="0.9"/>
                <circle cx="8" cy="15" r="4" fill="white" fillOpacity="0.9"/>
                <circle cx="16" cy="12" r="5" fill="white" fillOpacity="0.9"/>
                <circle cx="24" cy="14" r="4" fill="white" fillOpacity="0.9"/>
                <path d="M3 18 Q0 16 1 20 Q0 23 3 22" fill="#22d3ee" opacity="0.9"/>
                <path d="M29 18 Q32 16 31 20 Q32 23 29 22" fill="#22d3ee" opacity="0.9"/>
              </svg>
            </div>
            <span className="text-2xl font-bold text-white">Pushpaka</span>
          </div>
          <p className="text-slate-400 text-sm">Create your account to get started</p>
        </div>

        <div className="card">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="label">Name</label>
              <input
                type="text"
                className="input"
                placeholder="Your name"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
                minLength={2}
              />
            </div>

            <div>
              <label className="label">Email</label>
              <input
                type="email"
                className="input"
                placeholder="you@example.com"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                required
              />
            </div>

            <div>
              <label className="label">Password</label>
              <input
                type="password"
                className="input"
                placeholder="Min. 8 characters"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                required
                minLength={8}
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full justify-center py-2.5"
            >
              {loading ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  Creating account...
                </>
              ) : (
                'Create account'
              )}
            </button>
          </form>

          <p className="text-center text-sm text-slate-500 mt-4">
            Already have an account?{' '}
            <Link href="/login" className="text-brand-400 hover:text-brand-300 font-medium">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
