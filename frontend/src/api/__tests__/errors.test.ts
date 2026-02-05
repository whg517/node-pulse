/**
 * Tests for API error classes and type guards
 */

import { describe, it, expect } from 'vitest'
import {
  ApiError,
  AuthenticationError,
  AuthorizationError,
  ValidationError,
  NotFoundError,
  RateLimitError,
  isApiError,
  isAuthenticationError,
  isValidationError,
  isNotFoundError,
} from '../errors'

describe('Error Classes', () => {
  describe('ApiError', () => {
    it('should create base ApiError with all properties', () => {
      const details = { field: 'value' }
      const error = new ApiError('Test error', 'ERR_TEST', details, 500)

      expect(error).toBeInstanceOf(Error)
      expect(error).toBeInstanceOf(ApiError)
      expect(error.message).toBe('Test error')
      expect(error.code).toBe('ERR_TEST')
      expect(error.details).toEqual(details)
      expect(error.status).toBe(500)
      expect(error.name).toBe('ApiError')
    })

    it('should default to status 500 if not provided', () => {
      const error = new ApiError('Test', 'ERR_TEST')
      expect(error.status).toBe(500)
    })
  })

  describe('AuthenticationError', () => {
    it('should create AuthenticationError with status 401', () => {
      const error = new AuthenticationError('Auth failed')

      expect(error).toBeInstanceOf(ApiError)
      expect(error).toBeInstanceOf(AuthenticationError)
      expect(error.message).toBe('Auth failed')
      expect(error.code).toBe('ERR_AUTHENTICATION')
      expect(error.status).toBe(401)
      expect(error.name).toBe('AuthenticationError')
    })

    it('should accept custom message and details', () => {
      const details = { attempts: 3 }
      const error = new AuthenticationError('Invalid credentials', details)

      expect(error.message).toBe('Invalid credentials')
      expect(error.details).toEqual(details)
    })
  })

  describe('AuthorizationError', () => {
    it('should create AuthorizationError with status 403', () => {
      const error = new AuthorizationError('Access forbidden')

      expect(error).toBeInstanceOf(ApiError)
      expect(error).toBeInstanceOf(AuthorizationError)
      expect(error.code).toBe('ERR_AUTHORIZATION')
      expect(error.status).toBe(403)
      expect(error.name).toBe('AuthorizationError')
    })
  })

  describe('ValidationError', () => {
    it('should create ValidationError with status 400', () => {
      const error = new ValidationError('Invalid input')

      expect(error).toBeInstanceOf(ApiError)
      expect(error).toBeInstanceOf(ValidationError)
      expect(error.code).toBe('ERR_VALIDATION')
      expect(error.status).toBe(400)
      expect(error.name).toBe('ValidationError')
    })
  })

  describe('NotFoundError', () => {
    it('should create NotFoundError with status 404', () => {
      const error = new NotFoundError('Resource not found')

      expect(error).toBeInstanceOf(ApiError)
      expect(error).toBeInstanceOf(NotFoundError)
      expect(error.code).toBe('ERR_NOT_FOUND')
      expect(error.status).toBe(404)
      expect(error.name).toBe('NotFoundError')
    })
  })

  describe('RateLimitError', () => {
    it('should create RateLimitError with status 429', () => {
      const error = new RateLimitError('Too many requests')

      expect(error).toBeInstanceOf(ApiError)
      expect(error).toBeInstanceOf(RateLimitError)
      expect(error.code).toBe('ERR_RATE_LIMIT')
      expect(error.status).toBe(429)
      expect(error.name).toBe('RateLimitError')
    })
  })
})

describe('Type Guards', () => {
  describe('isApiError', () => {
    it('should return true for ApiError instances', () => {
      const error = new ApiError('Test', 'ERR_TEST')
      expect(isApiError(error)).toBe(true)
    })

    it('should return true for ApiError subclasses', () => {
      const authError = new AuthenticationError('Test')
      expect(isApiError(authError)).toBe(true)
    })

    it('should return false for non-ApiError errors', () => {
      const error = new Error('Regular error')
      expect(isApiError(error)).toBe(false)
    })

    it('should return false for non-Error values', () => {
      expect(isApiError('string')).toBe(false)
      expect(isApiError(null)).toBe(false)
      expect(isApiError(undefined)).toBe(false)
      expect(isApiError({})).toBe(false)
    })

    it('should narrow type in TypeScript', () => {
      const error: unknown = new AuthenticationError('Test')

      if (isApiError(error)) {
        // TypeScript should know error is ApiError here
        expect(error.code).toBeDefined()
        expect(error.status).toBeDefined()
      } else {
        // This should not happen
        expect(true).toBe(false)
      }
    })
  })

  describe('isAuthenticationError', () => {
    it('should return true for AuthenticationError', () => {
      const error = new AuthenticationError('Test')
      expect(isAuthenticationError(error)).toBe(true)
    })

    it('should return false for other errors', () => {
      expect(isAuthenticationError(new ValidationError('Test'))).toBe(false)
      expect(isAuthenticationError(new Error('Test'))).toBe(false)
    })
  })

  describe('isValidationError', () => {
    it('should return true for ValidationError', () => {
      const error = new ValidationError('Test')
      expect(isValidationError(error)).toBe(true)
    })

    it('should return false for other errors', () => {
      expect(isValidationError(new AuthenticationError('Test'))).toBe(false)
      expect(isValidationError(new Error('Test'))).toBe(false)
    })
  })

  describe('isNotFoundError', () => {
    it('should return true for NotFoundError', () => {
      const error = new NotFoundError('Test')
      expect(isNotFoundError(error)).toBe(true)
    })

    it('should return false for other errors', () => {
      expect(isNotFoundError(new AuthenticationError('Test'))).toBe(false)
      expect(isNotFoundError(new Error('Test'))).toBe(false)
    })
  })
})
