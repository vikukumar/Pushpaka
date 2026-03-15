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
    const token = localStorage.getItem('pushpaka_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }
  return config
})

// Handle 401 - redirect to login
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && typeof window !== 'undefined') {
      localStorage.removeItem('pushpaka_token')
      localStorage.removeItem('pushpaka_user')
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
  }) => apiClient.post('/projects', data),
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
