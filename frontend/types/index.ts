export interface User {
  id: string
  email: string
  name: string
  role: string
  created_at: string
}

export interface Project {
  id: string
  user_id: string
  name: string
  repo_url: string
  branch: string
  install_command: string
  build_command: string
  start_command: string
  run_dir: string
  port: number
  framework: string
  status: 'active' | 'inactive' | 'building'
  is_private: boolean
  cpu_limit: string
  memory_limit: string
  restart_policy: string
  deploy_target: string
  k8s_namespace: string
  created_at: string
  updated_at: string
}

export interface AuditLog {
  id: string
  user_id: string
  action: string
  resource: string
  resource_id: string
  metadata: string
  ip_addr: string
  user_agent: string
  created_at: string
}

export interface WebhookConfig {
  id: string
  project_id: string
  provider: string
  branch: string
  webhook_url: string
  created_at: string
}

export type DeploymentStatus = 'queued' | 'building' | 'running' | 'failed' | 'stopped'

export interface Deployment {
  id: string
  project_id: string
  user_id: string
  commit_sha: string
  commit_msg: string
  branch: string
  status: DeploymentStatus
  image_tag: string
  container_id: string
  url: string
  error_msg: string
  started_at: string | null
  finished_at: string | null
  created_at: string
  updated_at: string
}

export interface DeploymentLog {
  id: string
  deployment_id: string
  level: 'info' | 'error' | 'debug' | 'warn'
  message: string
  stream: 'stdout' | 'stderr' | 'system'
  created_at: string
}

export interface Domain {
  id: string
  project_id: string
  user_id: string
  domain: string
  verified: boolean
  ssl_enabled: boolean
  created_at: string
}

export interface EnvVar {
  id: string
  project_id: string
  key: string
  has_value: boolean
  created_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiResponse<T> {
  data?: T
  error?: string
}

export interface SystemInfo {
  docker: {
    available: boolean
    host: string
  }
  git: {
    available: boolean
    version: string
  }
  workers: {
    total: number
    active_jobs: number
    idle: number
    queue_mode: 'redis' | 'in-process'
    /** false when workers run as separate Redis-connected processes (untracked by API) */
    tracked: boolean
  }
  runtime: {
    os: string
    arch: string
    in_container: boolean
  }
}
