/**
 * Unified API Client with JWT Interceptor
 *
 * Provides a base API client function with common configuration
 * for all API calls. Handles JWT authentication, token refresh,
 * error parsing, and response formatting consistently.
 *
 * Key Features:
 * - 10-second timeout on refresh requests
 * - Logout endpoint bypass (prevents infinite logout loops)
 * - 5xx error handling
 * - Token redaction in logs for security
 */

import { API_BASE_URL } from '../config/constants'
import { useAuthStore } from '../stores/authStore'
import {
  ApiError,
  AuthenticationError,
  ValidationError,
  NotFoundError,
  AuthorizationError,
} from './errors'

// Re-export error classes for consumers
export {
  ApiError,
  AuthenticationError,
  ValidationError,
  NotFoundError,
  AuthorizationError,
}

// ============== Module-level State ==============

/**
 * Refresh request timeout in milliseconds
 */
const REFRESH_TIMEOUT_MS = 10_000

/**
 * Pending request AbortControllers for cleanup on logout
 */
const pendingRequests = new Set<AbortController>()
let inFlightRefreshPromise: Promise<string> | null = null

function shouldLogApiClientDebug(): boolean {
  return import.meta.env.DEV && import.meta.env.MODE !== 'test'
}

function shouldLogApiClientErrors(): boolean {
  return import.meta.env.MODE !== 'test'
}

// ============== Utility Functions ==============

/**
 * Check if endpoint should bypass token refresh
 * Login: 401 means invalid credentials, not expired token
 * Logout: prevent infinite loops
 */
function shouldBypassRefresh(endpoint: string): boolean {
  return endpoint.includes('/auth/login') || endpoint.includes('/auth/logout')
}

// ============== Main API Client ==============

/**
 * Base API client function
 *
 * Handles all common API call logic:
 * - Sets Content-Type headers
 * - Includes JWT access token in Authorization header
 * - Automatically refreshes token on 401 response
 * - Retries original request after successful refresh
 * - Parses error responses
 * - Maps HTTP status codes to appropriate error classes
 *
 * @param endpoint - API endpoint path (e.g., '/api/v1/nodes')
 * @param options - Fetch request options (method, body, etc.)
 * @returns Parsed JSON response data
 * @throws ApiError or its subclasses on HTTP errors
 *
 * @example
 * const data = await apiClient<{ data: Node[] }>('/api/v1/nodes')
 */
export async function apiClient<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  return makeRequest<T>(endpoint, options, false)
}

/**
 * Internal request function with retry logic
 */
/**
 * Get CSRF token from cookie as fallback
 */
function getCsrfTokenFromCookie(): string | null {
  if (typeof document === 'undefined') return null
  const cookies = document.cookie.split(';')
  for (const cookie of cookies) {
    const [name, value] = cookie.trim().split('=')
    if (name === 'csrf_token' && value) {
      return value
    }
  }
  return null
}

async function makeRequest<T>(
  endpoint: string,
  options: RequestInit,
  isRetry: boolean
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const authStore = useAuthStore.getState()

  // Create AbortController for this request
  const abortController = new AbortController()
  pendingRequests.add(abortController)

  // Get current auth state
  const { accessToken } = authStore
  let { csrfToken } = useAuthStore.getState()

  // Fallback: Get CSRF token from cookie if not in store
  // This handles cases where the store state is not properly hydrated
  if (!csrfToken) {
    csrfToken = getCsrfTokenFromCookie()
  }

  // Determine if this is a state-changing request that needs CSRF protection
  const method = (options.method || 'GET').toUpperCase()
  const isMutation = method === 'POST' || method === 'PUT' || method === 'PATCH' || method === 'DELETE'

  const config: RequestInit = {
    ...options,
    signal: abortController.signal,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken && { Authorization: `Bearer ${accessToken}` }),
      ...(isMutation && csrfToken && { 'X-CSRF-Token': csrfToken }),
      ...options.headers,
    },
    credentials: 'include', // Send HttpOnly cookies (refresh_token)
  }

  try {
    const response = await fetch(url, config)

    // Handle successful response
    if (response.ok) {
      return response.json()
    }

    // Handle 401 Unauthorized - attempt token refresh
    if (response.status === 401 && !isRetry && !shouldBypassRefresh(endpoint)) {
      try {
        await performRefresh()
        // Retry original request with new token
        return makeRequest<T>(endpoint, options, true)
      } catch (refreshError) {
        // 429 rate-limit or 409 conflict on refresh: don't force logout, propagate as-is
        if (refreshError instanceof ApiError && (refreshError.status === 429 || refreshError.status === 409)) {
          throw refreshError
        }
        authStore.clearAuth()
        cancelPendingRequests()
        throw new AuthenticationError('Session expired. Please login again.')
      }
    }

    // Handle 5xx errors - don't logout, just throw
    if (response.status >= 500) {
      if (shouldLogApiClientErrors()) {
        console.error(`[apiClient] Server error (${response.status}) on ${endpoint}`)
      }
      throw new ApiError(
        'Server temporarily unavailable. Please try again later.',
        'ERR_SERVER_ERROR',
        undefined,
        response.status
      )
    }

    // Handle other error responses
    await handleError(response)

    // This should never be reached, but TypeScript needs it
    throw new ApiError('Unknown error', 'ERR_UNKNOWN', undefined, response.status)
  } finally {
    pendingRequests.delete(abortController)
  }
}

