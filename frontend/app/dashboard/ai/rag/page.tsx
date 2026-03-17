'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ragApi } from '@/lib/api'
import { PageHeader } from '@/components/ui/PageHeader'
import { Database, Plus, Trash2, Loader2, FileText, Search } from 'lucide-react'

interface RAGDoc {
  id: string
  title: string
  content: string
  created_at: string
}

export default function RAGPage() {
  const qc = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [search, setSearch] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['rag-docs'],
    queryFn: () => ragApi.list(),
  })

  const { mutate: create, isPending: creating } = useMutation({
    mutationFn: () => ragApi.create({ title, content }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['rag-docs'] })
      setTitle('')
      setContent('')
      setShowCreate(false)
    },
  })

  const { mutate: deleteDoc, isPending: deleting } = useMutation({
    mutationFn: (id: string) => ragApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rag-docs'] }),
  })

  const docs: RAGDoc[] = (data?.data ?? []).filter((d: RAGDoc) =>
    search === '' ||
    d.title.toLowerCase().includes(search.toLowerCase()) ||
    d.content.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="flex flex-col min-h-screen">
      <PageHeader
        title="Knowledge Base"
        description="RAG documents for AI context and support agent"
        icon={<Database className="text-cyan-400" size={22} />}
        actions={
          <button
            onClick={() => setShowCreate((v) => !v)}
            className="flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-semibold text-white transition-all"
            style={{
              background: 'linear-gradient(135deg,#4338ca,#6366f1)',
              boxShadow: '0 4px 14px rgba(99,102,241,0.4)',
            }}
          >
            <Plus size={15} />
            Add Document
          </button>
        }
      />

      <div className="flex-1 p-4 md:p-6 space-y-4">
        {/* Create form */}
        {showCreate && (
          <div
            className="rounded-xl p-5 space-y-3"
            style={{ background: 'rgba(99,102,241,0.06)', border: '1px solid rgba(99,102,241,0.2)' }}
          >
            <h3 className="text-sm font-semibold text-slate-200">New Knowledge Document</h3>
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Document title (e.g. Deployment Troubleshooting)"
              className="w-full rounded-lg px-3 py-2 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:ring-2 focus:ring-brand-500/50"
              style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.08)' }}
            />
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={6}
              placeholder="Paste or type the knowledge content that the AI should use as context..."
              className="w-full rounded-lg px-3 py-2 text-sm text-slate-300 placeholder-slate-600 resize-y focus:outline-none focus:ring-2 focus:ring-brand-500/50"
              style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.08)', minHeight: '120px' }}
            />
            <div className="flex items-center gap-2">
              <button
                onClick={() => create()}
                disabled={creating || !title.trim() || !content.trim()}
                className="flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-semibold text-white transition-all disabled:opacity-50"
                style={{ background: 'linear-gradient(135deg,#4338ca,#6366f1)', boxShadow: '0 4px 14px rgba(99,102,241,0.4)' }}
              >
                {creating ? <Loader2 size={14} className="animate-spin" /> : <Plus size={14} />}
                Save Document
              </button>
              <button
                onClick={() => { setShowCreate(false); setTitle(''); setContent('') }}
                className="px-4 py-2 rounded-xl text-sm font-medium text-slate-400 hover:text-slate-200 transition-colors"
                style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.08)' }}
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* Search */}
        <div
          className="flex items-center gap-2 px-3 rounded-xl"
          style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
        >
          <Search size={14} className="text-slate-600 shrink-0" />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search knowledge base…"
            className="flex-1 bg-transparent py-2.5 text-sm text-slate-300 placeholder-slate-600 focus:outline-none"
          />
        </div>

        {/* Document list */}
        {isLoading ? (
          <div className="flex justify-center py-16">
            <Loader2 size={20} className="animate-spin text-brand-400" />
          </div>
        ) : docs.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Database size={36} className="text-slate-700 mb-3" />
            <p className="text-sm font-medium text-slate-400">
              {search ? 'No matching documents' : 'No documents yet'}
            </p>
            <p className="text-xs text-slate-600 mt-1">
              {search
                ? 'Try a different search term'
                : 'Add knowledge documents to improve AI context quality'}
            </p>
          </div>
        ) : (
          <div className="grid gap-3 md:grid-cols-2">
            {docs.map((doc) => (
              <div
                key={doc.id}
                className="rounded-xl p-4 group"
                style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
              >
                <div className="flex items-start gap-3">
                  <div
                    className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0"
                    style={{ background: 'rgba(6,182,212,0.12)', border: '1px solid rgba(6,182,212,0.2)' }}
                  >
                    <FileText size={16} className="text-cyan-400" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-semibold text-slate-200 truncate">{doc.title}</p>
                    <p className="text-xs text-slate-500 mt-0.5 line-clamp-2">{doc.content}</p>
                    <p className="text-[11px] text-slate-700 mt-1.5">
                      Added {new Date(doc.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <button
                    onClick={() => deleteDoc(doc.id)}
                    disabled={deleting}
                    className="shrink-0 p-1.5 rounded-lg text-slate-700 hover:text-red-400 hover:bg-red-500/10 transition-colors opacity-0 group-hover:opacity-100"
                    title="Delete document"
                  >
                    {deleting ? <Loader2 size={14} className="animate-spin" /> : <Trash2 size={14} />}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
