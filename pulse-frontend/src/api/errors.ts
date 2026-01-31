/**
 * Custom API Error Classes
 *
 * Provides typed error classes for different API error scenarios
 */

/**
 * Base API Error class
 * All API errors extend this class for consistent error handling
 */
export class ApiError extends Error {
  code: string
  details?: unknown
  status: number

  constructor(message: string, code: string, details?: unknown, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.details = details
    this.status = status || 500
    Object.setPrototypeOf(this, ApiError.prototype)
  }
}

/**
 * Authentication Error (401)
 * Thrown when user is not authenticated or session is invalid
 */
export class AuthenticationError extends ApiError {
  constructor(message: string = 'Authentication failed', details?: unknown) {
    super(message, 'ERR_AUTHENTICATION', details, 401)
    this.name = 'AuthenticationError'
    Object.setPrototypeOf(this, AuthenticationError.prototype)
  }
}

/**
 * Authorization Error (403)
 * Thrown when user lacks permission for the requested resource
 */
export class AuthorizationError extends ApiError {
  constructor(message: string = 'Access forbidden', details?: unknown) {
    super(message, 'ERR_AUTHORIZATION', details, 403)
    this.name = 'AuthorizationError'
    Object.setPrototypeOf(this, AuthorizationError.prototype)
  }
}

/**
 * Validation Error (400)
 * Thrown when request parameters are invalid
 */
export class ValidationError extends ApiError {
  constructor(message: string = 'Validation failed', details?: unknown) {
    super(message, 'ERR_VALIDATION', details, 400)
    this.name = 'ValidationError'
    Object.setPrototypeOf(this, ValidationError.prototype)
  }
}

/**
 * Not Found Error (404)
 * Thrown when requested resource does not exist
 */
export class NotFoundError extends ApiError {
  constructor(message: string = 'Resource not found', details?: unknown) {
    super(message, 'ERR_NOT_FOUND', details, 404)
    this.name = 'NotFoundError'
    Object.setPrototypeOf(this, NotFoundError.prototype)
  }
}

/**
 * Rate Limit Error (429)
 * Thrown when API rate limit is exceeded
 */
export class RateLimitError extends ApiError {
  constructor(message: string = 'Rate limit exceeded', details?: unknown) {
    super(message, 'ERR_RATE_LIMIT', details, 429)
    this.name = 'RateLimitError'
    Object.setPrototypeOf(this, RateLimitError.prototype)
  }
}

/**
 * Type guard for API errors
 * Use this to narrow error types in catch blocks
 *
 * @example
 * try {
 *   await apiCall()
 * } catch (error) {
 *   if (isApiError(error)) {
 *     console.error(error.code, error.status)
 *   }
 * }
 */
export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

/**
 * Type guard for Authentication errors
 */
export function isAuthenticationError(error: unknown): error is AuthenticationError {
  return error instanceof AuthenticationError
}

/**
 * Type guard for Validation errors
 */
export function isValidationError(error: unknown): error is ValidationError {
  return error instanceof ApiError && error.code === 'ERR_VALIDATION'
}

/**
 * Type guard for Not Found errors
 */
export function isNotFoundError(error: unknown): error is NotFoundError {
  return error instanceof ApiError && error.code === 'ERR_NOT_FOUND'
}
