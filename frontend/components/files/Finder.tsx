'use client'

import { useState, useMemo, useCallback } from 'react'
import {
  ChevronRight, ChevronLeft, Search, Grid, List,
  File, Folder, MoreHorizontal, Info, Download,
  Trash2, Edit3, Share2, CornerUpLeft, Plus,
  FileText, FileCode, FileImage, FileAudio, FileVideo,
  Archive, Database, Terminal, Globe, HardDrive
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { format } from 'date-fns'

interface FileEntry {
  name: string
  path: string
  is_dir: boolean
  size?: number
  updated_at?: string
  children?: FileEntry[]
}

interface FinderProps {
  files: FileEntry[]
  onOpen: (entry: FileEntry) => void
  onDelete: (path: string) => void
  onCreateFile: () => void
  onCreateFolder: () => void
  onRefresh: () => void
  isLoading?: boolean
}

export function Finder({
  files,
  onOpen,
  onDelete,
  onCreateFile,
  onCreateFolder,
  onRefresh,
  isLoading
}: FinderProps) {
  const [currentPath, setCurrentPath] = useState<string[]>([])
  const [selectedPath, setSelectedPath] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [searchQuery, setSearchQuery] = useState('')

  // Compute current folder's files
  const currentFiles = useMemo(() => {
    let current = files
    for (const part of currentPath) {
      const found = current.find(f => f.name === part && f.is_dir)
      if (found && found.children) {
        current = found.children
      } else {
        return []
      }
    }
    return current.filter(f => f.name.toLowerCase().includes(searchQuery.toLowerCase()))
  }, [files, currentPath, searchQuery])

  const selectedFile = useMemo(() => {
    return currentFiles.find(f => f.path === selectedPath) || null
  }, [currentFiles, selectedPath])

  const navigateTo = (entry: FileEntry) => {
    if (entry.is_dir) {
      setCurrentPath(prev => [...prev, entry.name])
      setSelectedPath(null)
    } else {
      onOpen(entry)
    }
  }

  const navigateUp = () => {
    if (currentPath.length > 0) {
      setCurrentPath(prev => prev.slice(0, -1))
      setSelectedPath(null)
    }
  }

  const navigateToBreadcrumb = (index: number) => {
    setCurrentPath(prev => prev.slice(0, index + 1))
    setSelectedPath(null)
  }

  const getFileIcon = (entry: FileEntry) => {
    if (entry.is_dir) return <Folder className="text-amber-400 fill-amber-400/20" size={viewMode === 'grid' ? 44 : 18} />
    const ext = entry.name.split('.').pop()?.toLowerCase()
    const color = "text-slate-400"
    
    if (['ts', 'tsx', 'js', 'jsx', 'go', 'py', 'rs', 'c', 'cpp'].includes(ext!)) 
      return <FileCode className="text-blue-400" size={viewMode === 'grid' ? 44 : 18} />
    if (['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp'].includes(ext!)) 
      return <FileImage className="text-purple-400" size={viewMode === 'grid' ? 44 : 18} />
    if (['mp4', 'mov', 'webm'].includes(ext!)) 
      return <FileVideo className="text-pink-400" size={viewMode === 'grid' ? 44 : 18} />
    if (['mp3', 'wav', 'ogg'].includes(ext!)) 
      return <FileAudio className="text-rose-400" size={viewMode === 'grid' ? 44 : 18} />
    if (['zip', 'rar', 'gz', 'tar'].includes(ext!)) 
      return <Archive className="text-amber-600" size={viewMode === 'grid' ? 44 : 18} />
    if (['sql', 'db', 'sqlite'].includes(ext!)) 
      return <Database className="text-emerald-400" size={viewMode === 'grid' ? 44 : 18} />
    if (['md', 'txt', 'pdf', 'doc', 'docx'].includes(ext!)) 
      return <FileText className="text-slate-300" size={viewMode === 'grid' ? 44 : 18} />
    
    return <File className={color} size={viewMode === 'grid' ? 44 : 18} />
  }

  const formatSize = (bytes?: number) => {
    if (bytes === undefined) return '--'
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  return (
    <div className="flex flex-col h-full bg-slate-950 text-slate-200 overflow-hidden font-sans">
      {/* Toolbar */}
      <div className="h-14 flex items-center justify-between px-4 border-b border-slate-800 bg-slate-900/50 backdrop-blur-md z-10">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1">
            <button
              onClick={navigateUp}
              disabled={currentPath.length === 0}
              className="p-1.5 rounded-lg hover:bg-slate-800 disabled:opacity-30 transition-colors"
            >
              <ChevronLeft size={20} />
            </button>
            <button
              disabled
              className="p-1.5 rounded-lg hover:bg-slate-800 disabled:opacity-30 transition-colors"
            >
              <ChevronRight size={20} />
            </button>
          </div>
          
          {/* Breadcrumbs */}
          <div className="flex items-center gap-1 text-sm overflow-x-auto no-scrollbar max-w-md">
            <button 
              onClick={() => setCurrentPath([])}
              className="flex items-center gap-1 px-2 py-1 rounded-md hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
            >
              <HardDrive size={14} />
              <span>Deploy</span>
            </button>
            {currentPath.map((part, i) => (
              <div key={i} className="flex items-center gap-1 shrink-0">
                <ChevronRight size={14} className="text-slate-600" />
                <button
                  onClick={() => navigateToBreadcrumb(i)}
                  className="px-2 py-1 rounded-md hover:bg-slate-800 text-slate-300 hover:text-white transition-colors"
                >
                  {part}
                </button>
              </div>
            ))}
          </div>
        </div>

        <div className="flex items-center gap-4">
          {/* Search */}
          <div className="relative group">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 group-focus-within:text-brand-400 transition-colors" size={16} />
            <input
              className="bg-slate-800/50 border border-slate-700/50 rounded-full pl-9 pr-4 py-1.5 text-sm w-64 focus:outline-none focus:ring-2 focus:ring-brand-500/30 focus:bg-slate-800 transition-all"
              placeholder="Search files..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>

          <div className="h-6 w-px bg-slate-800 mx-1" />

          {/* View Modes */}
          <div className="flex items-center bg-slate-800/50 p-1 rounded-lg border border-slate-700/50">
            <button
              onClick={() => setViewMode('grid')}
              className={cn("p-1.5 rounded-md transition-all", viewMode === 'grid' ? "bg-slate-700 text-white shadow-sm" : "text-slate-500 hover:text-slate-300")}
            >
              <Grid size={18} />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={cn("p-1.5 rounded-md transition-all", viewMode === 'list' ? "bg-slate-700 text-white shadow-sm" : "text-slate-500 hover:text-slate-300")}
            >
              <List size={18} />
            </button>
          </div>
          
          <button
            onClick={onRefresh}
            className="p-2 rounded-lg hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
            title="Refresh"
          >
            <CornerUpLeft size={18} className={isLoading ? "animate-spin" : ""} />
          </button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden relative">
        {/* Main Content Area */}
        <div 
          className="flex-1 overflow-y-auto p-6"
          onClick={() => setSelectedPath(null)}
        >
          {isLoading ? (
            <div className="h-full flex flex-col items-center justify-center gap-4">
              <div className="w-12 h-12 rounded-full border-4 border-slate-800 border-t-brand-500 animate-spin" />
              <p className="text-slate-500 text-sm animate-pulse">Scanning file system...</p>
            </div>
          ) : currentFiles.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center gap-4 opacity-40">
              <Folder size={64} className="text-slate-700" strokeWidth={1} />
              <p className="text-lg font-medium">Empty Folder</p>
            </div>
          ) : viewMode === 'grid' ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-x-4 gap-y-8">
              {currentFiles.map((file) => (
                <div
                  key={file.path}
                  onClick={(e) => { e.stopPropagation(); setSelectedPath(file.path) }}
                  onDoubleClick={() => navigateTo(file)}
                  className={cn(
                    "flex flex-col items-center group cursor-default select-none",
                  )}
                >
                  <div className={cn(
                    "p-4 rounded-xl transition-all duration-200 flex flex-col items-center gap-2",
                    selectedPath === file.path 
                      ? "bg-brand-600/20 ring-1 ring-brand-500/50 shadow-lg shadow-brand-500/10" 
                      : "group-hover:bg-slate-800/50"
                  )}>
                    <div className="transform group-active:scale-95 transition-transform duration-100">
                      {getFileIcon(file)}
                    </div>
                  </div>
                  <span className={cn(
                    "mt-2 text-[13px] px-2 py-0.5 rounded-full text-center truncate w-full break-all",
                    selectedPath === file.path ? "bg-brand-500 text-white shadow-md shadow-brand-900/40" : "text-slate-300"
                  )}>
                    {file.name}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <div className="w-full">
              <table className="w-full text-left text-sm border-separate border-spacing-0">
                <thead>
                  <tr className="text-slate-500 font-medium">
                    <th className="pb-4 pl-4 border-b border-slate-800">Name</th>
                    <th className="pb-4 border-b border-slate-800">Size</th>
                    <th className="pb-4 border-b border-slate-800">Kind</th>
                    <th className="pb-4 border-b border-slate-800">Modified</th>
                  </tr>
                </thead>
                <tbody className="before:block before:h-2">
                  {currentFiles.map((file) => (
                    <tr
                      key={file.path}
                      onClick={(e) => { e.stopPropagation(); setSelectedPath(file.path) }}
                      onDoubleClick={() => navigateTo(file)}
                      className={cn(
                        "group transition-colors",
                        selectedPath === file.path ? "bg-brand-500/20" : "hover:bg-slate-900"
                      )}
                    >
                      <td className="py-2.5 pl-4 rounded-l-lg">
                        <div className="flex items-center gap-3">
                          {getFileIcon(file)}
                          <span className={selectedPath === file.path ? "text-white font-medium" : "text-slate-200"}>{file.name}</span>
                        </div>
                      </td>
                      <td className="py-2.5 text-slate-400">{file.is_dir ? '--' : formatSize(file.size)}</td>
                      <td className="py-2.5 text-slate-400">{file.is_dir ? 'Folder' : (file.name.split('.').pop()?.toUpperCase() || 'File')}</td>
                      <td className="py-2.5 rounded-r-lg text-slate-400">Mar 19, 2026</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* Floating Action Button for New */}
          <div className="absolute bottom-8 right-8 flex flex-col gap-3">
             <button
              onClick={onCreateFolder}
              className="w-12 h-12 bg-slate-800 hover:bg-slate-700 text-white rounded-full flex items-center justify-center shadow-2xl border border-slate-700 hover:scale-110 active:scale-95 transition-all group"
              title="New Folder"
            >
              <FolderPlus size={24} className="text-amber-400" />
            </button>
            <button
              onClick={onCreateFile}
              className="w-14 h-14 bg-brand-600 hover:bg-brand-500 text-white rounded-full flex items-center justify-center shadow-2xl shadow-brand-900/40 hover:scale-110 active:scale-95 transition-all group"
              title="New File"
            >
              <Plus size={32} />
            </button>
          </div>
        </div>

        {/* Sidebar (Preview) */}
        <div className={cn(
          "w-80 border-l border-slate-800 bg-slate-900/30 backdrop-blur-sm transition-all duration-300",
          selectedFile ? "translate-x-0 opacity-100" : "translate-x-full opacity-0 absolute right-0"
        )}>
          {selectedFile && (
            <div className="p-6 h-full flex flex-col overflow-y-auto">
              <div className="flex flex-col items-center text-center gap-6 py-8">
                <div className="p-10 bg-slate-800/40 rounded-3xl shadow-inner ring-1 ring-slate-700/50">
                  {getFileIcon(selectedFile)}
                </div>
                <div>
                  <h2 className="text-lg font-bold text-white break-all px-2">{selectedFile.name}</h2>
                  <p className="text-slate-500 text-sm mt-1 uppercase tracking-widest font-semibold">
                    {selectedFile.is_dir ? 'Folder' : (selectedFile.name.split('.').pop() || 'File')}
                  </p>
                </div>
              </div>

              <div className="space-y-6 mt-4">
                <div className="p-4 bg-slate-800/30 rounded-2xl border border-slate-700/30 space-y-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">Kind</span>
                    <span className="text-slate-300 font-medium">{selectedFile.is_dir ? 'Folder' : 'Plain Text'}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">Size</span>
                    <span className="text-slate-300 font-medium">{formatSize(selectedFile.size)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">Where</span>
                    <span className="text-slate-300 font-medium truncate max-w-[140px]" title={selectedFile.path}>{selectedFile.path}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">Modified</span>
                    <span className="text-slate-300 font-medium">Mar 19, 2026, 22:53</span>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <button
                    onClick={() => onOpen(selectedFile)}
                    className="flex flex-col items-center justify-center gap-2 p-4 bg-slate-800 hover:bg-slate-700 rounded-2xl border border-slate-700 transition-all text-sm group"
                  >
                    <Edit3 size={20} className="text-brand-400 group-hover:scale-110 transition-transform" />
                    <span>Open</span>
                  </button>
                  <button
                    onClick={() => onDelete(selectedFile.path)}
                    className="flex flex-col items-center justify-center gap-2 p-4 bg-slate-800 hover:bg-red-900/20 rounded-2xl border border-slate-700 hover:border-red-900/50 transition-all text-sm group"
                  >
                    <Trash2 size={20} className="text-red-500 group-hover:scale-110 transition-transform" />
                    <span>Delete</span>
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function FolderPlus(props: any) {
  return (
    <svg {...props} width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2z" />
      <line x1="12" y1="10" x2="12" y2="16" />
      <line x1="9" y1="13" x2="15" y2="13" />
    </svg>
  )
}
