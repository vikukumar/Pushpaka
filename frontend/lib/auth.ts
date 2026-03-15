import { create } from 'zustand'
import { User } from '@/types'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  setAuth: (user: User, token: string) => void
  clearAuth: () => void
  loadFromStorage: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,

  setAuth: (user, token) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('pushpaka_token', token)
      localStorage.setItem('pushpaka_user', JSON.stringify(user))
    }
    set({ user, token, isAuthenticated: true })
  },

  clearAuth: () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('pushpaka_token')
      localStorage.removeItem('pushpaka_user')
    }
    set({ user: null, token: null, isAuthenticated: false })
  },

  loadFromStorage: () => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('pushpaka_token')
      const userStr = localStorage.getItem('pushpaka_user')
      if (token && userStr) {
        try {
          const user = JSON.parse(userStr)
          set({ user, token, isAuthenticated: true })
        } catch {
          localStorage.removeItem('pushpaka_token')
          localStorage.removeItem('pushpaka_user')
        }
      }
    }
  },
}))
