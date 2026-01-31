/**
 * Unified API Client
 *
 * Provides a base API client function with common configuration
 * for all API calls. Handles authentication, error parsing, and
 * response formatting consistently.
 */

import { API_BASE_URL } from '../config/constants'
import {
  ApiError,
  AuthenticationError,
  ValidationError,
  NotFoundError,
  RateLimitError,
  AuthorizationError,
} from './errors'

/**
 * Base API client function
 *
 * Handles all common API call logic:
 * - Sets Content-Type headers
 * - Includes Session Cookie (credentials: 'include')
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
  const url = `${API_BASE_URL}${endpoint}`

  const config: RequestInit = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include', // Send Session Cookie for authentication
  }

  const response = await fetch(url, config)

  // Handle error responses
  if (!response.ok) {
    await handleError(response)
  }

  return response.json()
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
