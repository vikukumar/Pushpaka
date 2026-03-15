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
    <span className={cn(c.class)}>
      <span className={cn('w-1.5 h-1.5 rounded-full', c.dot)} />
      {c.label}
    </span>
  )
}
