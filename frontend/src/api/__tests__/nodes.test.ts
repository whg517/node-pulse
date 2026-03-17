/**
 * Tests for nodes API functions
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  fetchNodes,
  fetchNode,
  createNode,
  updateNode,
  deleteNode,
  fetchNodeStatus,
} from '../nodes'
import { ValidationError, NotFoundError } from '../errors'

describe('nodes API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('fetchNode', () => {
    it('fetches a single node by ID', async () => {
      const mockNode = {
        id: 'node-1',
        name: 'Node 1',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['production'],
        status: 'online' as const,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      const mockResponse = { data: mockNode }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )
      vi.stubGlobal('fetch', mockFetch)

      const result = await fetchNode('node-1')

      const fetchCall = (fetch as unknown as { mock: { calls: string[][] } }).mock.calls[0]
      expect(fetchCall[0]).toContain('/api/v1/nodes/node-1')
      expect(result).toEqual(mockResponse)

      vi.unstubAllGlobals()
    })
  })

  describe('fetchNodes', () => {
    it('should fetch all nodes successfully', async () => {
      const mockResponse = {
        data: [
          {
            id: 'node-1',
            name: 'Node 1',
            ip: '192.168.1.1',
            region: 'us-east',
            tags: ['production'],
            status: 'online' as const,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        ],
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await fetchNodes()

      const fetchCall = (fetch as any).mock.calls[0]
      expect(fetchCall[0]).toContain('/api/v1/nodes')
      expect(fetchCall[1]).toMatchObject({
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      })
      expect(result).toEqual(mockResponse)
      expect(result.data).toHaveLength(1)

      vi.unstubAllGlobals()
    })

    it('should throw error on fetch failure', async () => {
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ code: 'ERR_AUTHENTICATION', message: 'Unauthorized' }),
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(fetchNodes()).rejects.toThrow()

      vi.unstubAllGlobals()
    })
  })

  describe('createNode', () => {
    it('should create a new node successfully', async () => {
      const newNode = {
        name: 'New Node',
        ip: '192.168.1.100',
        region: 'us-west',
        tags: ['test'],
      }
      const mockResponse = {
        data: {
          id: 'node-2',
          ...newNode,
          status: 'offline' as const,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await createNode(newNode)

      const fetchCall = (fetch as any).mock.calls[0]
      expect(fetchCall[0]).toContain('/api/v1/nodes')
      expect(fetchCall[1]).toEqual(
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(newNode),
        })
      )
      expect(result.data.name).toBe('New Node')

      vi.unstubAllGlobals()
    })

    it('should throw ValidationError for invalid node data', async () => {
      const invalidNode = {
        name: '',
        ip: 'invalid',
        region: '',
        tags: [],
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 400,
          json: async () => ({
            code: 'ERR_VALIDATION',
            message: 'Invalid node data',
            details: { field: 'ip' },
          }),
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(createNode(invalidNode)).rejects.toThrow(ValidationError)

      vi.unstubAllGlobals()
    })
  })

  describe('updateNode', () => {
    it('should update a node successfully', async () => {
      const updates = {
        name: 'Updated Node',
        tags: ['production', 'critical'],
      }
      const mockResponse = {
        data: {
          id: 'node-1',
          name: 'Updated Node',
          ip: '192.168.1.1',
          region: 'us-east',
          tags: ['production', 'critical'],
          status: 'online' as const,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T01:00:00Z',
        },
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await updateNode('node-1', updates)

      const fetchCall = (fetch as any).mock.calls[0]
      expect(fetchCall[0]).toContain('/api/v1/nodes/node-1')
      expect(fetchCall[1]).toEqual(
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(updates),
        })
      )
      expect(result.data.name).toBe('Updated Node')

      vi.unstubAllGlobals()
    })

    it('should throw NotFoundError for non-existent node', async () => {
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 404,
          json: async () => ({
            code: 'ERR_NOT_FOUND',
            message: 'Node not found',
          }),
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      await expect(updateNode('invalid-id', { name: 'Test' })).rejects.toThrow(
        NotFoundError
      )

      vi.unstubAllGlobals()
    })
  })

  describe('deleteNode', () => {
    it('should delete a node successfully', async () => {
      const mockResponse = {
        message: 'Node deleted successfully',
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await deleteNode('node-1')

      const fetchCall = (fetch as any).mock.calls[0]
      expect(fetchCall[0]).toContain('/api/v1/nodes/node-1')
      expect(fetchCall[1]).toEqual(
        expect.objectContaining({
          method: 'DELETE',
        })
      )
      expect(result.message).toBe('Node deleted successfully')

      vi.unstubAllGlobals()
    })
  })

  describe('fetchNodeStatus', () => {
    it('should fetch node status successfully', async () => {
      const mockResponse = {
        data: {
          status: 'online',
          last_heartbeat: '2024-01-01T12:00:00Z',
        },
      }
      const mockFetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        } as Response)
      )

      vi.stubGlobal('fetch', mockFetch)

      const result = await fetchNodeStatus('node-1')

      const fetchCall = (fetch as any).mock.calls[0]
      expect(fetchCall[0]).toContain('/api/v1/nodes/node-1/status')
      expect(fetchCall[1]).toMatchObject({
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      })
      expect(result.data.status).toBe('online')
      expect(result.data.last_heartbeat).toBe('2024-01-01T12:00:00Z')

      vi.unstubAllGlobals()
    })
  })
})
