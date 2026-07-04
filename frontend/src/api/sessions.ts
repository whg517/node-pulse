/**
 * Session Management API endpoints
 *
 * Provides typed functions for viewing and managing user sessions.
 */

import { apiClient } from './client'
import type { SessionListResponse, SessionInfoResponse } from '../types/auth'

// Re-export types for convenience
export type { SessionListResponse, SessionInfoResponse }

/**
 * Get all active sessions for current user
 *
 * Retrieves a list of all active sessions including the current session.
 *
 * @returns Session list response with sessions array
 * @throws AuthenticationError if not authenticated
 * @throws ApiError if request fails
 *
 * @example
 * const { data } = await getSessions()
 * console.log(`You have ${data.total} active sessions`)
 */
export async function getSessions(): Promise<SessionListResponse> {
  return apiClient<SessionListResponse>('/api/v1/auth/sessions', {
    method: 'GET',
  })
}

/**
 * Delete (revoke) a specific session
 *
 * Revokes a session by ID. If revoking the current session,
 * user will be logged out.
 *
 * @param sessionId - The ID of the session to revoke
 * @returns void on success
 * @throws AuthenticationError if not authenticated
 * @throws NotFoundError if session doesn't exist
 * @throws ApiError if request fails
 *
 * @example
 * await deleteSession('session-123')
 * console.log('Session revoked')
 */
export async function deleteSession(sessionId: string): Promise<void> {
  await apiClient<void>(`/api/v1/auth/sessions/${sessionId}`, {
    method: 'DELETE',
  })
}

/**
 * Get current session info
 *
 * Returns information about the current session including
 * the current session ID and active sessions count.
 *
 * @returns Session info response
 * @throws AuthenticationError if not authenticated
 * @throws ApiError if request fails
 *
 * @example
 * const { data } = await getSessionInfo()
 * console.log(`Current session: ${data.current_session_id}`)
 */
export async function getSessionInfo(): Promise<SessionInfoResponse> {
  return apiClient<SessionInfoResponse>('/api/v1/auth/session-info', {
    method: 'GET',
  })
}

/**
 * Revoke all sessions for the current user
 *
 * Revokes every active session including the current one. After calling this,
 * the user must log in again. Backend endpoint: POST /auth/sessions/revoke-all.
 *
 * @returns void on success
 * @throws AuthenticationError if not authenticated
 * @throws ApiError if request fails
 */
export async function revokeAllSessions(): Promise<void> {
  await apiClient<void>('/api/v1/auth/sessions/revoke-all', {
    method: 'POST',
  })
}
