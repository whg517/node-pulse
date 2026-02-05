import { create } from 'zustand'
import { login as apiLogin, logout as apiLogout, getMe as apiGetMe } from '../api/auth'
import { SESSION_EXPIRY_HOURS } from '../config/constants'
import type { User } from './types'

// ============== Types ==============
export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  role: 'admin' | 'operator' | 'viewer' | null
  sessionId: string | null
  sessionExpiry: number | null
  isLoading: boolean
}

export interface AuthActions {
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setUser: (user: User) => void
  clearAuth: () => void
  checkSession: () => boolean
  restoreSession: () => Promise<void>
}

type AuthStore = AuthState & AuthActions

// ============== Store ==============
export const useAuthStore = create<AuthStore>((set, get) => ({
  // State
  user: null,
  isAuthenticated: false,
  role: null,
  sessionId: null,
  sessionExpiry: null,
  isLoading: false,

  // Actions
  login: async (username: string, password: string) => {
    const response = await apiLogin({ username, password })
    const expiry = Date.now() + SESSION_EXPIRY_HOURS * 60 * 60 * 1000

    const user: User = {
      id: response.data.user_id,
      username: response.data.username,
      role: response.data.role,
    }

    set({
      user,
      isAuthenticated: true,
      role: response.data.role,
      sessionId: response.data.user_id, // Using user_id as session identifier
      sessionExpiry: expiry,
    })
  },

  logout: async () => {
    try {
      await apiLogout()
    } catch (error) {
      console.error('Logout API call failed:', error)
      // Continue with local logout even if API call fails
    } finally {
      set({
        user: null,
        isAuthenticated: false,
        role: null,
        sessionId: null,
        sessionExpiry: null,
      })
    }
  },

  setUser: (user: User) => {
    set({
      user,
      isAuthenticated: true,
      role: user.role,
    })
  },

  clearAuth: () => {
    set({
      user: null,
      isAuthenticated: false,
      role: null,
      sessionId: null,
      sessionExpiry: null,
    })
  },

  checkSession: () => {
    const state = get()
    if (!state.sessionExpiry) return false

    const now = Date.now()
    const isValid = state.isAuthenticated && state.sessionExpiry > now

    if (!isValid) {
      set({
        user: null,
        isAuthenticated: false,
        role: null,
        sessionId: null,
        sessionExpiry: null,
      })
    }

    return isValid
  },

  restoreSession: async () => {
    set({ isLoading: true })
    try {
      const response = await apiGetMe()
      const expiry = Date.now() + SESSION_EXPIRY_HOURS * 60 * 60 * 1000

      const user: User = {
        id: response.data.user_id,
        username: response.data.username,
        role: response.data.role,
      }

      set({
        user,
        isAuthenticated: true,
        role: response.data.role,
        sessionId: response.data.user_id,
        sessionExpiry: expiry,
        isLoading: false,
      })
    } catch (error) {
      // Session is invalid or expired
      set({
        user: null,
        isAuthenticated: false,
        role: null,
        sessionId: null,
        sessionExpiry: null,
        isLoading: false,
      })
    }
  },
}))
