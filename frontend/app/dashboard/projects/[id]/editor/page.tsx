'use client'

import dynamic from 'next/dynamic'
import { useState, useCallback, useRef, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { filesApi, aiApi } from '@/lib/api'
import { cn } from '@/lib/utils'
import {
  Loader2, Save, ChevronRight, ChevronDown, Folder, FolderOpen,
  Sparkles, X, Send, AlertTriangle, Files, Search, GitBranch,
} from 'lucide-react'
import toast from 'react-hot-toast'

// Lazy-load Monaco to avoid SSR issues
const MonacoEditor = dynamic(() => import('@monaco-editor/react'), { ssr: false })

// ─── Types ────────────────────────────────────────────────────────────────────
interface FileEntry {
  name: string
  path: string
  is_dir: boolean
  size?: number
  children?: FileEntry[]
}

interface OpenTab {
  entry: FileEntry
  content: string
  originalContent: string
}

// ─── Language detection ────────────────────────────────────────────────────────
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

// ─── File icon (SVG with per-extension colour) ────────────────────────────────
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

function FileIcon({ name, size = 13 }: { name: string; size?: number }) {
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

// ─── File Tree Node ────────────────────────────────────────────────────────────
function TreeNode({
  entry,
  depth,
  activePath,
  onSelect,
}: {
  entry: FileEntry
  depth: number
  activePath: string | null
  onSelect: (e: FileEntry) => void
}) {
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
        'flex items-center gap-1.5 w-full text-left cursor-default transition-colors',
        isActive ? 'bg-[#37373d]' : 'hover:bg-[rgba(255,255,255,0.05)]',
      )}
      style={{ paddingLeft: `${18 + depth * 12}px`, paddingTop: 3, paddingBottom: 3 }}
    >
      <FileIcon name={entry.name} />
      <span className={cn('text-[12px] truncate ml-0.5', isActive ? 'text-white' : 'text-[#cccccc]')}>
        {entry.name}
      </span>
    </button>
  )
}

// ─── AI Copilot Panel ──────────────────────────────────────────────────────────
interface CopilotMessage { role: 'user' | 'ai'; text: string }

function CopilotPanel({
  projectId,
  selectedPath,
  getCode,
}: {
  projectId: string
  selectedPath: string | null
  getCode: () => string
}) {
  const [open, setOpen] = useState(true)
  const [messages, setMessages] = useState<CopilotMessage[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  const send = useCallback(async () => {
    const msg = input.trim()
    if (!msg) return
    setInput('')
    const code = getCode()
    const context = selectedPath
      ? `\n\nFile: ${selectedPath}\n\`\`\`\n${code.slice(0, 3000)}\n\`\`\``
      : ''
    setMessages((m) => [...m, { role: 'user', text: msg }])
    setLoading(true)
    try {
      const res = await aiApi.chat(`${msg}${context}`)
      setMessages((m) => [...m, { role: 'ai', text: res.data?.reply ?? '...' }])
      setTimeout(() => bottomRef.current?.scrollIntoView({ behavior: 'smooth' }), 50)
    } catch {
      setMessages((m) => [...m, { role: 'ai', text: '⚠ AI request failed. Check AI settings.' }])
    } finally {
      setLoading(false)
    }
  }, [input, selectedPath, getCode])

  return (
    <div
      className="flex flex-col"
      style={{
        width: open ? 300 : 42,
        flexShrink: 0,
        background: '#252526',
        borderLeft: '1px solid #3e3e42',
        transition: 'width 0.2s ease',
      }}
    >
      <div
        className="flex items-center justify-between px-3 cursor-pointer select-none"
        style={{ height: 35, borderBottom: '1px solid #3e3e42', minWidth: 0 }}
        onClick={() => setOpen((v) => !v)}
      >
        {open ? (
          <>
            <span className="flex items-center gap-1.5 text-[11px] font-semibold text-[#cccccc] uppercase tracking-wide">
              <Sparkles size={12} className="text-brand-400" /> Copilot
            </span>
            <X size={13} className="text-[#6e6e6e] hover:text-[#cccccc]" />
          </>
        ) : (
          <Sparkles size={16} className="text-brand-400 mx-auto" />
        )}
      </div>

      {open && (
        <>
          <div className="flex-1 overflow-y-auto p-3 space-y-3">
            {messages.length === 0 && (
              <p className="text-[#6e6e6e] text-[12px] text-center mt-10">
                Ask anything about your code.
                <br />
                <span className="text-[#4e4e4e]">The open file is included as context.</span>
              </p>
            )}
            {messages.map((m, i) => (
              <div
                key={i}
                className={cn('rounded px-3 py-2 text-[12px]', m.role === 'user' ? 'ml-4' : 'mr-4')}
                style={{ background: m.role === 'user' ? 'rgba(99,102,241,0.15)' : '#2d2d30', color: '#cccccc' }}
              >
                <pre className="whitespace-pre-wrap font-sans leading-relaxed">{m.text}</pre>
              </div>
            ))}
            {loading && (
              <div className="flex items-center gap-2 text-[#6e6e6e] text-[11px]">
                <Loader2 size={11} className="animate-spin" /> Thinking…
              </div>
            )}
            <div ref={bottomRef} />
          </div>

          <div className="flex gap-1.5 p-2" style={{ borderTop: '1px solid #3e3e42' }}>
            <input
              className="flex-1 rounded px-2 py-1 text-[12px] text-[#cccccc] placeholder-[#6e6e6e] outline-none focus:border-brand-500"
              style={{ background: '#3c3c3c', border: '1px solid #555' }}
              placeholder="Ask about this code…"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && send()}
            />
            <button
              onClick={send}
              disabled={loading || !input.trim()}
              className="shrink-0 rounded px-2 py-1 text-[12px] bg-brand-600 hover:bg-brand-500 disabled:opacity-40 text-white transition-colors"
            >
              <Send size={11} />
            </button>
          </div>
        </>
      )}
    </div>
  )
}

