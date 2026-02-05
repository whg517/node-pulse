/**
 * Tests for API client functionality
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiClient } from '../client'
import { AuthenticationError, ValidationError, NotFoundError } from '../errors'

describe('apiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('successful requests', () => {
    it('should make GET request with correct headers and credentials', async () => {
      const mockResponse = { data: [{ id: '1', name: 'Test' }] }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await apiClient<typeof mockResponse>('/api/v1/test')

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/test'),
        expect.objectContaining({
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
        })
      )
      expect(result).toEqual(mockResponse)

      vi.unstubAllGlobals()
    })

    it('should make POST request with body', async () => {
      const mockResponse = { data: { id: '1' } }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const body = { name: 'Test' }
      const result = await apiClient<typeof mockResponse>('/api/v1/test', {
        method: 'POST',
        body: JSON.stringify(body),
      })

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/test'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(body),
        })
      )
      expect(result).toEqual(mockResponse)

      vi.unstubAllGlobals()
    })

    it('should merge custom headers with default headers', async () => {
      const mockResponse = { data: 'test' }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await apiClient<typeof mockResponse>('/api/v1/test', {
        headers: {
          'X-Custom-Header': 'custom-value',
        },
      })

      expect(fetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: {
            'Content-Type': 'application/json',
            'X-Custom-Header': 'custom-value',
          },
        })
      )

      vi.unstubAllGlobals()
    })
  })

  describe('error handling', () => {
    it('should throw ValidationError for 400 status', async () => {
      const errorResponse = {
        code: 'ERR_VALIDATION',
        message: 'Invalid input',
        details: { field: 'email' },
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 400,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(apiClient('/api/v1/test')).rejects.toThrow(ValidationError)

      try {
        await apiClient('/api/v1/test')
        expect.fail('Should have thrown ValidationError')
      } catch (error) {
        expect(error).toBeInstanceOf(ValidationError)
        const validationError = error as ValidationError
        expect(validationError.message).toBe('Invalid input')
        expect(validationError.code).toBe('ERR_VALIDATION')
        expect(validationError.details).toEqual({ field: 'email' })
        expect(validationError.status).toBe(400)
      }

      vi.unstubAllGlobals()
    })

    it('should throw AuthenticationError for 401 status', async () => {
      const errorResponse = {
        code: 'ERR_AUTHENTICATION',
        message: 'Unauthorized',
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(apiClient('/api/v1/test')).rejects.toThrow(AuthenticationError)

      vi.unstubAllGlobals()
    })

    it('should throw NotFoundError for 404 status', async () => {
      const errorResponse = {
        code: 'ERR_NOT_FOUND',
        message: 'Resource not found',
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 404,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(apiClient('/api/v1/test')).rejects.toThrow(NotFoundError)

      vi.unstubAllGlobals()
    })

    it('should throw ApiError for unknown status codes', async () => {
      const errorResponse = {
        code: 'ERR_INTERNAL',
        message: 'Internal server error',
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 500,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      try {
        await apiClient('/api/v1/test')
        expect.fail('Should have thrown ApiError')
      } catch (error) {
        expect(error).toBeInstanceOf(Error)
        expect(error).toHaveProperty('status', 500)
      }

      vi.unstubAllGlobals()
    })

    it('should handle non-JSON error responses', async () => {
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 500,
          statusText: 'Internal Server Error',
          headers: new Headers(),
          json: async () => {
            throw new Error('Invalid JSON')
          },
          type: 'basic',
          url: 'http://localhost:8000/api/v1/test',
          redirected: false,
          clone: () => ({} as Response),
          body: null,
          bodyUsed: false,
          arrayBuffer: async () => new ArrayBuffer(0),
          blob: async () => new Blob(),
          formData: async () => new FormData(),
          text: async () => '',
          bytes: async () => new Uint8Array(),
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      try {
        await apiClient('/api/v1/test')
        expect.fail('Should have thrown ApiError')
      } catch (error) {
        expect(error).toBeInstanceOf(Error)
        const apiError = error as Error
        expect(apiError.message).toBe('Internal Server Error')
      }

      vi.unstubAllGlobals()
    })
  })

  describe('URL construction', () => {
    it('should use API_BASE_URL from constants', async () => {
      const mockResponse = { data: 'test' }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await apiClient<typeof mockResponse>('/api/v1/test')

      const fetchCall = (fetch as any).mock.calls[0]
      const url = fetchCall[0]

      expect(url).toContain('http://localhost:8080')
      expect(url).toContain('/api/v1/test')

      vi.unstubAllGlobals()
    })
  })
})
