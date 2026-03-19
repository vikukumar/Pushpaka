import { DeploymentStatus } from '@/types'
import { cn } from '@/lib/utils'

const config: Record<DeploymentStatus, { label: string; dot: string; class: string }> = {
  running:  { label: 'Running',  dot: 'bg-emerald-400', class: 'badge-running' },
  building: { label: 'Building', dot: 'bg-amber-400 animate-pulse', class: 'badge-building' },
  queued:   { label: 'Queued',   dot: 'bg-blue-400',    class: 'badge-queued' },
  failed:   { label: 'Failed',   dot: 'bg-red-400',     class: 'badge-failed' },
  stopped:  { label: 'Stopped',  dot: 'bg-slate-400',   class: 'badge-stopped' },
}

export function StatusBadge({ status }: { status: DeploymentStatus }) {
  const c = config[status] ?? config.stopped
  return (
    <span className={cn('px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider flex items-center gap-1.5 shadow-sm', c.class)}>
      <span className={cn('w-1.5 h-1.5 rounded-full ring-2 ring-opacity-20', c.dot, 
        status === 'running' ? 'ring-emerald-400' : 
        status === 'building' ? 'ring-amber-400' : 
        status === 'failed' ? 'ring-red-400' : 
        status === 'queued' ? 'ring-blue-400' : 'ring-slate-400'
      )} />
      {c.label}
    </span>
  )
}
