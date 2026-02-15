/**
 * Unified API Client with JWT Interceptor
 *
 * Provides a base API client function with common configuration
 * for all API calls. Handles JWT authentication, token refresh,
 * error parsing, and response formatting consistently.
 *
 * Key Features:
 * - Request coalescing for concurrent 401s (single shared refresh Promise)
 * - 10-second timeout on refresh requests
 * - Token expiry pre-check with mutex (proactive refresh)
 * - Logout endpoint bypass (prevents infinite logout loops)
 * - 5xx error handling (graceful degradation, no logout)
 * - Token redaction in logs for security
 */

import { API_BASE_URL } from '../config/constants'
import { useAuthStore } from '../stores/authStore'
import {
  ApiError,
  AuthenticationError,
  ValidationError,
  NotFoundError,
  RateLimitError,
  AuthorizationError,
} from './errors'

// ============== Module-level State ==============

/**
 * Shared refresh Promise for request coalescing
 * All concurrent 401s await the same refresh, preventing race conditions
 */
let refreshPromise: Promise<string> | null = null

/**
 * Mutex for token expiry pre-check
 * Prevents multiple simultaneous pre-checks from triggering duplicate refreshes
 */
let preCheckMutex = false

/**
 * Track consecutive refresh failures for graceful degradation
 */
let consecutiveRefreshFailures = 0

/**
 * Maximum consecutive failures before forcing logout
 */
const MAX_CONSECUTIVE_FAILURES = 3

/**
 * Pre-check threshold: refresh if token expires within 30 seconds
 */
const PRE_CHECK_THRESHOLD_MS = 30_000

/**
 * Refresh request timeout in milliseconds
 */
const REFRESH_TIMEOUT_MS = 10_000

/**
 * Pending request AbortControllers for cleanup on logout
 */
const pendingRequests = new Set<AbortController>()

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
 * - Token expiry pre-check with proactive refresh
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
  const { accessToken, tokenExpiresAt } = authStore

  // Token expiry pre-check with mutex (proactive refresh)
  // Skip for logout endpoint to prevent infinite loops
  if (!shouldBypassRefresh(endpoint) && !isRetry && accessToken && tokenExpiresAt) {
    const now = Date.now()
    const timeUntilExpiry = tokenExpiresAt - now

    // If token expires within threshold, proactively refresh
    if (timeUntilExpiry <= PRE_CHECK_THRESHOLD_MS && timeUntilExpiry > 0 && !preCheckMutex) {
      preCheckMutex = true
      try {
        console.log('[apiClient] Token expiring soon, proactive refresh...')
        await performRefresh()
      } catch (error) {
        // Log but continue - the 401 interceptor will handle it
        console.warn('[apiClient] Proactive refresh failed:', error)
      } finally {
        preCheckMutex = false
      }
    }
  }

  // Get fresh token after potential pre-check refresh
  const currentToken = useAuthStore.getState().accessToken

  const config: RequestInit = {
    ...options,
    signal: abortController.signal,
    headers: {
      'Content-Type': 'application/json',
      ...(currentToken && { Authorization: `Bearer ${currentToken}` }),
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
      // Request coalescing: all concurrent 401s await the same refresh Promise
      if (!refreshPromise) {
        refreshPromise = performRefresh()
          .finally(() => {
            refreshPromise = null
          })
      }

      try {
        await refreshPromise
        // Retry original request with new token
        return makeRequest<T>(endpoint, options, true)
      } catch (_error) {
        // Refresh failed - check if we should force logout
        if (consecutiveRefreshFailures >= MAX_CONSECUTIVE_FAILURES) {
          console.error('[apiClient] Too many consecutive refresh failures, forcing logout')
          authStore.clearAuth()
          cancelPendingRequests()
        }
        throw new AuthenticationError('Session expired. Please login again.')
      }
    }

    // Handle 5xx errors - don't logout, just throw
    if (response.status >= 500) {
      console.error(`[apiClient] Server error (${response.status}) on ${endpoint}`)
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
      // Check for 5xx errors - don't count against consecutive failures
      if (response.status >= 500) {
        console.error('[apiClient] Refresh failed with server error:', response.status)
        throw new ApiError('Server error during refresh', 'ERR_SERVER_ERROR', undefined, response.status)
      }

      consecutiveRefreshFailures++
      console.error(`[apiClient] Token refresh failed (consecutive: ${consecutiveRefreshFailures})`)
      throw new AuthenticationError('Failed to refresh token')
    }

    const data = await response.json()

    // Update store with new access token
    const expiresIn = data.data.expires_in || 900 // Default 15 minutes
    const expiry = Date.now() + expiresIn * 1000
    authStore.setAccessToken(data.data.access_token, expiresIn * 1000)

    // Reset consecutive failures on success
    consecutiveRefreshFailures = 0

    console.log(`[apiClient] Token refreshed successfully, expires at ${new Date(expiry).toISOString()}`)

    return data.data.access_token
  } catch (error) {
    clearTimeout(timeoutId)

    // Handle timeout specifically
    if (error instanceof Error && error.name === 'AbortError') {
      console.error('[apiClient] Token refresh timed out after 10 seconds')
      consecutiveRefreshFailures++
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
  console.log('[apiClient] Cancelled all pending requests')
}

/**
 * Reset consecutive failure counter (for testing/manual reset)
 */
export function resetFailureCount(): void {
  consecutiveRefreshFailures = 0
}

/**
 * Reset all module state (for testing only)
 */
export function resetModuleState(): void {
  consecutiveRefreshFailures = 0
  refreshPromise = null
  preCheckMutex = false
  pendingRequests.clear()
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
    console.error(`[apiClient] Auth error: ${code} - ${message}`)
  }

  // Map HTTP status codes to error classes
  switch (response.status) {
    case 400:
      throw new ValidationError(message, details)
    case 401:
      throw new AuthenticationError(message, details)
    case 403:
      throw new AuthorizationError(message, details)
    case 404:
      throw new NotFoundError(message, details)
    case 429:
      throw new RateLimitError(message, details)
    default:
      throw new ApiError(message, code, details, response.status)
  }
}
