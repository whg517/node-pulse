/**
 * Authentication API endpoints
 *
 * Provides typed functions for user authentication operations
 * including login, logout, token refresh, and session management.
 *
 * Each function makes exactly ONE API call - no duplicate calls.
 */

import { apiClient } from './client'
import type {
  LoginRequest,
  LoginResponse,
  LogoutResponse,
  GetMeResponse,
  RefreshResponse,
} from '../types/auth'
import { AuthenticationError } from './errors'

// Re-export types for convenience
export type {
  LoginRequest,
  LoginResponse,
  LogoutResponse,
  GetMeResponse,
  RefreshResponse,
}

/**
 * User login
 *
 * Authenticates user with username and password.
 * On success, access token is returned in response body
 * and refresh token is set in HttpOnly cookie.
 *
 * @param credentials - User credentials (username, password)
 * @returns Login response with user info and access token
 * @throws AuthenticationError if login fails
 * @throws ValidationError if credentials are invalid
 *
 * @example
 * try {
 *   const { data } = await login({ username: 'admin', password: 'pass123' })
 *   console.log('Logged in as', data.username)
 * } catch (error) {
 *   if (isAuthenticationError(error)) {
 *     console.error('Login failed:', error.message)
 *   }
 * }
 */
export async function login(credentials: LoginRequest): Promise<LoginResponse> {
  try {
    return await apiClient<LoginResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    })
  } catch (error) {
    // Ensure authentication errors are properly typed
    if (error instanceof AuthenticationError) {
      throw error
    }
    // Wrap other errors as AuthenticationError
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Login failed'
    )
  }
}

/**
 * Refresh access token
 *
 * Uses the refresh token from HttpOnly cookie to get a new access token.
 * Implements token rotation - old refresh token is deleted, new one is set.
 *
 * @returns Refresh response with new access token
 * @throws AuthenticationError if refresh token is invalid or expired
 *
 * @example
 * try {
 *   const { data } = await refreshToken()
 *   console.log('New access token:', data.access_token)
 * } catch (error) {
 *   console.error('Token refresh failed:', error.message)
 * }
 */
export async function refreshToken(): Promise<RefreshResponse> {
  try {
    return await apiClient<RefreshResponse>('/api/v1/auth/refresh', {
      method: 'POST',
    })
  } catch (error) {
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Failed to refresh token'
    )
  }
}

/**
 * User logout
 *
 * Clears the current user session.
 * Refresh token is deleted from database and HttpOnly cookie is cleared.
 *
 * @returns Logout response
 * @throws AuthenticationError if logout fails (but local state should still be cleared)
 *
 * @example
 * await logout()
 * console.log('Logged out successfully')
 */
export async function logout(): Promise<LogoutResponse> {
  try {
    return await apiClient<LogoutResponse>('/api/v1/auth/logout', {
      method: 'POST',
    })
  } catch (error) {
    // Re-throw for caller to handle, but they should still clear local state
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Logout failed'
    )
  }
}

/**
 * Get current user
 *
 * Retrieves the currently authenticated user's information.
 * Access token is sent in Authorization header.
 *
 * @returns Current user response
 * @throws AuthenticationError if not authenticated
 * @throws ApiError if request fails
 *
 * @example
 * try {
 *   const { data } = await getMe()
 *   console.log('Current user:', data.username)
 * } catch (error) {
 *   if (isAuthenticationError(error)) {
 *     console.error('Not authenticated')
 *   }
 * }
 */
export async function getMe(): Promise<GetMeResponse> {
  try {
    return await apiClient<GetMeResponse>('/api/v1/auth/me', {
      method: 'GET',
    })
  } catch (error) {
    // Ensure authentication errors are properly typed
    if (error instanceof AuthenticationError) {
      throw error
    }
    // Wrap other errors as AuthenticationError
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Failed to get user information'
    )
  }
}

/** Response shape for the change-password endpoint. */
export interface ChangePasswordResponse {
  message: string
  // The backend revokes all other sessions after a successful password change;
  // the current session is kept (best-effort within the last minute).
  sessions_revoked: boolean
  timestamp?: string
}

/**
 * Change the current user's password
 *
 * Requires the current password for verification. On success the backend
 * revokes all other sessions; the current session is kept (best-effort).
 * Backend endpoint: POST /auth/password/change (CSRF-protected).
 *
 * @param currentPassword - Current password for verification
 * @param newPassword - New password (must pass backend strength validation)
 * @returns Change-password response with sessions_revoked flag
 * @throws AuthenticationError if current password is wrong or not authenticated
 * @throws ValidationError if the new password is weak or identical to the current
 */
export async function changePassword(
  currentPassword: string,
  newPassword: string
): Promise<ChangePasswordResponse> {
  return apiClient<ChangePasswordResponse>('/api/v1/auth/password/change', {
    method: 'POST',
    body: JSON.stringify({
      current_password: currentPassword,
      new_password: newPassword,
    }),
  })
}

/** Request a password-reset link. Anti-enumeration: response is identical whether the email exists. */
export async function requestPasswordReset(email: string): Promise<{ message: string }> {
  return apiClient<{ message: string }>('/api/v1/auth/password/reset/request', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

/** Confirm a password reset using the token from the email link. */
export async function confirmPasswordReset(token: string, newPassword: string): Promise<{ message: string }> {
  return apiClient<{ message: string }>('/api/v1/auth/password/reset/confirm', {
    method: 'POST',
    body: JSON.stringify({ token, new_password: newPassword }),
  })
}
