'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { projectsApi } from '@/lib/api'
import { Project } from '@/types'
import { Header } from '@/components/layout/Header'
import { ProjectCard } from '@/components/dashboard/ProjectCard'
import { Plus, FolderGit2, Search } from 'lucide-react'

export default function ProjectsPage() {
  const [search, setSearch] = useState('')
  const router = useRouter()

  const { data, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectsApi.list().then((r) => r.data),
  })

  const projects: Project[] = data?.data || []
  const filtered = projects.filter((p) =>
    p.name.toLowerCase().includes(search.toLowerCase()) ||
    p.repo_url.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="flex flex-col min-h-screen">
      <Header title="Projects" subtitle={`${projects.length} project${projects.length !== 1 ? 's' : ''}`} />

      <div className="p-6 space-y-5">
        {/* Toolbar */}
        <div className="flex items-center gap-3">
          <div className="relative flex-1 max-w-sm">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
            <input
              type="text"
              placeholder="Search projects..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="input pl-8"
            />
          </div>
          <Link href="/dashboard/projects/new" className="btn-primary ml-auto">
            <Plus size={15} />
            New Project
          </Link>
        </div>

        {/* Project grid */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="card animate-pulse">
                <div className="h-4 bg-slate-700 rounded w-2/3 mb-3" />
                <div className="h-3 bg-slate-800 rounded w-full mb-4" />
                <div className="h-8 bg-slate-800 rounded" />
              </div>
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <div className="card text-center py-16">
            <FolderGit2 size={48} className="mx-auto text-slate-700 mb-4" />
            {search ? (
              <p className="text-slate-400">No projects matching &ldquo;{search}&rdquo;</p>
            ) : (
              <>
                <h3 className="text-white font-semibold mb-2">No projects yet</h3>
                <p className="text-slate-400 text-sm mb-5">
                  Connect a Git repository to get started
                </p>
                <Link href="/dashboard/projects/new" className="btn-primary inline-flex">
                  <Plus size={15} />
                  Create your first project
                </Link>
              </>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filtered.map((project) => (
              <ProjectCard key={project.id} project={project} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
