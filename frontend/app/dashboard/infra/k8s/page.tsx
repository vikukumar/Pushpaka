'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { k8sApi } from '@/lib/api'
import { PageHeader } from '@/components/ui/PageHeader'
import {
  Server,
  RefreshCw,
  Loader2,
  CheckCircle2,
  XCircle,
  Clock,
  Layers,
  GitBranch,
  Network,
  RotateCcw,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface K8sPod {
  name: string
  namespace: string
  status: string
  ready: string
  restarts: number
  age: string
  node?: string
}

interface K8sDeployment {
  name: string
  namespace: string
  ready: string
  up_to_date: number
  available: number
  age: string
}

interface K8sService {
  name: string
  namespace: string
  type: string
  cluster_ip: string
  external_ip?: string
  ports: string
  age: string
}

function PodStatusBadge({ status }: { status: string }) {
  const s = status.toLowerCase()
  if (s === 'running')
    return (
      <span className="flex items-center gap-1 text-xs text-emerald-400">
        <CheckCircle2 size={11} /> Running
      </span>
    )
  if (s === 'pending')
    return (
      <span className="flex items-center gap-1 text-xs text-amber-400">
        <Clock size={11} /> Pending
      </span>
    )
  return (
    <span className="flex items-center gap-1 text-xs text-red-400">
      <XCircle size={11} /> {status}
    </span>
  )
}

const TABS = ['Pods', 'Deployments', 'Services'] as const
type Tab = (typeof TABS)[number]

export default function K8sPage() {
  const [namespace, setNamespace] = useState('default')
  const [activeTab, setActiveTab] = useState<Tab>('Pods')
  const qc = useQueryClient()

  const { data: nsData } = useQuery({
    queryKey: ['k8s-namespaces'],
    queryFn: () => k8sApi.namespaces(),
  })

  const { data: podsData, isLoading: loadingPods, refetch: refetchPods } = useQuery({
    queryKey: ['k8s-pods', namespace],
    queryFn: () => k8sApi.pods(namespace),
    enabled: activeTab === 'Pods',
    refetchInterval: 20_000,
  })

  const { data: deployData, isLoading: loadingDeploy, refetch: refetchDeploy } = useQuery({
    queryKey: ['k8s-deployments', namespace],
    queryFn: () => k8sApi.deployments(namespace),
    enabled: activeTab === 'Deployments',
    refetchInterval: 20_000,
  })

  const { data: svcData, isLoading: loadingSvc, refetch: refetchSvc } = useQuery({
    queryKey: ['k8s-services', namespace],
    queryFn: () => k8sApi.services(namespace),
    enabled: activeTab === 'Services',
    refetchInterval: 30_000,
  })

  const { mutate: rollout, isPending: rollingOut } = useMutation({
    mutationFn: ({ ns, name }: { ns: string; name: string }) => k8sApi.rollout(ns, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['k8s-deployments'] }),
  })

  const namespaces: string[] = nsData?.data ?? ['default']
  const pods: K8sPod[] = podsData?.data ?? []
  const deployments: K8sDeployment[] = deployData?.data ?? []
  const services: K8sService[] = svcData?.data ?? []

  const isLoading =
    (activeTab === 'Pods' && loadingPods) ||
    (activeTab === 'Deployments' && loadingDeploy) ||
    (activeTab === 'Services' && loadingSvc)

  const refetch =
    activeTab === 'Pods'
      ? refetchPods
      : activeTab === 'Deployments'
      ? refetchDeploy
      : refetchSvc

  return (
    <div className="flex flex-col min-h-screen">
      <PageHeader
        title="Kubernetes"
        description="Workload overview — pods, deployments, services"
        icon={<Server className="text-indigo-400" size={22} />}
        actions={
          <div className="flex items-center gap-2">
            <select
              value={namespace}
              onChange={(e) => setNamespace(e.target.value)}
              className="rounded-lg px-3 py-1.5 text-xs font-medium text-slate-300 focus:outline-none focus:ring-2 focus:ring-brand-500/50"
              style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)' }}
            >
              {namespaces.map((ns) => (
                <option key={ns} value={ns}>{ns}</option>
              ))}
            </select>
            <button
              onClick={() => refetch()}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-slate-200 transition-colors"
              style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.08)' }}
            >
              <RefreshCw size={12} />
              Refresh
            </button>
          </div>
        }
      />

      <div className="flex-1 p-4 md:p-6 space-y-4">
        {/* Tabs */}
        <div
          className="flex gap-1 p-1 rounded-xl w-fit"
          style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.07)' }}
        >
          {TABS.map((tab) => {
            const Icon = tab === 'Pods' ? Layers : tab === 'Deployments' ? GitBranch : Network
            return (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={cn(
                  'flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-semibold transition-all',
                  activeTab === tab ? 'text-white' : 'text-slate-500 hover:text-slate-300'
                )}
                style={
                  activeTab === tab
                    ? {
                        background: 'linear-gradient(135deg,rgba(99,102,241,0.25),rgba(99,102,241,0.12))',
                        boxShadow: '0 1px 6px rgba(99,102,241,0.2)',
                      }
                    : undefined
                }
              >
                <Icon size={13} />
                {tab}
              </button>
            )
          })}
        </div>

        {/* Content */}
        {isLoading ? (
          <div className="flex justify-center py-16">
            <Loader2 size={20} className="animate-spin text-brand-400" />
          </div>
        ) : (
          <>
            {/* Pods */}
            {activeTab === 'Pods' && (
              <div className="rounded-xl overflow-hidden" style={{ border: '1px solid rgba(255,255,255,0.07)' }}>
                <table className="w-full text-sm">
                  <thead>
                    <tr style={{ background: 'rgba(255,255,255,0.03)', borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
                      {['Name', 'Ready', 'Status', 'Restarts', 'Age'].map((h) => (
                        <th key={h} className="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">{h}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y" style={{ borderColor: 'rgba(255,255,255,0.04)' }}>
                    {pods.length === 0 ? (
                      <tr><td colSpan={5} className="text-center py-10 text-sm text-slate-600">No pods found in namespace &quot;{namespace}&quot;</td></tr>
                    ) : pods.map((p) => (
                      <tr key={p.name} className="hover:bg-white/[0.015] transition-colors">
                        <td className="px-4 py-3 font-mono text-xs text-slate-300 max-w-xs truncate">{p.name}</td>
                        <td className="px-4 py-3 text-xs text-slate-500 font-mono">{p.ready}</td>
                        <td className="px-4 py-3"><PodStatusBadge status={p.status} /></td>
                        <td className="px-4 py-3 text-xs text-slate-500">{p.restarts}</td>
                        <td className="px-4 py-3 text-xs text-slate-600">{p.age}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {/* Deployments */}
            {activeTab === 'Deployments' && (
              <div className="rounded-xl overflow-hidden" style={{ border: '1px solid rgba(255,255,255,0.07)' }}>
                <table className="w-full text-sm">
                  <thead>
                    <tr style={{ background: 'rgba(255,255,255,0.03)', borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
                      {['Name', 'Ready', 'Up-to-date', 'Available', 'Age', 'Actions'].map((h) => (
                        <th key={h} className="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">{h}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y" style={{ borderColor: 'rgba(255,255,255,0.04)' }}>
                    {deployments.length === 0 ? (
                      <tr><td colSpan={6} className="text-center py-10 text-sm text-slate-600">No deployments found in namespace &quot;{namespace}&quot;</td></tr>
                    ) : deployments.map((d) => (
                      <tr key={d.name} className="hover:bg-white/[0.015] transition-colors">
                        <td className="px-4 py-3 font-mono text-xs text-slate-300">{d.name}</td>
                        <td className="px-4 py-3 text-xs font-mono text-slate-500">{d.ready}</td>
                        <td className="px-4 py-3 text-xs text-slate-500">{d.up_to_date}</td>
                        <td className="px-4 py-3 text-xs text-slate-500">{d.available}</td>
                        <td className="px-4 py-3 text-xs text-slate-600">{d.age}</td>
                        <td className="px-4 py-3">
                          <button
                            onClick={() => rollout({ ns: namespace, name: d.name })}
                            disabled={rollingOut}
                            className="flex items-center gap-1 px-2.5 py-1 rounded-lg text-[11px] font-medium transition-colors"
                            style={{
                              background: 'rgba(99,102,241,0.1)',
                              border: '1px solid rgba(99,102,241,0.25)',
                              color: '#a5b4fc',
                            }}
                          >
                            {rollingOut ? <Loader2 size={11} className="animate-spin" /> : <RotateCcw size={11} />}
                            Rollout
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {/* Services */}
            {activeTab === 'Services' && (
              <div className="rounded-xl overflow-hidden" style={{ border: '1px solid rgba(255,255,255,0.07)' }}>
                <table className="w-full text-sm">
                  <thead>
                    <tr style={{ background: 'rgba(255,255,255,0.03)', borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
                      {['Name', 'Type', 'Cluster IP', 'External IP', 'Ports', 'Age'].map((h) => (
                        <th key={h} className="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">{h}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y" style={{ borderColor: 'rgba(255,255,255,0.04)' }}>
                    {services.length === 0 ? (
                      <tr><td colSpan={6} className="text-center py-10 text-sm text-slate-600">No services found in namespace &quot;{namespace}&quot;</td></tr>
                    ) : services.map((s) => (
                      <tr key={s.name} className="hover:bg-white/[0.015] transition-colors">
                        <td className="px-4 py-3 font-mono text-xs text-slate-300">{s.name}</td>
                        <td className="px-4 py-3 text-xs text-slate-500">
                          <span
                            className="px-2 py-0.5 rounded text-[10px] font-semibold"
                            style={{
                              background: s.type === 'LoadBalancer' ? 'rgba(6,182,212,0.12)' : 'rgba(99,102,241,0.12)',
                              color: s.type === 'LoadBalancer' ? '#22d3ee' : '#818cf8',
                            }}
                          >
                            {s.type}
                          </span>
                        </td>
                        <td className="px-4 py-3 font-mono text-xs text-slate-500">{s.cluster_ip}</td>
                        <td className="px-4 py-3 font-mono text-xs text-slate-500">{s.external_ip || '—'}</td>
                        <td className="px-4 py-3 font-mono text-xs text-slate-600">{s.ports}</td>
                        <td className="px-4 py-3 text-xs text-slate-600">{s.age}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
