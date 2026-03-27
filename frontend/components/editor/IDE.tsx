'use client'

import dynamic from 'next/dynamic'
import { useState, useCallback, useRef, useEffect, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { projectsApi, filesApi, systemFilesApi, editorStateApi, aiApi } from '@/lib/api'
import { cn } from '@/lib/utils'
import { 
  Loader2, Files, Search, GitBranch, Puzzle, FileCode as FileCodeIcon, X, Plus, FolderPlus, RotateCw,
  Sparkles, Send, Image
} from 'lucide-react'
import toast from 'react-hot-toast'
import { useConfirm, usePrompt } from '@/components/ui/Modal'
import { FileTree, FileEntry, FileIcon } from './FileTree'
import { SearchPanel, GitPanel, ExtensionsPanel } from './Panels'
import { ProjectSelector } from './ProjectSelector'
import { useEditorWS, WSMessage } from '@/hooks/useEditorWS'

const MonacoEditor = dynamic(() => import('@monaco-editor/react'), { ssr: false })

function detectLanguage(name: string): string {
  const lower = name.toLowerCase()
  if (lower === 'dockerfile' || lower.startsWith('dockerfile.')) return 'dockerfile'
  const ext = lower.split('.').pop() ?? ''
  const map: Record<string, string> = {
    ts: 'typescript', tsx: 'typescript', js: 'javascript', jsx: 'javascript',
    mjs: 'javascript', cjs: 'javascript', mts: 'typescript', cts: 'typescript',
    py: 'python', go: 'go', rs: 'rust', rb: 'ruby', java: 'java', kt: 'kotlin',
    json: 'json', jsonc: 'json',
    yaml: 'yaml', yml: 'yaml', toml: 'ini',
    html: 'html', htm: 'html', css: 'css', scss: 'scss', less: 'less',
    md: 'markdown', mdx: 'markdown', sh: 'shell', bash: 'shell', zsh: 'shell',
    sql: 'sql', xml: 'xml', env: 'ini', gitignore: 'ini',
    c: 'c', cpp: 'cpp', cs: 'csharp', php: 'php', swift: 'swift',
  }
  return map[ext] ?? 'plaintext'
}

interface OpenTab {
  path: string
  name: string
  content: string
  originalContent: string
  isBinary?: boolean
}

interface IDEProps {
  initialMode: 'system' | 'project'
  initialProjectId?: string
  initialFilePath?: string
}

function CopilotPanel({ workspaceId, selectedPath, getCode }: { workspaceId: string; selectedPath: string | null; getCode: () => string }) {
  const [messages, setMessages] = useState<{ role: 'user' | 'ai'; text: string }[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  const send = useCallback(async () => {
    const msg = input.trim()
    if (!msg) return
    setInput('')
    const code = getCode()
    const context = selectedPath ? `\n\nFile: ${selectedPath}\n\`\`\`\n${code.slice(0, 2000)}\n\`\`\`` : ''
    setMessages(m => [...m, { role: 'user', text: msg }])
    setLoading(true)
    try {
      const res = await aiApi.chat(`${msg}${context}`)
      setMessages(m => [...m, { role: 'ai', text: res.data?.reply || '...' }])
      setTimeout(() => bottomRef.current?.scrollIntoView({ behavior: 'smooth' }), 50)
    } catch {
      setMessages(m => [...m, { role: 'ai', text: 'AI failed' }])
    } finally { setLoading(false) }
  }, [input, selectedPath, getCode])

  return (
    <div className="w-80 bg-[#252526] border-l border-[#3e3e42] flex flex-col">
      <div className="h-9 px-3 flex items-center gap-2 border-b border-[#3e3e42] text-[11px] font-bold uppercase text-[#bbbbbe]">
        <Sparkles size={14} className="text-brand-400" /> Copilot
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-3 scrollbar-thin">
        {messages.map((m, i) => (
          <div key={i} className={cn("p-2 rounded text-[12px]", m.role === 'user' ? "bg-brand-500/10 ml-4" : "bg-[#2d2d30] mr-4 text-[#cccccc]")}>
            <pre className="whitespace-pre-wrap font-sans">{m.text}</pre>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <div className="p-2 border-t border-[#3e3e42] flex gap-2">
        <input className="flex-1 bg-[#3c3c3c] border border-[#3e3e42] rounded px-2 py-1 text-[12px] text-white outline-none" value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && send()} placeholder="Ask AI..." />
        <button onClick={send} disabled={loading} className="bg-brand-600 p-1.5 rounded"><Send size={14} className="text-white" /></button>
      </div>
    </div>
  )
}

export default function IDE({ initialMode, initialProjectId, initialFilePath }: IDEProps) {
  const [mode, setMode] = useState(initialMode)
  // Use pathname to resolve ID if placeholder '_' is passed from Next.js static routing
  const resolvedProjectId = useMemo(() => {
    if (typeof window === 'undefined') return initialProjectId
    if (initialProjectId && initialProjectId !== '_') return initialProjectId
    const parts = window.location.pathname.split('/')
    const idx = parts.indexOf('projects')
    if (idx !== -1 && parts[idx+1] && parts[idx+1] !== 'editor') return parts[idx+1]
    return initialProjectId
  }, [initialProjectId])

  const [projectId, setProjectId] = useState(resolvedProjectId)
  const [activeSidebarTab, setActiveSidebarTab] = useState<'explorer' | 'search' | 'git' | 'extensions' | null>('explorer')
  const [showCopilot, setShowCopilot] = useState(false)
  const [tabs, setTabs] = useState<OpenTab[]>([])
  const [activePath, setActivePath] = useState<string | null>(null)
  const [cursorPos, setCursorPos] = useState({ line: 1, col: 1 })
  const [loadingFile, setLoadingFile] = useState(false)
  const [saving, setSaving] = useState(false)

  // Fetch project details for display name
  const { data: project } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.get(projectId!).then((r: any) => r.data),
    enabled: !!projectId && mode === 'project'
  })

  const { confirm, Component: ConfirmModal } = useConfirm()
  const { prompt, Component: PromptModal } = usePrompt()

  const workspaceId = useMemo(() => (mode === 'system' && !projectId) ? 'system' : projectId!, [mode, projectId])

  // WebSocket for sync
  const handleWSMessage = useCallback((msg: WSMessage) => {
    if (msg.type === 'file:update' && msg.path) {
      setTabs(prev => prev.map(t => t.path === msg.path ? { ...t, content: msg.content || '', originalContent: msg.content || '' } : t))
      if (activePath === msg.path) toast(`File updated by ${msg.user || 'another user'}`, { icon: '🔄' })
    }
  }, [activePath])

  const { sendMessage } = useEditorWS(workspaceId, handleWSMessage)

  const editorRef = useRef<{ getValue: () => string } | null>(null)

  // API selection
  const apiClient = useMemo(() => {
    if (mode === 'system' && !projectId) {
      return {
        ...systemFilesApi,
        refetch: async () => {} // No-op for system files
      }
    }
    return {
      list: () => filesApi.list(projectId!),
      read: (path: string) => filesApi.read(projectId!, path),
      save: (path: string, content: string) => filesApi.save(projectId!, path, content),
      createFile: (path: string) => filesApi.createFile(projectId!, path),
      createDirectory: (path: string) => filesApi.createDirectory(projectId!, path),
      delete: (path: string) => filesApi.delete(projectId!, path),
      refetch: () => filesApi.sync(projectId!)
    }
  }, [mode, projectId])

  const { data: filesData, isLoading: treeLoading, refetch: refreshTree } = useQuery({
    queryKey: ['files', workspaceId],
    queryFn: () => apiClient.list().then(r => r.data as { files: FileEntry[] }),
    enabled: !!workspaceId && (mode === 'system' || !!projectId)
  })
  const files: FileEntry[] = filesData?.files ?? []

  const openFile = useCallback(async (entry: FileEntry) => {
    if (entry.is_dir) return
    const existing = tabs.find(t => t.path === entry.path)
    if (existing) { setActivePath(entry.path); return }

    setLoadingFile(true)
    try {
      const res = await apiClient.read(entry.path)
      const content = res.data.content || ''
      setTabs(prev => [...prev, { 
        path: entry.path, 
        name: entry.name, 
        content, 
        originalContent: content,
        isBinary: res.data.is_binary 
      }])
      setActivePath(entry.path)
    } catch (err: any) {
      if (err.response?.status === 415 || err.response?.data?.is_binary) {
        setTabs(prev => [...prev, { 
          path: entry.path, 
          name: entry.name, 
          content: '', 
          originalContent: '', 
          isBinary: true 
        }])
        setActivePath(entry.path)
      } else {
        toast.error(`Cannot open ${entry.name}`)
      }
    } finally {
      setLoadingFile(false)
    }
  }, [tabs, apiClient])

  // Handle initial file path
  useEffect(() => {
    if (initialFilePath && files.length > 0) {
      const findPath = (list: FileEntry[]): FileEntry | null => {
        for (const f of list) {
          if (f.path === initialFilePath) return f
          if (f.children) {
            const found = findPath(f.children)
            if (found) return found
          }
        }
        return null
      }
      const entry = findPath(files)
      if (entry) openFile(entry)
    }
  }, [initialFilePath, files, openFile])

  const handleSave = useCallback(async () => {
    const activeTab = tabs.find(t => t.path === activePath)
    if (!activeTab) return
    const current = editorRef.current?.getValue() ?? activeTab.content
    setSaving(true)
    try {
      await apiClient.save(activeTab.path, current)
      setTabs(prev => prev.map(t => t.path === activeTab.path ? { ...t, originalContent: current } : t))
      sendMessage({ type: 'file:write', path: activeTab.path, content: current })
      toast.success('Saved')
    } catch {
      toast.error('Save failed')
    } finally {
      setSaving(false)
    }
  }, [activePath, tabs, apiClient, sendMessage])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault()
        handleSave()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleSave])

  const handleCreateFile = async () => {
    const name = await prompt({ title: 'New File', message: 'Enter file path:', placeholder: 'index.js' })
    if (!name) return
    try {
      await apiClient.createFile(name)
      refreshTree()
    } catch (e: any) { toast.error(e.response?.data?.error || 'Failed') }
  }

  const handleSync = async () => {
    if (mode === 'system' || !projectId) {
      refreshTree()
      return
    }
    const t = toast.loading('Syncing with repository...')
    try {
      await apiClient.refetch()
      await refreshTree()
      toast.success('Synced', { id: t })
    } catch (e: any) {
      toast.error('Sync failed', { id: t })
    }
  }

  const handleDelete = async (path: string) => {
    const ok = await confirm({ title: 'Delete', message: `Delete ${path}?`, type: 'error' })
    if (!ok) return
    try {
      await apiClient.delete(path)
      setTabs(prev => prev.filter(t => t.path !== path))
      refreshTree()
    } catch (e: any) { toast.error('Delete failed') }
  }

  const activeTab = tabs.find(t => t.path === activePath)

  if (mode === 'project' && !projectId) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#1e1e1e] text-[#858585]">
        <Loader2 className="animate-spin mr-2" /> Initializing project...
      </div>
    )
  }

  return (
    <div className="flex-1 flex overflow-hidden bg-[#1e1e1e]">
      {/* Activity Bar */}
      <div className="w-12 bg-[#333333] flex flex-col items-center py-2 gap-4 border-r border-[#252526]">
        {(['explorer', 'search', 'git', 'extensions'] as const).map(tab => {
          const Icon = { explorer: Files, search: Search, git: GitBranch, extensions: Puzzle }[tab]
          return (
            <button
              key={tab}
              onClick={() => setActiveSidebarTab(v => v === tab ? null : tab)}
              className={cn("p-2 transition-all", activeSidebarTab === tab ? "text-white border-l-2 border-white" : "text-[#858585] hover:text-[#cccccc]")}
            >
              <Icon size={24} />
            </button>
          )
        })}
        <div className="flex-1" />
        <button
          onClick={() => setShowCopilot(v => !v)}
          className={cn("p-2 transition-all mb-2", showCopilot ? "text-brand-400" : "text-[#858585] hover:text-[#cccccc]")}
        >
          <Sparkles size={24} />
        </button>
      </div>

      {/* Sidebar */}
      {activeSidebarTab && (
        <div className="w-64 bg-[#252526] flex flex-col border-r border-[#3e3e42]">
          {activeSidebarTab === 'explorer' && (
            <>
              {mode === 'system' && (
                <ProjectSelector 
                  currentProjectId={projectId} 
                  onSelect={(id) => { setProjectId(id); setTabs([]); setActivePath(null) }}
                  onSelectSystem={() => { setProjectId(undefined); setTabs([]); setActivePath(null) }}
                />
              )}
              <div className="h-9 px-3 flex items-center justify-between border-b border-[#3e3e42] uppercase text-[11px] font-bold tracking-widest text-[#bbbbbe]">
                <span className="truncate max-w-[150px]">
                  Explorer {project ? `(${project.name})` : projectId ? `(${projectId.slice(0, 8)})` : ''}
                </span>
                <div className="flex gap-1">
                  <button onClick={handleSync} title="Sync with Repository" className="p-1 hover:bg-[#37373d] rounded"><RotateCw size={14} /></button>
                  <button onClick={handleCreateFile} title="New File" className="p-1 hover:bg-[#37373d] rounded"><Plus size={14} /></button>
                </div>
              </div>
              <FileTree files={files} activePath={activePath} onSelect={openFile} onDelete={handleDelete} />
            </>
          )}
          {activeSidebarTab === 'search' && <SearchPanel files={files} onSelect={openFile} />}
          {activeSidebarTab === 'git' && <GitPanel mode={projectId ? 'project' : 'system'} />}
          {activeSidebarTab === 'extensions' && <ExtensionsPanel />}
        </div>
      )}

      {/* Editor Main */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Tabs */}
        <div className="h-9 bg-[#2d2d2d] flex items-end overflow-x-auto border-b border-[#252526] no-scrollbar">
          {tabs.map(tab => (
            <div
              key={tab.path}
              onClick={() => setActivePath(tab.path)}
              className={cn(
                "h-full px-3 flex items-center gap-2 border-r border-[#252526] cursor-default text-[12px] min-w-[120px] max-w-[200px] transition-colors group/tab",
                tab.path === activePath ? "bg-[#1e1e1e] text-white border-t-2 border-brand-500" : "bg-[#2d2d2d] text-[#9d9d9d] hover:bg-[#333333]"
              )}
            >
              <FileIcon name={tab.name} />
              <span className="truncate flex-1">{tab.name}</span>
              <button 
                onClick={(e) => { e.stopPropagation(); setTabs(prev => prev.filter(t => t.path !== tab.path)) }} 
                className="p-0.5 hover:bg-[#45454d] rounded opacity-0 group-hover/tab:opacity-100"
              >
                <X size={12}/>
              </button>
            </div>
          ))}
        </div>

        {/* Monaco */}
        <div className="flex-1 relative bg-[#1e1e1e]">
          {loadingFile && <div className="absolute inset-0 z-50 bg-slate-900/40 backdrop-blur-[1px] flex items-center justify-center"><Loader2 size={32} className="text-brand-500 animate-spin" /></div>}
          {activeTab?.isBinary ? (
            <div className="h-full flex flex-col items-center justify-center text-slate-400 gap-4 bg-[#1e1e1e]">
              <Image size={48} className="text-slate-600" strokeWidth={1} />
              <div className="text-center">
                <p className="text-sm font-medium">{activeTab.name}</p>
                <p className="text-xs text-slate-500 mt-1 italic">Binary file — preview not supported</p>
              </div>
            </div>
          ) : activeTab ? (
            <MonacoEditor
              key={`${workspaceId}-${activePath}`}
              height="100%"
              theme="vs-dark"
              language={detectLanguage(activeTab.name)}
              value={activeTab.content}
              onMount={(editor) => { editorRef.current = editor as any }}
              onChange={(v) => {
                setTabs(prev => prev.map(t => t.path === activeTab.path ? { ...t, content: v ?? '' } : t))
                sendMessage({ type: 'cursor:move', path: activeTab.path, cursor: cursorPos })
              }}
              options={{ fontSize: 13, minimap: { enabled: true }, automaticLayout: true }}
            />
          ) : (
            <div className="h-full flex flex-col items-center justify-center text-slate-600 gap-4">
              <FileCodeIcon size={48} strokeWidth={1} />
              <p className="italic text-xs">Select a file to start editing</p>
            </div>
          )}
        </div>

        {/* Status Bar */}
        <div className="h-6 bg-brand-700 text-white flex items-center justify-between px-3 text-[11px] shrink-0 font-medium">
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1.5"><FileCodeIcon size={12}/> {projectId ? 'Project Workspace' : 'System Workspace'}</span>
            {activePath && <span className="opacity-80">/deploy{activePath}</span>}
          </div>
          <div className="flex items-center gap-4">
            {saving && <Loader2 size={12} className="animate-spin" />}
            <span>Workspace: {project?.name || workspaceId}</span>
          </div>
        </div>
      </div>

      {showCopilot && (
        <CopilotPanel 
          workspaceId={workspaceId} 
          selectedPath={activePath} 
          getCode={() => editorRef.current?.getValue() || activeTab?.content || ''} 
        />
      )}

      {ConfirmModal}
      {PromptModal}
    </div>
  )
}
