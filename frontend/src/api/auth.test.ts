import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { login, logout } from './auth'
import { AuthenticationError } from './errors'
import { SESSION_COOKIE_NAME } from './auth'

describe('auth API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
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

  describe('SESSION_COOKIE_NAME', () => {
    it('should export SESSION_COOKIE_NAME constant', () => {
      expect(SESSION_COOKIE_NAME).toBe('session_id')
    })
  })
})