// ─── Activity bar icon ────────────────────────────────────────────────────────
function ActivityIcon({
  icon: Icon,
  active,
  title,
  onClick,
}: {
  icon: React.ElementType
  active?: boolean
  title: string
  onClick?: () => void
}) {
  return (
    <button
      title={title}
      onClick={onClick}
      className="flex items-center justify-center w-full transition-colors"
      style={{
        height: 48,
        color: active ? '#ffffff' : '#858585',
        borderLeft: active ? '2px solid #ffffff' : '2px solid transparent',
      }}
    >
      <Icon size={24} />
    </button>
  )
}

// ─── Breadcrumbs ──────────────────────────────────────────────────────────────
function Breadcrumbs({ path }: { path: string }) {
  const parts = path.split('/').filter(Boolean)
  return (
    <div
      className="flex items-center gap-0.5 px-3 text-[12px] select-none shrink-0 overflow-x-auto"
      style={{ height: 28, borderBottom: '1px solid #3e3e42', background: '#1e1e1e', color: '#9d9d9d' }}
    >
      {parts.map((part, i) => (
        <span key={i} className="flex items-center gap-0.5 shrink-0">
          {i > 0 && <ChevronRight size={11} className="text-[#555]" />}
          <span style={{ color: i === parts.length - 1 ? '#cccccc' : '#9d9d9d' }}>{part}</span>
        </span>
      ))}
    </div>
  )
}

