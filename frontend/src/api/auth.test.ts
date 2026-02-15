import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { login, logout, refreshToken } from './auth'
import { AuthenticationError } from './errors'
import { resetModuleState } from './client'

// Mock localStorage for cross-tab logout sync
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
vi.stubGlobal('localStorage', localStorageMock)

describe('auth API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset module-level state between tests
    resetModuleState()
  })

  afterEach(() => {
    document.cookie = ''
  })

  describe('login', () => {
    it('should successfully login with valid credentials', async () => {
      const mockResponse = {
        data: {
          user_id: '123',
          username: 'admin',
          role: 'admin',
          access_token: 'test-access-token',
        },
        message: 'Login successful',
        timestamp: '2026-01-26T10:00:00Z',
      }

      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await login({ username: 'admin', password: 'password123' })

      const fetchCalls = (fetch as any).mock.calls
      const fetchCall = fetchCalls[0]
      expect(fetchCall[0]).toContain('/api/v1/auth/login')
      expect(fetchCall[1]).toEqual(
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ username: 'admin', password: 'password123' }),
        })
      )
      expect(result).toEqual(mockResponse)

      vi.unstubAllGlobals()
    })

    it('should throw AuthenticationError for invalid credentials', async () => {
      const errorResponse = {
        code: 'ERR_INVALID_CREDENTIALS',
        message: 'Invalid username or password',
        details: {
          failed_attempts: 3,
          remaining_attempts: 2,
        },
      }

      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(login({ username: 'admin', password: 'wrong' })).rejects.toThrow(AuthenticationError)

      vi.unstubAllGlobals()
    })

    it('should throw AuthenticationError for locked account', async () => {
      const errorResponse = {
        code: 'ERR_AUTHENTICATION',
        message: 'Account locked due to too many failed login attempts',
        details: {
          locked_until: '2026-01-26T11:00:00Z',
          lock_duration_minutes: 10,
        },
      }

      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(login({ username: 'admin', password: 'wrong' })).rejects.toThrow(AuthenticationError)

      vi.unstubAllGlobals()
    })

    it('should throw AuthenticationError for rate limit', async () => {
      const errorResponse = {
        code: 'ERR_RATE_LIMIT',
        message: 'Too many login attempts. Please try again later.',
      }

      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 429,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(login({ username: 'admin', password: 'wrong' })).rejects.toThrow()

      vi.unstubAllGlobals()
    })

    it('should include error code in thrown AuthenticationError', async () => {
      const errorResponse = {
        code: 'ERR_AUTHENTICATION',
        message: 'Invalid username or password',
        details: { failed_attempts: 3 },
      }

      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      try {
        await login({ username: 'admin', password: 'wrong' })
        expect.fail('Should have thrown an error')
      } catch (error) {
        expect(error).toBeInstanceOf(AuthenticationError)
        const authError = error as AuthenticationError
        expect(authError.code).toBe('ERR_AUTHENTICATION')
        expect(authError.message).toBe('Invalid username or password')
        expect(authError.details).toEqual({ failed_attempts: 3 })
        expect(authError.status).toBe(401)
      }

      vi.unstubAllGlobals()
    })

    it('should include account locked details in thrown AuthenticationError', async () => {
      const errorResponse = {
        code: 'ERR_AUTHENTICATION',
        message: 'Account locked',
        details: {
          locked_until: '2026-01-26T12:00:00Z',
          lock_duration_minutes: 10,
        },
      }

      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
          json: async () => errorResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      try {
        await login({ username: 'admin', password: 'wrong' })
        expect.fail('Should have thrown an error')
      } catch (error) {
        expect(error).toBeInstanceOf(AuthenticationError)
        const authError = error as AuthenticationError
        expect(authError.code).toBe('ERR_AUTHENTICATION')
        expect(authError.details).toEqual({
          locked_until: '2026-01-26T12:00:00Z',
          lock_duration_minutes: 10,
        })
        expect(authError.status).toBe(401)
      }

      vi.unstubAllGlobals()
    })
  })

  describe('logout', () => {
    it('should successfully logout', async () => {
      const mockResponse = {
        message: 'Logout successful',
        timestamp: '2026-01-26T10:30:00Z',
      }

      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await logout()

      const fetchCalls = (fetch as any).mock.calls
      const fetchCall = fetchCalls[0]
      expect(fetchCall[0]).toContain('/api/v1/auth/logout')
      expect(fetchCall[1]).toEqual(
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
        })
      )
      expect(result).toEqual(mockResponse)

      vi.unstubAllGlobals()
    })

    it('should throw AuthenticationError on logout failure', async () => {
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(logout()).rejects.toThrow(AuthenticationError)

      vi.unstubAllGlobals()
    })
  })

  describe('refreshToken', () => {
    it('should successfully refresh token', async () => {
      const mockResponse = {
        data: {
          access_token: 'new_access_token',
        },
        message: 'Token refreshed successfully',
        timestamp: '2024-02-05T00:00:00Z',
      }

      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      })

      vi.stubGlobal('fetch', mockFetch)

      const result = await refreshToken()

      expect(result.data.access_token).toBe('new_access_token')
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/auth/refresh'),
        expect.objectContaining({
          method: 'POST',
          credentials: 'include',
        })
      )

      vi.unstubAllGlobals()
    })

    it('should handle refresh token errors', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({
          code: 'ERR_INVALID_REFRESH_TOKEN',
          message: 'Invalid or expired refresh token',
        }),
      })

      vi.stubGlobal('fetch', mockFetch)

      await expect(refreshToken()).rejects.toThrow(AuthenticationError)

      vi.unstubAllGlobals()
    })
  })
})
