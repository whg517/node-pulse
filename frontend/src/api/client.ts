/**
 * Unified API Client with JWT Interceptor
 *
 * Provides a base API client function with common configuration
 * for all API calls. Handles JWT authentication, token refresh,
 * error parsing, and response formatting consistently.
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

/**
 * Module-level variables for concurrent refresh control
 */
let refreshPromise: Promise<void> | null = null
const MAX_REFRESH_RETRY = 1
let refreshRetryCount = 0
let consecutiveRefreshFailures = 0 // Track consecutive failures for exponential backoff

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
async function makeRequest<T>(
  endpoint: string,
  options: RequestInit,
  isRetry: boolean
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  // Use getState() instead of hook to access store outside React context
  const authStore = useAuthStore.getState()

  // Get access token from store
  const accessToken = authStore.accessToken

  const config: RequestInit = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken && { Authorization: `Bearer ${accessToken}` }),
      ...options.headers,
    },
    credentials: 'include', // Send HttpOnly cookies (refresh_token)
  }

  const response = await fetch(url, config)

  // Handle successful response
  if (response.ok) {
    return response.json()
  }

  // Handle 401 Unauthorized - attempt token refresh
  if (response.status === 401 && !isRetry && refreshRetryCount < MAX_REFRESH_RETRY) {
    // If no refresh in progress, start one
    if (!refreshPromise) {
      refreshRetryCount++
      refreshPromise = refreshToken()
        .then(() => {
          // Reset consecutive failures on success
          consecutiveRefreshFailures = 0
        })
        .catch((error) => {
          // Increment consecutive failures counter
          consecutiveRefreshFailures++
          console.error(`[apiClient] Token refresh failed (consecutive failures: ${consecutiveRefreshFailures}):`, error)
          throw error
        })
        .finally(() => {
          refreshPromise = null
        })
    }

    // Wait for refresh to complete
    try {
      await refreshPromise
      // Retry original request with new token
      return makeRequest<T>(endpoint, options, true)
    } catch (error) {
      // Refresh failed - clear auth and throw
      // Only logout if we've had multiple consecutive failures (allows transient network errors)
      if (consecutiveRefreshFailures >= 3) {
        authStore.clearAuth()
      }
      throw new AuthenticationError('Session expired. Please login again.')
    }
  }

  // Handle error responses
  await handleError(response)

  // This should never be reached, but TypeScript needs it
  throw new ApiError('Unknown error', 'ERR_UNKNOWN', undefined, response.status)
}

/**
 * Refresh access token using refresh token from cookie
 */
async function refreshToken(): Promise<void> {
  const authStore = useAuthStore.getState()

  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include', // Send HttpOnly cookies
    })

    if (!response.ok) {
      throw new Error('Failed to refresh token')
    }

    const data = await response.json()

    // Update store with new access token
    const expiry = 15 * 60 * 1000 // 15 minutes in milliseconds
    authStore.setAccessToken(data.data.access_token, expiry)

    // Reset retry count on success
    refreshRetryCount = 0
  } catch (error) {
    console.error('[apiClient] Token refresh failed:', error)
    throw error
  }
}

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
