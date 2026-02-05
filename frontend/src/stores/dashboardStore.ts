import { create } from 'zustand'
import type {
  DashboardFilter,
  TimeRange,
  Node,
  ComparisonState,
  ExtendedTimeRange,
  GroupByType,
  MetricType,
} from './types'

// ============== Types ==============
export interface DashboardState {
  filters: DashboardFilter
  timeRange: TimeRange
  refreshInterval: number // seconds
  autoRefresh: boolean
  top5AbnormalNodes: Node[]
  comparison: ComparisonState
}

export interface DashboardActions {
  setFilters: (filters: DashboardFilter) => void
  setTimeRange: (range: TimeRange) => void
  setRefreshInterval: (interval: number) => void
  toggleAutoRefresh: () => void
  setTop5AbnormalNodes: (nodes: Node[]) => void

  // Comparison actions
  setComparisonNodeIds: (nodeIds: string[]) => void
  setComparisonMetrics: (metrics: MetricType[]) => void
  setComparisonTimeRange: (range: ExtendedTimeRange) => void
  setComparisonCustomTimeRange: (range: { start: string; end: string }) => void
  setComparisonGroupBy: (groupBy: GroupByType) => void
  resetComparison: () => void
}

type DashboardStore = DashboardState & DashboardActions

// ============== Default Values ==============
const defaultFilters: DashboardFilter = {
  region: null,
  status: 'all',
  searchQuery: '',
}

const defaultComparisonState: ComparisonState = {
  selectedNodeIds: [],
  selectedMetrics: ['latency_ms'],
  timeRange: '24h',
  groupBy: 'none',
}

// ============== Store ==============
export const useDashboardStore = create<DashboardStore>((set) => ({
  // State
  filters: defaultFilters,
  timeRange: '24h',
  refreshInterval: 5, // 5 seconds
  autoRefresh: true,
  top5AbnormalNodes: [],
  comparison: defaultComparisonState,

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

  // Comparison actions
  setComparisonNodeIds: (nodeIds: string[]) => {
    set((state) => ({
      comparison: { ...state.comparison, selectedNodeIds: nodeIds },
    }))
  },

  setComparisonMetrics: (metrics: MetricType[]) => {
    set((state) => ({
      comparison: { ...state.comparison, selectedMetrics: metrics },
    }))
  },

  setComparisonTimeRange: (range: ExtendedTimeRange) => {
    set((state) => ({
      comparison: {
        ...state.comparison,
        timeRange: range,
        customTimeRange: range !== 'custom' ? undefined : state.comparison.customTimeRange,
      },
    }))
  },

  setComparisonCustomTimeRange: (range: { start: string; end: string }) => {
    set((state) => ({
      comparison: { ...state.comparison, customTimeRange: range },
    }))
  },

  setComparisonGroupBy: (groupBy: GroupByType) => {
    set((state) => ({
      comparison: { ...state.comparison, groupBy },
    }))
  },

  resetComparison: () => {
    set({ comparison: defaultComparisonState })
  },
}))
