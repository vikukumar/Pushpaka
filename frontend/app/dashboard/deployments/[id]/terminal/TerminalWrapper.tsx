'use client'

import dynamic from 'next/dynamic'
import { usePathname } from 'next/navigation'
import { Header } from '@/components/layout/Header'

const TerminalClient = dynamic(() => import('./TerminalClient'), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center h-64 text-slate-500 text-sm">
      Loading terminal…
    </div>
  ),
})

export default function TerminalWrapper() {
  const pathname = usePathname()
  const id = pathname.split('/')[3] || ''

  return (
    <div className="flex flex-col h-screen">
      <Header
        title="Web Terminal"
        subtitle={id ? `Deployment ${id.slice(0, 8)}` : ''}
      />
      <div
        className="flex-1 p-4"
        style={{ background: '#0d0f14' }}
      >
        <div
          className="h-full rounded-lg overflow-hidden"
          style={{ border: '1px solid rgba(99,102,241,0.25)', background: '#0d0f14' }}
        >
          {id && <TerminalClient deploymentId={id} />}
        </div>
      </div>
    </div>
  )
}
