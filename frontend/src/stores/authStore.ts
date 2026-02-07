import { create } from 'zustand'
import { login as apiLogin, logout as apiLogout, getMe as apiGetMe, refreshToken as apiRefreshToken } from '../api/auth'
import { ACCESS_TOKEN_EXPIRY_MINUTES } from '../config/constants'
import type { User } from './types'

// ============== Types ==============
export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  role: 'admin' | 'operator' | 'viewer' | null
  accessToken: string | null
  tokenExpiresAt: number | null
  refreshPromise: Promise<void> | null
  refreshRetryCount: number
  isLoading: boolean
}

export interface AuthActions {
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setUser: (user: User) => void
  clearAuth: () => void
  setAccessToken: (token: string, expiresIn: number) => void
  startTokenExpirationCheck: () => void
  stopTokenExpirationCheck: () => void
  restoreSession: () => Promise<void>
}

type AuthStore = AuthState & AuthActions

// Token pre-refresh interval ID
let preRefreshIntervalId: number | null = null

// ============== Store ==============
export const useAuthStore = create<AuthStore>((set, get) => ({
  // State
  user: null,
  isAuthenticated: false,
  role: null,
  accessToken: null,
  tokenExpiresAt: null,
  refreshPromise: null,
  refreshRetryCount: 0,
  isLoading: false,

  // Actions
  login: async (username: string, password: string) => {
    const response = await apiLogin({ username, password })
    const expiry = Date.now() + ACCESS_TOKEN_EXPIRY_MINUTES * 60 * 1000

    const user: User = {
      id: response.data.user_id,
      username: response.data.username,
      role: response.data.role,
    }

    set({
      user,
      isAuthenticated: true,
      role: response.data.role,
      accessToken: response.data.access_token,
      tokenExpiresAt: expiry,
      refreshPromise: null,
      refreshRetryCount: 0,
    })

    // Start pre-refresh timer
    get().startTokenExpirationCheck()
  },

  logout: async () => {
    // Stop pre-refresh timer
    get().stopTokenExpirationCheck()

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
        accessToken: null,
        tokenExpiresAt: null,
        refreshPromise: null,
        refreshRetryCount: 0,
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
    // Stop pre-refresh timer
    get().stopTokenExpirationCheck()

    set({
      user: null,
      isAuthenticated: false,
      role: null,
      accessToken: null,
      tokenExpiresAt: null,
      refreshPromise: null,
      refreshRetryCount: 0,
    })
  },

  setAccessToken: (token: string, expiresIn: number) => {
    const expiry = Date.now() + expiresIn * 1000
    set({
      accessToken: token,
      tokenExpiresAt: expiry,
      refreshPromise: null,
      refreshRetryCount: 0,
    })
  },

  startTokenExpirationCheck: () => {
    // Clear existing interval if any
    get().stopTokenExpirationCheck()

    // Check every minute
    preRefreshIntervalId = window.setInterval(() => {
      const state = get()
      if (!state.tokenExpiresAt) {
        return
      }

      const now = Date.now()
      const timeUntilExpiry = state.tokenExpiresAt - now
      const PRE_REFRESH_THRESHOLD = 30 * 1000 // 30 seconds

      // If token expires in less than 30 seconds, refresh it
      if (timeUntilExpiry < PRE_REFRESH_THRESHOLD && timeUntilExpiry > 0) {
        console.log('[AuthStore] Token expiring soon, refreshing...')
        apiRefreshToken()
          .then((response) => {
            const expiry = Date.now() + ACCESS_TOKEN_EXPIRY_MINUTES * 60 * 1000
            set({
              accessToken: response.data.access_token,
              tokenExpiresAt: expiry,
            })
            console.log('[AuthStore] Token refreshed successfully')
          })
          .catch((error) => {
            console.error('[AuthStore] Failed to refresh token:', error)
            // If refresh fails, clear auth state
            get().clearAuth()
          })
      }
    }, 60 * 1000) // Check every minute
  },

  stopTokenExpirationCheck: () => {
    if (preRefreshIntervalId !== null) {
      clearInterval(preRefreshIntervalId)
      preRefreshIntervalId = null
    }
  },

  restoreSession: async () => {
    set({ isLoading: true })
    try {
      const response = await apiGetMe()

      const user: User = {
        id: response.data.user_id,
        username: response.data.username,
        role: response.data.role,
      }

      set({
        user,
        isAuthenticated: true,
        role: response.data.role,
        isLoading: false,
      })
    } catch (error) {
      // Session is invalid or expired
      set({
        user: null,
        isAuthenticated: false,
        role: null,
        accessToken: null,
        tokenExpiresAt: null,
        isLoading: false,
      })
    }
  },
}))
