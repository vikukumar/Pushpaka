'use client'

import { useSearchParams } from 'next/navigation'
import { Header } from '@/components/layout/Header'
import IDE from '@/components/editor/IDE'
import { Suspense } from 'react'
import { Loader2 } from 'lucide-react'

function EditorContent() {
  const searchParams = useSearchParams()
  const initialFile = searchParams?.get('file')

  return (
    <div className="flex flex-col h-screen bg-[#1e1e1e] text-[#cccccc]">
      <Header title="System Editor" subtitle="Global workspace management" />
      <IDE initialMode="system" initialFilePath={initialFile || undefined} />
    </div>
  )
}

export default function GlobalEditorPage() {
  return (
    <Suspense fallback={
      <div className="h-screen flex items-center justify-center bg-[#1e1e1e]">
        <Loader2 className="animate-spin text-brand-500" size={32} />
      </div>
    }>
      <EditorContent />
    </Suspense>
  )
}
