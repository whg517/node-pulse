/**
 * Authentication Store
 *
 * Zustand store for authentication state management.
 *
 * Key Features:
 * - No pre-refresh timer (relies on 401 interceptor)
 * - Cross-tab logout sync via localStorage events
 * - Simplified, single-responsibility API
 * - NO token persistence (tokens in memory only)
 */

import { create } from 'zustand'
import { login as apiLogin, logout as apiLogout, getMe as apiGetMe } from '../api/auth'
import { cancelPendingRequests } from '../api/client'
import { ACCESS_TOKEN_EXPIRY_MINUTES } from '../config/constants'
import type { User } from './types'
import type { UserRole } from '../types/auth'

// ============== Types ==============

export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  role: UserRole | null
  accessToken: string | null
  tokenExpiresAt: number | null
  isLoading: boolean
  refreshFailureCount: number
}

export interface AuthActions {
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setUser: (user: User) => void
  clearAuth: () => void
  setAccessToken: (token: string, expiresIn: number) => void
  restoreSession: () => Promise<void>
}

type AuthStore = AuthState & AuthActions

// ============== Cross-Tab Sync Constants ==============

const LOGOUT_EVENT_KEY = 'auth:logout'
const LOGOUT_EVENT_VALUE = 'logout'

// ============== Store ==============

export const useAuthStore = create<AuthStore>((set) => ({
  // State
  user: null,
  isAuthenticated: false,
  role: null,
  accessToken: null,
  tokenExpiresAt: null,
  isLoading: false,
  refreshFailureCount: 0,

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
      refreshFailureCount: 0,
    })
  },

  logout: async () => {
    try {
      await apiLogout()
    } catch (error) {
      console.error('Logout API call failed:', error)
      // Continue with local logout even if API call fails
    } finally {
      // Broadcast logout to other tabs
      broadcastLogout()

      // Cancel any pending requests
      cancelPendingRequests()

      set({
        user: null,
        isAuthenticated: false,
        role: null,
        accessToken: null,
        tokenExpiresAt: null,
        refreshFailureCount: 0,
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
    // Broadcast logout to other tabs
    broadcastLogout()

    // Cancel any pending requests
    cancelPendingRequests()

    set({
      user: null,
      isAuthenticated: false,
      role: null,
      accessToken: null,
      tokenExpiresAt: null,
      refreshFailureCount: 0,
    })
  },

  setAccessToken: (token: string, expiresIn: number) => {
    const expiry = Date.now() + expiresIn
    set({
      accessToken: token,
      tokenExpiresAt: expiry,
      refreshFailureCount: 0,
    })
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

      const currentState = useAuthStore.getState()
      
      set({
        user,
        isAuthenticated: true,
        role: response.data.role,
        isLoading: false,
        accessToken: currentState.accessToken,
        tokenExpiresAt: currentState.tokenExpiresAt,
      })
    } catch (error) {
      // Session is invalid or expired
      // Only clear auth if we were loading from previous state
      const currentState = useAuthStore.getState()
      if (currentState.isAuthenticated) {
        set({
          isLoading: false,
        })
      } else {
        set({
          user: null,
          isAuthenticated: false,
          role: null,
          accessToken: null,
          tokenExpiresAt: null,
           isLoading: false,
         })
       }
     }
   },
}))

// ============== Cross-Tab Logout Sync ==============

/**
 * Broadcast logout event to other tabs via localStorage
 */
function broadcastLogout(): void {
  try {
    localStorage.setItem(LOGOUT_EVENT_KEY, LOGOUT_EVENT_VALUE)
    // Remove immediately after setting to allow re-triggering
    setTimeout(() => {
      localStorage.removeItem(LOGOUT_EVENT_KEY)
    }, 100)
  } catch (error) {
    console.error('Failed to broadcast logout:', error)
  }
}

/**
 * Setup cross-tab logout listener
 * Call this once in App.tsx during initialization
 * Returns cleanup function
 */
export function setupCrossTabLogoutSync(): () => void {
  const handleStorageEvent = (event: StorageEvent) => {
    if (event.key === LOGOUT_EVENT_KEY && event.newValue === LOGOUT_EVENT_VALUE) {
      if (import.meta.env.DEV) {
        console.log('[AuthStore] Received logout event from another tab')
      }

      // Cancel any pending requests
      cancelPendingRequests()

      // Clear auth state without calling API
      useAuthStore.setState({
        user: null,
        isAuthenticated: false,
        role: null,
        accessToken: null,
        tokenExpiresAt: null,
        refreshFailureCount: 0,
      })
    }
  }

  window.addEventListener('storage', handleStorageEvent)

  // Return cleanup function
  return () => {
    window.removeEventListener('storage', handleStorageEvent)
  }
}

/**
 * Setup visibility change handler for session validation
 * Call this once in App.tsx during initialization
 * Returns cleanup function
 */
export function setupVisibilityHandler(): () => void {
  const handleVisibilityChange = async () => {
    if (document.visibilityState === 'visible') {
      const { isAuthenticated, accessToken, tokenExpiresAt } = useAuthStore.getState()

      // Only validate if we think we're authenticated
      if (isAuthenticated && accessToken && tokenExpiresAt) {
        const now = Date.now()

        // If token is expired, try to restore session
        if (now >= tokenExpiresAt) {
          if (import.meta.env.DEV) {
            console.log('[AuthStore] Token expired on tab focus, restoring session...')
          }
          try {
            await useAuthStore.getState().restoreSession()
          } catch (error) {
            console.warn('[AuthStore] Session restoration failed on tab focus:', error)
          }
         }
       }
     }
   }

   document.addEventListener('visibilitychange', handleVisibilityChange)

   // Return cleanup function
   return () => {
     document.removeEventListener('visibilitychange', handleVisibilityChange)
   }
}
