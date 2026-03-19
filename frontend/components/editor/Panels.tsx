'use client'

import { useState, useCallback } from 'react'
import { Search, GitBranch, Puzzle, Sparkles, FileCode as FileCodeIcon, Loader2 } from 'lucide-react'
import { FileIcon, FileEntry } from './FileTree'
import { cn } from '@/lib/utils'

// ─── Search Panel ─────────────────────────────────────────────────────────────
export function SearchPanel({ files, onSelect }: { files: FileEntry[]; onSelect: (e: FileEntry) => void }) {
  const [query, setQuery] = useState('')
  
  const flatten = useCallback((list: FileEntry[]): FileEntry[] => {
    let res: FileEntry[] = []
    for (const f of list) {
      if (!f.is_dir) res.push(f)
      if (f.children) res = [...res, ...flatten(f.children)]
    }
    return res
  }, [])

  const allFiles = flatten(files)
  const filtered = allFiles.filter(f => f.name.toLowerCase().includes(query.toLowerCase()) || f.path.toLowerCase().includes(query.toLowerCase())).slice(0, 50)

  return (
    <div className="flex flex-col h-full bg-[#252526]">
      <div className="px-3 py-2 uppercase text-[11px] font-semibold tracking-widest text-[#bbbbbe]" style={{ borderBottom: '1px solid #3e3e42' }}>
        Search Files
      </div>
      <div className="p-3 space-y-4">
        <div className="relative">
          <input
            className="w-full rounded px-2 py-1.5 text-[12px] text-[#cccccc] placeholder-[#6e6e6e] outline-none transition-all duration-200 focus:ring-1 focus:ring-brand-500/50"
            style={{ background: '#3c3c3c', border: '1px solid #3e3e42' }}
            placeholder="Search by name or path..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
          />
        </div>
        
        <div className="space-y-1 overflow-y-auto max-h-[calc(100vh-180px)] scrollbar-thin pr-1">
          {query.length > 0 && filtered.length === 0 && (
            <p className="text-[11px] text-[#6e6e6e] p-2 italic text-center">No matches found</p>
          )}
          {filtered.map(f => (
            <button
              key={f.path}
              onClick={() => onSelect(f)}
              className="w-full text-left p-2 hover:bg-[#37373d] rounded group transition-colors"
            >
              <div className="flex items-center gap-2">
                <FileIcon name={f.name} size={14} />
                <span className="text-[12px] text-[#cccccc] truncate font-medium group-hover:text-white">{f.name}</span>
              </div>
              <div className="text-[10px] text-[#6e6e6e] truncate ml-6 mt-0.5 opacity-60 group-hover:opacity-100">{f.path}</div>
            </button>
          ))}
          {!query && <p className="text-[11px] text-[#6e6e6e] p-4 text-center leading-relaxed">Type to search for files across the workspace.</p>}
        </div>
      </div>
    </div>
  )
}

// ─── Git Panel ────────────────────────────────────────────────────────────────
export function GitPanel({ mode }: { mode: 'system' | 'project' }) {
  return (
    <div className="flex flex-col h-full bg-[#252526]">
      <div className="px-3 py-2 uppercase text-[11px] font-semibold tracking-widest text-[#bbbbbe]" style={{ borderBottom: '1px solid #3e3e42' }}>
        Source Control
      </div>
      <div className="p-8 text-center flex flex-col items-center gap-4">
        <GitBranch size={48} className="text-slate-700" />
        <div className="space-y-2">
          <p className="text-[13px] text-slate-400 font-medium">
            {mode === 'system' ? 'System Mode' : 'No Repository Hooked'}
          </p>
          <p className="text-[11px] text-slate-500 leading-relaxed">
            {mode === 'system' 
              ? 'The global system editor operates directly on the server\'s deployment root. Select a project to use Git features.'
              : 'Local git features are syncing... commit and push directly from here.'}
          </p>
        </div>
        {mode === 'project' && (
          <button className="bg-brand-600 hover:bg-brand-500 text-white rounded px-4 py-1.5 text-[11px] transition-colors shadow-lg shadow-brand-500/10">
            Initialize Repository
          </button>
        )}
      </div>
    </div>
  )
}

// ─── Extensions Panel ─────────────────────────────────────────────────────────
export function ExtensionsPanel() {
  const extensions = [
    { name: 'Monaco Editor', version: 'v0.34.1', desc: 'Core editor functionality', icon: <FileCodeIcon size={16} className="text-blue-400" />, active: true },
    { name: 'Pushpaka AI', version: 'v1.2.0', desc: 'Intelligent code analysis', icon: <Sparkles size={16} className="text-brand-400" />, active: true },
    { name: 'Go Support', version: 'v0.38.0', desc: 'Rich Go language support', icon: <div className="w-4 h-4 rounded-full bg-cyan-500 flex items-center justify-center text-[8px] text-white font-bold">GO</div>, active: true },
    { name: 'Tailwind CSS', version: 'v3.3.0', desc: 'Intelligent CSS tooling', icon: <div className="w-4 h-4 rounded-full bg-sky-400 flex items-center justify-center text-[10px] text-white font-bold">T</div>, active: false },
  ]

  return (
    <div className="flex flex-col h-full bg-[#252526]">
      <div className="px-3 py-2 uppercase text-[11px] font-semibold tracking-widest text-[#bbbbbe]" style={{ borderBottom: '1px solid #3e3e42' }}>
        Extensions
      </div>
      <div className="px-3 py-2">
        <input 
          className="w-full bg-[#3c3c3c] border border-[#3e3e42] rounded px-2 py-1 text-[11px] text-white outline-none"
          placeholder="Search Extensions in Marketplace..."
        />
      </div>
      <div className="p-3 space-y-4 overflow-y-auto scrollbar-thin">
        <div className="text-[10px] font-bold text-[#bbbbbe] uppercase tracking-wider mb-2">Installed</div>
        {extensions.map(ext => (
          <div key={ext.name} className="flex gap-3 p-2 hover:bg-[#37373d] rounded group transition-all cursor-pointer border border-transparent hover:border-white/5">
            <div className="shrink-0 mt-1">{ext.icon}</div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center justify-between gap-2">
                <span className="text-[12px] font-medium text-[#cccccc] truncate group-hover:text-white">{ext.name}</span>
                <span className="text-[10px] text-slate-500">{ext.version}</span>
              </div>
              <p className="text-[11px] text-[#6e6e6e] line-clamp-1">{ext.desc}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
