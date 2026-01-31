import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useNodesStore } from '../nodesStore'
import * as nodesApi from '../../api/nodes'

// Mock the nodes API
vi.mock('../../api/nodes', () => ({
  fetchNodes: vi.fn(),
}))

describe('useNodesStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useNodesStore.setState({
      nodes: [],
      selectedNode: null,
      nodeStatuses: {},
    })
    vi.clearAllMocks()
  })

  it('should have initial state', () => {
    const { result } = renderHook(() => useNodesStore())

    expect(result.current.nodes).toEqual([])
    expect(result.current.selectedNode).toBeNull()
    expect(result.current.nodeStatuses).toEqual({})
  })

  it('should set nodes', () => {
    const { result } = renderHook(() => useNodesStore())

    const mockNodes = [
      {
        id: 'node-1',
        name: 'Node 1',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['tag1', 'tag2'],
        status: 'online' as const,
      },
      {
        id: 'node-2',
        name: 'Node 2',
        ip: '192.168.1.2',
        region: 'us-west',
        tags: ['tag3'],
        status: 'offline' as const,
      },
    ]

    act(() => {
      result.current.setNodes(mockNodes)
    })

    expect(result.current.nodes).toEqual(mockNodes)
  })

  it('should set selected node', () => {
    const { result } = renderHook(() => useNodesStore())

    const mockNode = {
      id: 'node-1',
      name: 'Node 1',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: ['tag1'],
      status: 'online' as const,
    }

    act(() => {
      result.current.setSelectedNode(mockNode)
    })

    expect(result.current.selectedNode).toEqual(mockNode)
  })

  it('should add node', () => {
    const { result } = renderHook(() => useNodesStore())

    const mockNode = {
      id: 'node-1',
      name: 'Node 1',
      ip: '192.168.1.1',
      region: 'us-east',
      tags: ['tag1'],
      status: 'online' as const,
    }

    act(() => {
      result.current.addNode(mockNode)
    })

    expect(result.current.nodes).toHaveLength(1)
    expect(result.current.nodes[0]).toEqual(mockNode)
  })

  it('should update node', () => {
    const { result } = renderHook(() => useNodesStore())

    // Set initial nodes
    const initialNodes = [
      {
        id: 'node-1',
        name: 'Node 1',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['tag1'],
        status: 'online' as const,
      },
    ]

    act(() => {
      result.current.setNodes(initialNodes)
    })

    // Update node
    act(() => {
      result.current.updateNode('node-1', { status: 'offline' as const, name: 'Updated Node' })
    })

    expect(result.current.nodes[0].status).toBe('offline')
    expect(result.current.nodes[0].name).toBe('Updated Node')
    expect(result.current.nodes[0].ip).toBe('192.168.1.1') // Other fields unchanged
  })

  it('should remove node', () => {
    const { result } = renderHook(() => useNodesStore())

    // Set initial nodes
    const initialNodes = [
      {
        id: 'node-1',
        name: 'Node 1',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['tag1'],
        status: 'online' as const,
      },
      {
        id: 'node-2',
        name: 'Node 2',
        ip: '192.168.1.2',
        region: 'us-west',
        tags: ['tag2'],
        status: 'offline' as const,
      },
    ]

    act(() => {
      result.current.setNodes(initialNodes)
      result.current.setSelectedNode(initialNodes[0])
    })

    // Remove node
    act(() => {
      result.current.removeNode('node-1')
    })

    expect(result.current.nodes).toHaveLength(1)
    expect(result.current.nodes[0].id).toBe('node-2')
    expect(result.current.selectedNode).toBeNull() // Selected node cleared if removed
  })

  it('should set node status', () => {
    const { result } = renderHook(() => useNodesStore())

    act(() => {
      result.current.setNodeStatus('node-1', 'online')
    })

    expect(result.current.nodeStatuses['node-1']).toBe('online')
  })

  it('should fetch nodes from API', async () => {
    const mockNodesData = [
      {
        id: 'node-1',
        name: 'Node 1',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['tag1'],
        status: 'online' as const,
      },
    ]

    vi.mocked(nodesApi.fetchNodes).mockResolvedValueOnce({ data: mockNodesData })

    const { result } = renderHook(() => useNodesStore())

    await act(async () => {
      await result.current.fetchNodes()
    })

    expect(result.current.nodes).toHaveLength(1)
    expect(result.current.nodes[0].id).toBe('node-1')
    expect(nodesApi.fetchNodes).toHaveBeenCalled()
  })

  it('should handle fetch nodes error', async () => {
    vi.mocked(nodesApi.fetchNodes).mockRejectedValueOnce(new Error('Network error'))

    const { result } = renderHook(() => useNodesStore())

    await expect(async () => {
      await act(async () => {
        await result.current.fetchNodes()
      })
    }).rejects.toThrow('Network error')
  })
})
