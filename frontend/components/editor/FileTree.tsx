'use client'

import { useState } from 'react'
import { ChevronDown, ChevronRight, Folder, FolderOpen, Edit2, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface FileEntry {
  name: string
  path: string
  is_dir: boolean
  size?: number
  children?: FileEntry[]
}

const EXT_COLORS: Record<string, string> = {
  ts: '#3178c6', tsx: '#3178c6', mts: '#3178c6', cts: '#3178c6',
  js: '#f0db4f', jsx: '#f0db4f', mjs: '#f0db4f', cjs: '#f0db4f',
  go: '#00ADD8', py: '#3572A5', rs: '#dea584', rb: '#CC342D',
  java: '#b07219', kt: '#7f52ff', cs: '#178600', php: '#4F5D95',
  swift: '#F05138', c: '#555555', cpp: '#f34b7d',
  json: '#f0db4f', jsonc: '#f0db4f',
  css: '#264de4', scss: '#cc6699', less: '#1d365d',
  html: '#e34c26', htm: '#e34c26', xml: '#e37933',
  md: '#083fa1', mdx: '#083fa1',
  sh: '#89e051', bash: '#89e051', zsh: '#89e051',
  yaml: '#cb171e', yml: '#cb171e', toml: '#9c4221',
  sql: '#e38c00', env: '#ecd53f', gitignore: '#f54d27',
  dockerfile: '#384d54',
}

export function FileIcon({ name, size = 13 }: { name: string; size?: number }) {
  const lower = name.toLowerCase()
  const isDockerfile = lower === 'dockerfile' || lower.startsWith('dockerfile.')
  const ext = isDockerfile ? 'dockerfile' : (lower.split('.').pop() ?? '')
  const color = EXT_COLORS[ext] ?? '#9da5b4'
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" className="shrink-0">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
        stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <polyline points="14,2 14,8 20,8"
        stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

interface TreeNodeProps {
  entry: FileEntry
  depth: number
  activePath: string | null
  onSelect: (e: FileEntry) => void
  onDelete?: (path: string) => void
}

function TreeNode({ entry, depth, activePath, onSelect, onDelete }: TreeNodeProps) {
  const [open, setOpen] = useState(depth === 0)

  if (entry.is_dir) {
    return (
      <div>
        <button
          onClick={() => setOpen((v) => !v)}
          className="flex items-center gap-1 w-full text-left transition-colors hover:bg-[rgba(255,255,255,0.05)] group"
          style={{ paddingLeft: `${4 + depth * 12}px`, paddingTop: 3, paddingBottom: 3 }}
        >
          <span className="text-[#c8c8c8] opacity-70 group-hover:opacity-100">
            {open ? <ChevronDown size={10} /> : <ChevronRight size={10} />}
          </span>
          {open
            ? <FolderOpen size={13} className="shrink-0" style={{ color: '#dcb67a' }} />
            : <Folder size={13} className="shrink-0" style={{ color: '#dcb67a' }} />
          }
          <span className="text-[12px] text-[#cccccc] truncate ml-1">{entry.name}</span>
        </button>
        {open && entry.children?.map((child) => (
          <TreeNode
            key={child.path}
            entry={child}
            depth={depth + 1}
            activePath={activePath}
            onSelect={onSelect}
            onDelete={onDelete}
          />
        ))}
      </div>
    )
  }

  const isActive = activePath === entry.path
  return (
    <button
      onClick={() => onSelect(entry)}
      className={cn(
        'flex items-center gap-1.5 w-full text-left cursor-default transition-colors group',
        isActive ? 'bg-[#37373d]' : 'hover:bg-[rgba(255,255,255,0.05)]',
      )}
      style={{ paddingLeft: `${18 + depth * 12}px`, paddingTop: 3, paddingBottom: 3 }}
    >
      <FileIcon name={entry.name} />
      <span className={cn('flex-1 text-[12px] truncate ml-0.5', isActive ? 'text-white' : 'text-[#cccccc]')}>
        {entry.name}
      </span>
      <div className="hidden group-hover:flex items-center gap-1 pr-2">
        <button
          onClick={(e) => { e.stopPropagation(); /* TODO: Rename */ }}
          className="p-0.5 hover:bg-[#45454d] rounded text-[#858585] hover:text-[#cccccc]"
        >
          <Edit2 size={11} />
        </button>
        <button
          onClick={(e) => { e.stopPropagation(); onDelete?.(entry.path) }}
          className="p-0.5 hover:bg-[#45454d] rounded text-[#858585] hover:text-[#f48771]"
        >
          <Trash2 size={11} />
        </button>
      </div>
    </button>
  )
}

export function FileTree({ files, activePath, onSelect, onDelete }: {
  files: FileEntry[]
  activePath: string | null
  onSelect: (e: FileEntry) => void
  onDelete?: (path: string) => void
}) {
  return (
    <div className="flex-1 overflow-y-auto group scrollbar-thin">
      {files.map((entry) => (
        <TreeNode
          key={entry.path}
          entry={entry}
          depth={0}
          activePath={activePath}
          onSelect={onSelect}
          onDelete={onDelete}
        />
      ))}
    </div>
  )
}
