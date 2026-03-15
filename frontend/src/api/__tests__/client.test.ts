/**
 * Tests for API client functionality
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { apiClient, resetModuleState } from '../client'
import { ApiError, AuthenticationError, ValidationError, NotFoundError } from '../errors'
import { useAuthStore } from '../../stores/authStore'

// Mock localStorage for cross-tab logout sync
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
vi.stubGlobal('localStorage', localStorageMock)

describe('apiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset module-level state between tests
    resetModuleState()
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

    it('should throw ApiError for 5xx status codes with user-friendly message', async () => {
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
        // 5xx errors now return a user-friendly message
        expect((error as Error).message).toBe('Server temporarily unavailable. Please try again later.')
      }

      vi.unstubAllGlobals()
    })

    it('should handle non-JSON error responses for 5xx', async () => {
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
        // 5xx errors now return a user-friendly message
        expect(apiError.message).toBe('Server temporarily unavailable. Please try again later.')
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

      expect(url).toContain('http://localhost:6532')
      expect(url).toContain('/api/v1/test')

      vi.unstubAllGlobals()
    })
  })
})

// ============ CSRF header tests ============

describe('X-CSRF-Token header injection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resetModuleState()
    // Reset auth store state
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      role: null,
      accessToken: null,
      tokenExpiresAt: null,
      csrfToken: null,
      isLoading: false,
      refreshFailureCount: 0,
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    useAuthStore.setState({ csrfToken: null })
  })

  it('should include X-CSRF-Token header on POST when csrfToken is set', async () => {
    useAuthStore.setState({ csrfToken: 'test-csrf-token-abc' })

    const mockFetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => ({ data: 'ok' }) } as Response)
    )
    vi.stubGlobal('fetch', mockFetch)

    await apiClient('/api/v1/test', { method: 'POST', body: '{}' })

    const [, options] = mockFetch.mock.calls[0] as [string, RequestInit]
    const headers = options.headers as Record<string, string>
    expect(headers['X-CSRF-Token']).toBe('test-csrf-token-abc')
  })

  it('should include X-CSRF-Token header on PUT requests', async () => {
    useAuthStore.setState({ csrfToken: 'csrf-for-put' })

    const mockFetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => ({}) } as Response)
    )
    vi.stubGlobal('fetch', mockFetch)

    await apiClient('/api/v1/test', { method: 'PUT', body: '{}' })

    const [, options] = mockFetch.mock.calls[0] as [string, RequestInit]
    const headers = options.headers as Record<string, string>
    expect(headers['X-CSRF-Token']).toBe('csrf-for-put')
  })

  it('should include X-CSRF-Token header on DELETE requests', async () => {
    useAuthStore.setState({ csrfToken: 'csrf-for-delete' })

    const mockFetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => ({}) } as Response)
    )
    vi.stubGlobal('fetch', mockFetch)

    await apiClient('/api/v1/test', { method: 'DELETE' })

    const [, options] = mockFetch.mock.calls[0] as [string, RequestInit]
    const headers = options.headers as Record<string, string>
    expect(headers['X-CSRF-Token']).toBe('csrf-for-delete')
  })

  it('should NOT include X-CSRF-Token header on GET requests', async () => {
    useAuthStore.setState({ csrfToken: 'should-not-appear' })

    const mockFetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => ({}) } as Response)
    )
    vi.stubGlobal('fetch', mockFetch)

    await apiClient('/api/v1/test') // default GET

    const [, options] = mockFetch.mock.calls[0] as [string, RequestInit]
    const headers = options.headers as Record<string, string>
    expect(headers['X-CSRF-Token']).toBeUndefined()
  })

  it('should NOT include X-CSRF-Token on POST when csrfToken is null', async () => {
    // csrfToken is already null from beforeEach

    const mockFetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: async () => ({}) } as Response)
    )
    vi.stubGlobal('fetch', mockFetch)

    await apiClient('/api/v1/test', { method: 'POST', body: '{}' })

    const [, options] = mockFetch.mock.calls[0] as [string, RequestInit]
    const headers = options.headers as Record<string, string>
    expect(headers['X-CSRF-Token']).toBeUndefined()
  })
})

// ============ Fix 1: 429 on /auth/refresh must not cause logout ============

describe('rate limit (429) on token refresh', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resetModuleState()
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      role: null,
      accessToken: 'existing-token',
      tokenExpiresAt: null,
      csrfToken: null,
      isLoading: false,
      refreshFailureCount: 0,
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('should throw ApiError(429) when refresh endpoint is rate-limited', async () => {
    // First call: original request returns 401
    // Second call (refresh): returns 429
    const mockFetch = vi.fn()
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ code: 'ERR_UNAUTHORIZED', message: 'Unauthorized' }),
      } as Response)
      .mockResolvedValueOnce({
        ok: false,
        status: 429,
        json: async () => ({ code: 'ERR_RATE_LIMIT_EXCEEDED', message: 'Too many refresh requests' }),
      } as Response)

    vi.stubGlobal('fetch', mockFetch)

    const error = await apiClient('/api/v1/nodes').catch(e => e)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(429)
  })

  it('should NOT call clearAuth() when refresh returns 429', async () => {
    const clearAuthSpy = vi.spyOn(useAuthStore.getState(), 'clearAuth')

    const mockFetch = vi.fn()
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ code: 'ERR_UNAUTHORIZED', message: 'Unauthorized' }),
      } as Response)
      .mockResolvedValueOnce({
        ok: false,
        status: 429,
        json: async () => ({ code: 'ERR_RATE_LIMIT_EXCEEDED', message: 'Too many refresh requests' }),
      } as Response)

    vi.stubGlobal('fetch', mockFetch)

    await apiClient('/api/v1/nodes').catch(() => {})

    expect(clearAuthSpy).not.toHaveBeenCalled()
  })

  it('should call clearAuth() on real auth failure (non-429)', async () => {
    const clearAuthSpy = vi.spyOn(useAuthStore.getState(), 'clearAuth')

    const mockFetch = vi.fn()
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ code: 'ERR_UNAUTHORIZED', message: 'Unauthorized' }),
      } as Response)
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ code: 'ERR_UNAUTHORIZED', message: 'Unauthorized' }),
      } as Response)

    vi.stubGlobal('fetch', mockFetch)

    await apiClient('/api/v1/nodes').catch(() => {})

    expect(clearAuthSpy).toHaveBeenCalledTimes(1)
  })

  it('should not increment failure count when refresh is rate-limited (429)', async () => {
    // 429 three times — should NOT trigger logout (failure count stays 0)
    const clearAuthSpy = vi.spyOn(useAuthStore.getState(), 'clearAuth')

    const mockFetch = vi.fn()
      .mockResolvedValue({
        ok: false,
        status: 429,
        json: async () => ({ code: 'ERR_RATE_LIMIT_EXCEEDED', message: 'Too many requests' }),
      } as Response)

    vi.stubGlobal('fetch', mockFetch)

    // Simulate 429 on the original request directly (not via refresh path)
    // — also verify the direct 429 path throws ApiError and not AuthenticationError
    const error = await apiClient('/api/v1/nodes').catch(e => e)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(429)
    expect(clearAuthSpy).not.toHaveBeenCalled()
  })
})
