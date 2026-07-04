/**
 * Admin Auth endpoints (admin only)
 *
 * Covers admin-only session/security actions that go beyond self-service.
 * Backend route: POST /api/v1/admin/auth/revoke-all/:userId (admin only) —
 * force-revoke all sessions for a target user (see G7 in docs/user-journey.md).
 */

import { apiClient } from './client'

/**
 * Force-revoke all sessions for a target user.
 *
 * Used by admins to terminate a user's sessions during a security incident
 * (e.g. suspected compromise). The target user's refresh tokens are revoked;
 * their access token is blacklisted on the next request.
 */
export async function adminRevokeAllUserSessions(userId: string): Promise<{ message: string }> {
  return apiClient<{ message: string }>(`/api/v1/admin/auth/revoke-all/${userId}`, {
    method: 'POST',
  })
}
