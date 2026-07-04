/**
 * System Configuration endpoints (admin only)
 *
 * Backend routes: GET /api/v1/config (masked config view),
 * GET /api/v1/config/validate (re-validate). See G11 in docs/user-journey.md.
 */

import { apiClient } from './client'

/** The effective server configuration with secrets masked/redacted. */
export interface SystemConfigDTO {
  server?: Record<string, unknown>
  database?: Record<string, unknown>
  cleanup?: Record<string, unknown>
  log?: Record<string, unknown>
  cors?: Record<string, unknown>
  admin?: Record<string, unknown>
  session?: Record<string, unknown>
  jwt?: Record<string, unknown>
}

export interface ValidateConfigResult {
  valid: boolean
  warnings?: string[]
  error?: string
}

/** Get the masked effective configuration (admin only). */
export async function getSystemConfig(): Promise<{ data: SystemConfigDTO; message: string; timestamp: string }> {
  return apiClient('/api/v1/config')
}

/** Re-validate the server configuration (admin only). */
export async function validateSystemConfig(): Promise<{ data: ValidateConfigResult; message?: string; timestamp?: string }> {
  return apiClient('/api/v1/config/validate')
}
