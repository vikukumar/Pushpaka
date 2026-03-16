'use client'

import { Header } from '@/components/layout/Header'
import { Activity } from 'lucide-react'

export default function ActivityPage() {
  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Activity" subtitle="Recent activity across all projects" />
      <div className="p-6">
        <div className="card text-center py-16">
          <Activity size={40} className="mx-auto text-slate-700 mb-4" />
          <p className="text-slate-400 text-sm">Activity feed coming soon.</p>
        </div>
      </div>
    </div>
  )
}
