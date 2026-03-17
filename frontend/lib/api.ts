import axios from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export const apiClient = axios.create({
  baseURL: `${API_URL}/api/v1`,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Attach token to every request
apiClient.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    try {
      const raw = localStorage.getItem('pushpaka_auth')
      if (raw) {
        const { state } = JSON.parse(raw)
        if (state?.token) {
          config.headers.Authorization = `Bearer ${state.token}`
        }
      }
    } catch {
      // ignore malformed storage
    }
  }
  return config
})

// Handle 401 - redirect to login
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && typeof window !== 'undefined') {
      localStorage.removeItem('pushpaka_auth')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth
export const authApi = {
  register: (data: { email: string; name: string; password: string }) =>
    apiClient.post('/auth/register', data),
  login: (data: { email: string; password: string }) =>
    apiClient.post('/auth/login', data),
}

// Projects
export const projectsApi = {
  list: () => apiClient.get('/projects'),
  get: (id: string) => apiClient.get(`/projects/${id}`),
  create: (data: {
    name: string
    repo_url: string
    branch?: string
    build_command?: string
    start_command?: string
    port?: number
    framework?: string
    is_private?: boolean
    git_token?: string
  }) => apiClient.post('/projects', data),
  update: (id: string, data: {
    name?: string
    branch?: string
    build_command?: string
    start_command?: string
    port?: number
    framework?: string
    is_private?: boolean
    git_token?: string
  }) => apiClient.put(`/projects/${id}`, data),
  delete: (id: string) => apiClient.delete(`/projects/${id}`),
}

// Deployments
export const deploymentsApi = {
  list: (limit = 20, offset = 0) =>
    apiClient.get(`/deployments?limit=${limit}&offset=${offset}`),
  get: (id: string) => apiClient.get(`/deployments/${id}`),
  trigger: (data: { project_id: string; branch?: string; commit_sha?: string }) =>
    apiClient.post('/deployments', data),
  rollback: (id: string) => apiClient.post(`/deployments/${id}/rollback`),
}

// Logs
export const logsApi = {
  get: (deploymentId: string) => apiClient.get(`/logs/${deploymentId}`),
}

// Domains
export const domainsApi = {
  list: () => apiClient.get('/domains'),
  add: (data: { project_id: string; domain: string }) =>
    apiClient.post('/domains', data),
  delete: (id: string) => apiClient.delete(`/domains/${id}`),
}

// Env vars
export const envApi = {
  list: (projectId: string) => apiClient.get(`/env?project_id=${projectId}`),
  set: (data: { project_id: string; key: string; value: string }) =>
    apiClient.post('/env', data),
  delete: (data: { project_id: string; key: string }) =>
    apiClient.delete('/env', { data }),
}

// System capabilities (public  no auth required)
export const systemApi = {
  get: () => apiClient.get('/system'),
}

// Notifications
export const notificationsApi = {
  getConfig: () => apiClient.get('/notifications/config'),
  updateConfig: (data: {
    slack_webhook_url?: string
    discord_webhook_url?: string
    smtp_host?: string
    smtp_port?: number
    smtp_username?: string
    smtp_password?: string
    smtp_from?: string
    smtp_to?: string
    notify_on_success?: boolean
    notify_on_failure?: boolean
  }) => apiClient.put('/notifications/config', data),
}

// Webhooks
export const webhooksApi = {
  list: () => apiClient.get('/webhooks'),
  create: (data: { project_id: string; provider?: string; branch?: string }) =>
    apiClient.post('/webhooks', data),
  delete: (id: string) => apiClient.delete(`/webhooks/${id}`),
}

// Audit logs
export const auditApi = {
  list: (limit = 50, offset = 0) =>
    apiClient.get(`/audit?limit=${limit}&offset=${offset}`),
}

// AI log analysis
export const aiApi = {
  analyzeLogs: (deploymentId: string) =>
    apiClient.post(`/deployments/${deploymentId}/analyze`),
}

