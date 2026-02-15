import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { getSessions, deleteSession, getSessionInfo } from '../sessions'
import { NotFoundError } from '../errors'

describe('sessions API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  describe('getSessions', () => {
    it('should successfully fetch sessions list', async () => {
      const mockResponse = {
        data: {
          sessions: [
            {
              id: 'session-1',
              user_id: 'user-123',
              created_at: '2024-01-01T00:00:00Z',
              last_used_at: '2024-01-02T00:00:00Z',
              expires_at: '2024-01-08T00:00:00Z',
              ip_address: '192.168.1.1',
              user_agent: 'Mozilla/5.0',
              is_current: true,
            },
            {
              id: 'session-2',
              user_id: 'user-123',
              created_at: '2024-01-01T10:00:00Z',
              last_used_at: '2024-01-01T10:00:00Z',
              expires_at: '2024-01-08T10:00:00Z',
              ip_address: '10.0.0.1',
              user_agent: 'Chrome/120.0',
              is_current: false,
            },
          ],
          total: 2,
        },
        message: 'Sessions retrieved successfully',
        timestamp: '2024-01-02T00:00:00Z',
      }

      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      })

      vi.stubGlobal('fetch', mockFetch)

      const result = await getSessions()

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/auth/sessions'),
        expect.objectContaining({
          method: 'GET',
          credentials: 'include',
        })
      )
      expect(result.data.sessions).toHaveLength(2)
      expect(result.data.total).toBe(2)
      expect(result.data.sessions[0].is_current).toBe(true)
    })

    it('should handle empty sessions list', async () => {
      const mockResponse = {
        data: {
          sessions: [],
          total: 0,
        },
        message: 'No sessions found',
        timestamp: '2024-01-02T00:00:00Z',
      }

      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      })

      vi.stubGlobal('fetch', mockFetch)

      const result = await getSessions()

      expect(result.data.sessions).toHaveLength(0)
      expect(result.data.total).toBe(0)
    })
  })

  describe('deleteSession', () => {
    it('should successfully delete a session', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ message: 'Session revoked' }),
      })

      vi.stubGlobal('fetch', mockFetch)

      await deleteSession('session-123')

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/auth/sessions/session-123'),
        expect.objectContaining({
          method: 'DELETE',
          credentials: 'include',
        })
      )
    })

    it('should throw NotFoundError for non-existent session', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({
          code: 'ERR_NOT_FOUND',
          message: 'Session not found',
        }),
      })

      vi.stubGlobal('fetch', mockFetch)

      await expect(deleteSession('non-existent')).rejects.toThrow(NotFoundError)
    })

    it('should throw error for unauthorized access', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({
          code: 'ERR_AUTHENTICATION',
          message: 'Unauthorized',
        }),
      })

      vi.stubGlobal('fetch', mockFetch)

      await expect(deleteSession('session-123')).rejects.toThrow()
    })
  })

  describe('getSessionInfo', () => {
    it('should successfully fetch session info', async () => {
      const mockResponse = {
        data: {
          current_session_id: 'session-current',
          active_sessions_count: 3,
        },
        message: 'Session info retrieved',
        timestamp: '2024-01-02T00:00:00Z',
      }

      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      })

      vi.stubGlobal('fetch', mockFetch)

      const result = await getSessionInfo()

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/auth/sessions/info'),
        expect.objectContaining({
          method: 'GET',
          credentials: 'include',
        })
      )
      expect(result.data.current_session_id).toBe('session-current')
      expect(result.data.active_sessions_count).toBe(3)
    })

    it('should handle unauthorized access', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({
          code: 'ERR_AUTHENTICATION',
          message: 'Unauthorized',
        }),
      })

      vi.stubGlobal('fetch', mockFetch)

      await expect(getSessionInfo()).rejects.toThrow()
    })
  })
})
