/**
 * Authentication API endpoints
 *
 * Provides typed functions for user authentication operations
 * including login, logout, token refresh, and session management.
 */

import { apiClient } from './client'
import type { LoginRequest, LoginResponse, LogoutResponse, GetMeResponse, RefreshResponse } from '../types/auth'
import { AuthenticationError } from './errors'

export type { LoginRequest, LoginResponse, LogoutResponse, GetMeResponse, RefreshResponse }

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
    if (error instanceof Error && error.name === 'AuthenticationError') {
      throw error
    }
    // Wrap other errors as AuthenticationError
    throw new AuthenticationError(
      error instanceof Error ? error.message : 'Failed to get user information'
    )
  }
}
