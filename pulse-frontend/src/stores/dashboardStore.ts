import { create } from 'zustand'
import type { DashboardFilter, TimeRange, Node } from './types'

// ============== Types ==============
export interface DashboardState {
  filters: DashboardFilter
  timeRange: TimeRange
  refreshInterval: number // seconds
  autoRefresh: boolean
  top5AbnormalNodes: Node[]
}

export interface DashboardActions {
  setFilters: (filters: DashboardFilter) => void
  setTimeRange: (range: TimeRange) => void
  setRefreshInterval: (interval: number) => void
  toggleAutoRefresh: () => void
  setTop5AbnormalNodes: (nodes: Node[]) => void
}

type DashboardStore = DashboardState & DashboardActions

// ============== Default Values ==============
const defaultFilters: DashboardFilter = {
  region: null,
  status: 'all',
  searchQuery: '',
}

// ============== Store ==============
export const useDashboardStore = create<DashboardStore>((set) => ({
  // State
  filters: defaultFilters,
  timeRange: '24h',
  refreshInterval: 5, // 5 seconds
  autoRefresh: true,
  top5AbnormalNodes: [],

  // Actions
  setFilters: (filters: DashboardFilter) => {
    set({ filters })
  },

  setTimeRange: (range: TimeRange) => {
    set({ timeRange: range })
  },

  setRefreshInterval: (interval: number) => {
    set({ refreshInterval: interval })
  },

  toggleAutoRefresh: () => {
    set((state) => ({ autoRefresh: !state.autoRefresh }))
  },

  setTop5AbnormalNodes: (nodes: Node[]) => {
    set({ top5AbnormalNodes: nodes })
  },
}))
