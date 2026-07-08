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
 * Complete 2FA login — exchange the MFA ticket returned by `login` plus a
 * TOTP code for access/refresh tokens. On success the response shape matches
 * a normal login.
 */
export async function mfaLogin(mfaTicket: string, code: string): Promise<LoginResponse> {
  try {
    return await apiClient<LoginResponse>('/api/v1/auth/login/mfa', {
      method: 'POST',
      body: JSON.stringify({ mfa_ticket: mfaTicket, code }),
    })
  } catch (error) {
    throw error instanceof AuthenticationError
      ? error
      : new AuthenticationError(error instanceof Error ? error.message : 'MFA login failed')
  }
}

// --- 2FA self-service (authenticated) ---

export interface MFASetupResponse {
  data: { secret: string; otpauth_uri: string; ticket: string }
  message: string
  timestamp: string
}

/** Begin 2FA enrollment: returns a TOTP secret + otpauth URI and a setup ticket. */
export async function mfaSetup(): Promise<MFASetupResponse> {
  return apiClient<MFASetupResponse>('/api/v1/auth/mfa/setup', { method: 'POST' })
}

/** Confirm enrollment by validating a live TOTP code against the pending secret. */
export async function mfaVerify(ticket: string, code: string): Promise<void> {
  await apiClient<void>('/api/v1/auth/mfa/verify', {
    method: 'POST',
    body: JSON.stringify({ ticket, code }),
  })
}

/** Turn 2FA off (requires the current password as a re-auth guard). */
export async function mfaDisable(password: string): Promise<void> {
  await apiClient<void>('/api/v1/auth/mfa/disable', {
    method: 'POST',
    body: JSON.stringify({ password }),
  })
}

/** Report whether the current user has 2FA enabled (preferences page). */
export async function mfaStatus(): Promise<{ data: { enabled: boolean } }> {
  return apiClient<{ data: { enabled: boolean} }>('/api/v1/auth/mfa/status')
}

// --- Notification preferences (F4 Phase 2, server-side email floor) ---

export interface NotificationPrefsDTO {
  user_id: string
  email_enabled: boolean
  min_alert_level: 'P0' | 'P1' | 'P2'
  notify_email?: string | null
  updated_at: string
}

export interface UpdateNotificationPrefsRequest {
  email_enabled?: boolean
  min_alert_level?: 'P0' | 'P1' | 'P2'
  notify_email?: string
}

/** Fetch the current user's server-side notification preferences. */
export async function getNotificationPrefs(): Promise<{ data: NotificationPrefsDTO }> {
  return apiClient<{ data: NotificationPrefsDTO }>('/api/v1/auth/notification-prefs')
}

/** Update the current user's server-side notification preferences (partial). */
export async function updateNotificationPrefs(
  req: UpdateNotificationPrefsRequest
): Promise<{ data: NotificationPrefsDTO }> {
  return apiClient<{ data: NotificationPrefsDTO }>('/api/v1/auth/notification-prefs', {
    method: 'PUT',
    body: JSON.stringify(req),
  })
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