// ─── Main Editor Page ──────────────────────────────────────────────────────────
export default function EditorPage() {
  const params = useParams<{ id: string }>()
  const projectId = params?.id ?? ''

  const [tabs, setTabs] = useState<OpenTab[]>([])
  const [activePath, setActivePath] = useState<string | null>(null)
  const activeTab = tabs.find((t) => t.entry.path === activePath) ?? null

  const [loadingFile, setLoadingFile] = useState(false)
  const [saving, setSaving] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [cursorPos, setCursorPos] = useState({ line: 1, col: 1 })

  const editorRef = useRef<{ getValue: () => string } | null>(null)
  const saveRef = useRef<() => void>(() => {})

  const { data: filesData, isLoading: treeLoading, error: treeError } = useQuery({
    queryKey: ['project-files', projectId],
    queryFn: () => filesApi.list(projectId).then((r) => r.data),
    retry: false,
  })
  const files: FileEntry[] = filesData?.files ?? []

  const openFile = useCallback(async (entry: FileEntry) => {
    if (entry.is_dir) return
    const existing = tabs.find((t) => t.entry.path === entry.path)
    if (existing) { setActivePath(entry.path); return }
    setLoadingFile(true)
    try {
      const res = await filesApi.read(projectId, entry.path)
      setTabs((prev) => [...prev, { entry, content: res.data.content, originalContent: res.data.content }])
      setActivePath(entry.path)
    } catch {
      toast.error(`Cannot open ${entry.name}`)
    } finally {
      setLoadingFile(false)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tabs, projectId])

  const closeTab = useCallback((path: string, e?: React.MouseEvent) => {
    e?.stopPropagation()
    setTabs((prev) => {
      const next = prev.filter((t) => t.entry.path !== path)
      if (activePath === path) {
        const idx = prev.findIndex((t) => t.entry.path === path)
        setActivePath(next[idx]?.entry.path ?? next[idx - 1]?.entry.path ?? null)
      }
      return next
    })
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activePath])

  const handleChange = useCallback((value: string | undefined) => {
    setTabs((prev) =>
      prev.map((t) => (t.entry.path === activePath ? { ...t, content: value ?? '' } : t))
    )
  }, [activePath])

  const handleSave = useCallback(async () => {
    if (!activeTab) return
    const current = editorRef.current?.getValue() ?? activeTab.content
    setSaving(true)
    try {
      await filesApi.save(projectId, activeTab.entry.path, current)
      setTabs((prev) =>
        prev.map((t) => t.entry.path === activeTab.entry.path ? { ...t, originalContent: current } : t)
      )
      toast.success('Saved')
    } catch {
      toast.error('Save failed')
    } finally {
      setSaving(false)
    }
  }, [activeTab, projectId])

  useEffect(() => { saveRef.current = handleSave }, [handleSave])

  const isDirty = (tab: OpenTab) => tab.content !== tab.originalContent

  const getCode = useCallback(
    () => editorRef.current?.getValue() ?? activeTab?.content ?? '',
    [activeTab]
  )

  return (
    <div
      className="flex flex-col"
      style={{ height: '100vh', background: '#1e1e1e', fontFamily: '"Segoe UI", system-ui, sans-serif' }}
    >
      {/* Title bar */}
      <div
        className="flex items-center justify-between px-3 shrink-0"
        style={{ height: 30, background: '#3c3c3c', borderBottom: '1px solid #252526', userSelect: 'none' }}
      >
        <div className="flex items-center gap-2 text-[12px] text-[#cccccc]">
          <span className="font-semibold text-white">Pushpaka</span>
          <ChevronRight size={10} className="text-[#6e6e6e]" />
          <span>Editor</span>
          {activeTab && (
            <>
              <ChevronRight size={10} className="text-[#6e6e6e]" />
              <span>{activeTab.entry.name}</span>
              {isDirty(activeTab) && <span style={{ color: '#e2c08d' }}>●</span>}
            </>
          )}
        </div>
        <div className="flex items-center gap-2">
          {activeTab && (
            <button
              onClick={handleSave}
              disabled={saving || !isDirty(activeTab)}
              className="flex items-center gap-1 rounded px-2 py-0.5 text-[11px] transition-colors"
              style={{
                background: isDirty(activeTab) ? '#0e639c' : 'transparent',
                color: isDirty(activeTab) ? '#ffffff' : '#6e6e6e',
              }}
            >
              {saving ? <Loader2 size={11} className="animate-spin" /> : <Save size={11} />}
              {saving ? 'Saving…' : 'Save'}
            </button>
          )}
          <button
            onClick={() => window.history.back()}
            className="text-[11px] px-2 py-0.5 rounded transition-colors"
            style={{ color: '#858585' }}
          >
            ← Back
          </button>
        </div>
      </div>

      {/* Main area */}
      <div className="flex flex-1 overflow-hidden">

        {/* Activity bar */}
        <div
          className="flex flex-col items-center shrink-0"
          style={{ width: 48, background: '#333333', borderRight: '1px solid #252526' }}
        >
          <ActivityIcon icon={Files} active={sidebarOpen} title="Explorer" onClick={() => setSidebarOpen((v) => !v)} />
          <ActivityIcon icon={Search} title="Search (not available)" />
          <ActivityIcon icon={GitBranch} title="Source Control (not available)" />
        </div>

        {/* Explorer sidebar */}
        {sidebarOpen && (
          <div
            className="flex flex-col shrink-0 overflow-hidden"
            style={{ width: 220, background: '#252526', borderRight: '1px solid #3e3e42' }}
          >
            <div
              className="flex items-center px-3 shrink-0 uppercase text-[11px] font-semibold tracking-widest"
              style={{ height: 35, borderBottom: '1px solid #3e3e42', color: '#bbbbbe' }}
            >
              Explorer
            </div>
            <div className="flex-1 overflow-y-auto py-1">
              {treeLoading && (
                <div className="flex justify-center mt-10">
                  <Loader2 size={16} className="animate-spin text-[#6e6e6e]" />
                </div>
              )}
              {treeError && (
                <div className="px-3 py-4 text-[11px] flex items-start gap-2" style={{ color: '#f48771' }}>
                  <AlertTriangle size={12} className="shrink-0 mt-0.5" />
                  <span>Trigger a deployment first to load project files.</span>
                </div>
              )}
              {!treeLoading && !treeError && files.length === 0 && (
                <p className="px-4 py-4 text-[11px] text-[#6e6e6e]">No files yet. Trigger a deployment first.</p>
              )}
              {files.map((entry) => (
                <TreeNode
                  key={entry.path}
                  entry={entry}
                  depth={0}
                  activePath={activePath}
                  onSelect={openFile}
                />
              ))}
            </div>
          </div>
        )}

        {/* Editor column */}
        <div className="flex-1 flex flex-col overflow-hidden" style={{ background: '#1e1e1e' }}>

          {/* Tab bar */}
          <div
            className="flex items-end shrink-0 overflow-x-auto"
            style={{ background: '#2d2d2d', borderBottom: '1px solid #252526', height: 35 }}
          >
            {tabs.length === 0 && (
              <div className="flex items-center px-4 h-full text-[12px] text-[#6e6e6e]">
                Open a file from Explorer
              </div>
            )}
            {tabs.map((tab) => {
              const isActive = tab.entry.path === activePath
              const dirty = isDirty(tab)
              return (
                <div
                  key={tab.entry.path}
                  onClick={() => setActivePath(tab.entry.path)}
                  className="flex items-center gap-1.5 px-3 h-full cursor-default select-none group shrink-0"
                  style={{
                    borderRight: '1px solid #252526',
                    borderTop: isActive ? '1px solid #0e639c' : '1px solid transparent',
                    background: isActive ? '#1e1e1e' : '#2d2d2d',
                    color: isActive ? '#ffffff' : '#9d9d9d',
                    maxWidth: 180,
                  }}
                >
                  <FileIcon name={tab.entry.name} size={12} />
                  <span className="text-[12px] truncate">{tab.entry.name}</span>
                  {dirty && <span className="text-[10px] group-hover:hidden" style={{ color: '#e2c08d' }}>●</span>}
                  <button
                    onClick={(e) => closeTab(tab.entry.path, e)}
                    className={cn(
                      'rounded p-0.5 transition-colors',
                      dirty ? 'hidden group-hover:flex items-center' : 'opacity-0 group-hover:opacity-100',
                    )}
                    style={{ color: '#858585' }}
                  >
                    <X size={12} />
                  </button>
                </div>
              )
            })}
          </div>

          {/* Breadcrumbs */}
          {activeTab && <Breadcrumbs path={activeTab.entry.path} />}

          {/* Editor */}
          <div className="flex-1 overflow-hidden relative">
            {loadingFile && (
              <div className="absolute inset-0 flex items-center justify-center z-10" style={{ background: 'rgba(0,0,0,0.4)' }}>
                <Loader2 size={24} className="animate-spin" style={{ color: '#0e639c' }} />
              </div>
            )}
            {!activeTab && !loadingFile && (
              <div className="h-full flex flex-col items-center justify-center select-none">
                <div className="text-center space-y-3">
                  <div className="text-[72px] opacity-[0.06]">⌨</div>
                  <p className="text-[14px]" style={{ color: '#858585' }}>Select a file to start editing</p>
                  <p className="text-[11px] text-[#4e4e4e]">Ctrl+S to save · Files load from the latest deployment&apos;s directory</p>
                </div>
              </div>
            )}
            {activeTab && (
              <MonacoEditor
                height="100%"
                language={detectLanguage(activeTab.entry.name)}
                value={activeTab.content}
                onChange={handleChange}
                onMount={(editor, monaco) => {
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  editorRef.current = editor as any
                  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => saveRef.current())
                  editor.onDidChangeCursorPosition((e) => {
                    setCursorPos({ line: e.position.lineNumber, col: e.position.column })
                  })
                }}
                theme="vs-dark"
                options={{
                  fontSize: 13,
                  fontFamily: '"Cascadia Code", "JetBrains Mono", "Fira Code", Consolas, monospace',
                  fontLigatures: true,
                  minimap: { enabled: true, scale: 1 },
                  scrollBeyondLastLine: false,
                  wordWrap: 'off',
                  tabSize: 2,
                  automaticLayout: true,
                  padding: { top: 8, bottom: 8 },
                  lineNumbers: 'on',
                  renderLineHighlight: 'line',
                  bracketPairColorization: { enabled: true },
                  smoothScrolling: true,
                  cursorBlinking: 'smooth',
                  cursorSmoothCaretAnimation: 'on',
                  renderWhitespace: 'selection',
                  selectionHighlight: true,
                  guides: { bracketPairs: true, indentation: true },
                }}
              />
            )}
          </div>
        </div>

        {/* Copilot */}
        <CopilotPanel
          projectId={projectId}
          selectedPath={activeTab?.entry.path ?? null}
          getCode={getCode}
        />
      </div>

      {/* Status bar */}
      <div
        className="flex items-center justify-between px-3 shrink-0 select-none"
        style={{ height: 22, background: '#007acc', color: '#ffffff', fontSize: 11 }}
      >
        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1">
            <GitBranch size={11} /> main
          </span>
          {activeTab && <span className="opacity-75 truncate max-w-[300px]">{activeTab.entry.path}</span>}
        </div>
        <div className="flex items-center gap-3">
          {activeTab && (
            <>
              <span>Ln {cursorPos.line}, Col {cursorPos.col}</span>
              <span>Spaces: 2</span>
              <span>UTF-8</span>
              <span className="capitalize">{detectLanguage(activeTab.entry.name)}</span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
