/**
 * useAuth Hook
 *
 * Thin wrapper around authStore for convenient access in components.
 * Single API call per action - no duplicate calls.
 *
 * @example
 * const { isAuthenticated, user, login, logout } = useAuth()
 */

import { useAuthStore } from '../stores/authStore'
import type { LoginRequest, LoginResponse, LogoutResponse } from '../api/auth'

export function useAuth() {
  const {
    isAuthenticated,
    user,
    role,
    tokenExpiresAt,
    isLoading,
    login: storeLogin,
    logout: storeLogout,
  } = useAuthStore()

  /**
   * Login with credentials
   * Makes a single API call through the store
   */
  const login = async (credentials: LoginRequest): Promise<void> => {
    return storeLogin(credentials.username, credentials.password)
  }

  /**
   * Logout current user
   * Makes a single API call through the store
   */
  const logout = async (): Promise<void> => {
    return storeLogout()
  }

  /**
   * Check if current session is valid (token not expired)
   */
  const isValidSession = (): boolean => {
    return tokenExpiresAt !== null && tokenExpiresAt > Date.now()
  }

  return {
    // State
    isAuthenticated,
    isLoading,
    user,
    userId: user?.id || null,
    username: user?.username || null,
    role,
    tokenExpiresAt,

    // Actions
    login,
    logout,
    isValidSession,
  }
}

// Re-export types for convenience
export type { LoginRequest, LoginResponse, LogoutResponse }
