import { create } from 'zustand'
import { fetchNodes as fetchNodesApi } from '../api/nodes'
import type { Node, NodeStatus } from './types'

// ============== Types ==============
export interface NodesState {
  nodes: Node[]
  selectedNode: Node | null
  nodeStatuses: Record<string, NodeStatus>
}

export interface NodesActions {
  setNodes: (nodes: Node[]) => void
  setSelectedNode: (node: Node | null) => void
  addNode: (node: Node) => void
  updateNode: (id: string, updates: Partial<Node>) => void
  removeNode: (id: string) => void
  setNodeStatus: (nodeId: string, status: NodeStatus) => void
  fetchNodes: () => Promise<void>
}

type NodesStore = NodesState & NodesActions

// ============== Store ==============
export const useNodesStore = create<NodesStore>((set, get) => ({
  // State
  nodes: [],
  selectedNode: null,
  nodeStatuses: {},

  // Actions
  setNodes: (nodes: Node[]) => {
    set({ nodes })
  },

  setSelectedNode: (node: Node | null) => {
    set({ selectedNode: node })
  },

  addNode: (node: Node) => {
    set((state) => ({
      nodes: [...state.nodes, node],
    }))
  },

  updateNode: (id: string, updates: Partial<Node>) => {
    set((state) => ({
      nodes: state.nodes.map((node) =>
        node.id === id ? { ...node, ...updates } : node
      ),
    }))
  },

  removeNode: (id: string) => {
    set((state) => ({
      nodes: state.nodes.filter((node) => node.id !== id),
      selectedNode:
        state.selectedNode?.id === id ? null : state.selectedNode,
    }))
  },

  setNodeStatus: (nodeId: string, status: NodeStatus) => {
    set((state) => ({
      nodeStatuses: {
        ...state.nodeStatuses,
        [nodeId]: status,
      },
    }))
  },

  fetchNodes: async () => {
    try {
      const response = await fetchNodesApi()

      const nodes: Node[] = response.data.map((node) => ({
        id: node.id,
        name: node.name,
        ip: node.ip,
        region: node.region,
        tags: node.tags || [],
        status: node.status || 'offline',
      }))

      set({ nodes })
    } catch (error) {
      console.error('Failed to fetch nodes:', error)
      throw error
    }
  },
}))