// ============== Token Refresh ==============

/**
 * Perform token refresh with timeout and error handling
 * @returns The new access token
 */
async function performRefresh(): Promise<string> {
  if (inFlightRefreshPromise) {
    return inFlightRefreshPromise
  }

  inFlightRefreshPromise = executeRefresh()

  try {
    return await inFlightRefreshPromise
  } finally {
    inFlightRefreshPromise = null
  }
}

async function executeRefresh(): Promise<string> {
  const authStore = useAuthStore.getState()

  // Create AbortController for timeout
  const abortController = new AbortController()
  const timeoutId = setTimeout(() => abortController.abort(), REFRESH_TIMEOUT_MS)

  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include', // Send HttpOnly cookies
      signal: abortController.signal,
    })

    clearTimeout(timeoutId)

    if (!response.ok) {
      // Check for 5xx errors
      if (response.status >= 500) {
        if (shouldLogApiClientErrors()) {
          console.error('[apiClient] Refresh failed with server error:', response.status)
        }
        throw new ApiError('Server error during refresh', 'ERR_SERVER_ERROR', undefined, response.status)
      }

      // 429 is a transient rate-limit, not an auth failure - don't count or logout
      if (response.status === 429) {
        if (shouldLogApiClientErrors()) {
          console.warn('[apiClient] Token refresh rate-limited (429), will retry later')
        }
        throw new ApiError('Too many refresh requests, please wait a moment', 'ERR_RATE_LIMIT_EXCEEDED', undefined, 429)
      }

      // 409 Conflict means token was already used by another concurrent refresh
      // This is not an auth failure - another refresh succeeded, so we should retry
      if (response.status === 409) {
        if (shouldLogApiClientErrors()) {
          console.warn('[apiClient] Token refresh conflict (409) - token already used by concurrent request')
        }
        // Wait a brief moment for the other refresh to complete and update the store
        await new Promise(resolve => setTimeout(resolve, 100))
        // Check if we now have a valid access token from the other refresh
        const { accessToken } = useAuthStore.getState()
        if (accessToken) {
          if (shouldLogApiClientDebug()) {
            console.log('[apiClient] Access token available after 409, returning existing token')
          }
          return accessToken
        }
        // If still no token, throw a retryable error
        throw new ApiError('Token already used, please retry', 'ERR_TOKEN_CONFLICT', undefined, 409)
      }

      throw new AuthenticationError('Failed to refresh token')
    }

    const data = await response.json()

    // Update store with new access token
    const expiresIn = data.data.expires_in || 900 // Default 15 minutes
    const expiry = Date.now() + expiresIn * 1000
    authStore.setAccessToken(data.data.access_token, expiresIn * 1000)

    if (shouldLogApiClientDebug()) {
      console.log(`[apiClient] Token refreshed successfully, expires at ${new Date(expiry).toISOString()}`)
    }

    return data.data.access_token
  } catch (error) {
    clearTimeout(timeoutId)

    // Handle timeout specifically
    if (error instanceof Error && error.name === 'AbortError') {
      if (shouldLogApiClientErrors()) {
        console.error('[apiClient] Token refresh timed out after 10 seconds')
      }
      throw new AuthenticationError('Token refresh timed out')
    }

    throw error
  }
}

/**
 * Cancel all pending requests (called on logout)
 */
export function cancelPendingRequests(): void {
  pendingRequests.forEach((controller) => {
    controller.abort()
  })
  pendingRequests.clear()
  if (shouldLogApiClientDebug()) {
    console.log('[apiClient] Cancelled all pending requests')
  }
}

/**
 * Reset all module state (for testing only)
 */
export function resetModuleState(): void {
  pendingRequests.clear()
  inFlightRefreshPromise = null
}

// ============== Error Handling ==============

/**
 * Parse HTTP error response and throw appropriate error class
 */
async function handleError(response: Response): Promise<never> {
  let errorData: unknown

  try {
    errorData = await response.json()
  } catch {
    // If response is not JSON, create minimal error data
    errorData = {
      code: 'ERR_UNKNOWN',
      message: response.statusText || 'API request failed',
    }
  }

  // Extract error information
  const message =
    typeof errorData === 'object' && errorData !== null && 'message' in errorData
      ? String(errorData.message)
      : 'API request failed'

  const code =
    typeof errorData === 'object' && errorData !== null && 'code' in errorData
      ? String(errorData.code)
      : 'ERR_UNKNOWN'

  const details =
    typeof errorData === 'object' && errorData !== null && 'details' in errorData
      ? errorData.details
      : undefined

  // Log token-redacted errors for debugging
  if (code.includes('TOKEN') || code.includes('AUTH')) {
    if (shouldLogApiClientErrors()) {
      console.error(`[apiClient] Auth error: ${code} - ${message}`)
    }
  }

  // Map HTTP status codes to error classes
  // Preserve original error code from backend when available
  switch (response.status) {
    case 400:
      throw new ValidationError(message, details)
    case 401:
      // Throw AuthenticationError for 401 responses, preserving details
      throw new AuthenticationError(message, details)
    case 403:
      throw new AuthorizationError(message, details)
    case 404:
      throw new NotFoundError(message, details)
    case 429:
      // Preserve original error code (ERR_RATE_LIMITED, ERR_RATE_LIMIT_EXCEEDED, etc.)
      throw new ApiError(message, code, details, 429)
    default:
      throw new ApiError(message, code, details, response.status)
  }
}
