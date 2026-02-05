/**
 * Authentication API endpoints
 *
 * Provides typed functions for user authentication operations
 * including login, logout, and session management.
 */

import { apiClient } from './client'
import type { LoginRequest, LoginResponse, LogoutResponse, GetMeResponse } from '../types/auth'
import { AuthenticationError } from './errors'

export type { LoginRequest, LoginResponse, LogoutResponse, GetMeResponse }

const SESSION_COOKIE_NAME = 'session_id'

/**
 * User login
 *
 * Authenticates user with username and password.
 * On success, Session Cookie is automatically set by the server.
 *
 * @param credentials - User credentials (username, password)
 * @returns Login response with user info and session details
 * @throws AuthenticationError if login fails
 * @throws ValidationError if credentials are invalid
 *
 * @example
 * try {
 *   const { user } = await login({ username: 'admin', password: 'pass123' })
 *   console.log('Logged in as', user.username)
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
    if (error instanceof Error && error.name === 'AuthenticationError') {
      throw error
    }
    // Wrap other errors as AuthenticationError
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Login failed'
    )
  }
}

/**
 * User logout
 *
 * Clears the current user session.
 * Session Cookie is automatically cleared by the server.
 *
 * @returns Logout response
 * @throws AuthenticationError if logout fails
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
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Logout failed'
    )
  }
}

/**
 * Session cookie name constant
 * Used by the server to set and read the session cookie
 */
export { SESSION_COOKIE_NAME }

/**
 * Get current user
 *
 * Retrieves the currently authenticated user's information.
 * Session Cookie is automatically sent by the browser.
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
    if (error instanceof Error && error.name === 'AuthenticationError') {
      throw error
    }
    // Wrap other errors as AuthenticationError
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Failed to get user information'
    )
  }
}
