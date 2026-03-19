'use client'

import { useQuery } from '@tanstack/react-query'
import { systemFilesApi } from '@/lib/api'
import { Finder } from '@/components/files/Finder'
import { useConfirm, usePrompt } from '@/components/ui/Modal'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'

export default function FileManagerPage() {
  const router = useRouter()
  const { data: filesData, isLoading, refetch } = useQuery({
    queryKey: ['system-files'],
    queryFn: () => systemFilesApi.list().then((r) => r.data),
  })

  const { confirm, Component: ConfirmModal } = useConfirm()
  const { prompt, Component: PromptModal } = usePrompt()

  const handleOpen = (entry: any) => {
    if (entry.is_dir) return
    
    // REDIRECTION OPTIMIZATION: If this is a project file, go straight to Project Editor
    const projectMatch = entry.path.match(/^\/deploy\/pushpaka\/([^/]+)(\/.*)?$/)
    if (projectMatch) {
      const projectId = projectMatch[1]
      const relativePath = projectMatch[2] || '/'
      router.push(`/dashboard/projects/${projectId}/editor?file=${encodeURIComponent(relativePath)}`)
      return
    }

    // Otherwise use Global Editor
    router.push(`/dashboard/editor?file=${encodeURIComponent(entry.path)}`)
  }

  const handleDelete = async (path: string) => {
    const ok = await confirm({
      title: 'Delete File',
      message: `Are you sure you want to permanently delete ${path}? This action cannot be undone.`,
      confirmText: 'Delete',
      type: 'error'
    })
    if (!ok) return

    try {
      await systemFilesApi.delete(path)
      toast.success('File deleted')
      refetch()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Delete failed')
    }
  }

  const handleCreateFile = async () => {
    const name = await prompt({
      title: 'New File',
      message: 'Enter the full path for the new file:',
      placeholder: '/path/to/file.txt',
      confirmText: 'Create'
    })
    if (!name) return

    try {
      await systemFilesApi.createFile(name)
      toast.success('File created')
      refetch()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Failed to create file')
    }
  }

  const handleCreateFolder = async () => {
    const name = await prompt({
      title: 'New Folder',
      message: 'Enter the path for the new folder:',
      placeholder: '/path/to/folder',
      confirmText: 'Create'
    })
    if (!name) return

    try {
      await systemFilesApi.createDirectory(name)
      toast.success('Folder created')
      refetch()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Failed to create folder')
    }
  }

  return (
    <div className="h-screen flex flex-col">
      <div className="flex-1 overflow-hidden">
        <Finder
          files={filesData?.files || []}
          isLoading={isLoading}
          onOpen={handleOpen}
          onDelete={handleDelete}
          onCreateFile={handleCreateFile}
          onCreateFolder={handleCreateFolder}
          onRefresh={refetch}
        />
      </div>
      {ConfirmModal}
      {PromptModal}
    </div>
  )
}
