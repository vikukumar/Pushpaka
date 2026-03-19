'use client'

import { useQuery } from '@tanstack/react-query'
import { projectsApi } from '@/lib/api'
import { Loader2, Box, Globe, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

export function ProjectSelector({ currentProjectId, onSelect, onSelectSystem }: {
  currentProjectId?: string
  onSelect: (id: string) => void
  onSelectSystem: () => void
}) {
  const { data: projects, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectsApi.list().then(r => r.data)
  })

  return (
    <div className="flex flex-col border-b border-[#3e3e42] bg-[#252526]">
      <div className="px-3 py-2 uppercase text-[10px] font-bold tracking-widest text-[#bbbbbe]/50">
        Workspace Context
      </div>
      <div className="px-2 pb-2 space-y-0.5">
        <button
          onClick={onSelectSystem}
          className={cn(
            "w-full flex items-center gap-2 px-2 py-1.5 rounded text-[12px] transition-all group",
            !currentProjectId ? "bg-brand-500/10 text-brand-300 ring-1 ring-brand-500/30" : "text-[#cccccc] hover:bg-white/5"
          )}
        >
          <Globe size={14} className={cn(!currentProjectId ? "text-brand-400" : "text-slate-500")} />
          <span className="flex-1 text-left font-medium">Global System Root</span>
          {!currentProjectId && <ChevronRight size={12} className="text-brand-500/50" />}
        </button>

        <div className="pt-1">
          <div className="px-2 py-1 text-[9px] font-bold text-slate-600 uppercase">Projects</div>
          {isLoading ? (
            <div className="flex items-center gap-2 px-2 py-1.5 grayscale opacity-50">
              <Loader2 size={12} className="animate-spin" />
              <span className="text-[11px]">Loading...</span>
            </div>
          ) : (
            projects?.data?.map((p: any) => (
              <button
                key={p.id}
                onClick={() => onSelect(p.id)}
                className={cn(
                  "w-full flex items-center gap-2 px-2 py-1.5 rounded text-[12px] transition-all group",
                  currentProjectId === p.id ? "bg-white/10 text-white shadow-sm" : "text-[#8e8e8e] hover:bg-white/5 hover:text-[#cccccc]"
                )}
              >
                <Box size={14} className={cn(currentProjectId === p.id ? "text-blue-400" : "text-slate-600")} />
                <span className="flex-1 text-left truncate">{p.name}</span>
                {currentProjectId === p.id && <div className="w-1 h-1 rounded-full bg-blue-500" />}
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
