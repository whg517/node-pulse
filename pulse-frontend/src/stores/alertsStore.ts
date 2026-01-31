import { create } from 'zustand'
import { fetchAlertRules, fetchAlertRecords } from '../api/alerts'
import type { AlertRule, AlertRecord, AlertFilter } from './types'

// ============== Types ==============
export interface AlertsState {
  alertRules: AlertRule[]
  alertRecords: AlertRecord[]
  filter: AlertFilter
}

export interface AlertsActions {
  setAlertRules: (rules: AlertRule[]) => void
  setAlertRecords: (records: AlertRecord[]) => void
  addAlertRule: (rule: AlertRule) => void
  updateAlertRule: (id: string, updates: Partial<AlertRule>) => void
  removeAlertRule: (id: string) => void
  setFilter: (filter: AlertFilter) => void
  fetchAlertRules: () => Promise<void>
  fetchAlertRecords: () => Promise<void>
}

type AlertsStore = AlertsState & AlertsActions

// ============== Default Filter ==============
const defaultFilter: AlertFilter = {
  level: 'all',
  status: 'all',
  nodeId: null,
  searchQuery: '',
}

// ============== Store ==============
export const useAlertsStore = create<AlertsStore>((set, get) => ({
  // State
  alertRules: [],
  alertRecords: [],
  filter: defaultFilter,

  // Actions
  setAlertRules: (rules: AlertRule[]) => {
    set({ alertRules: rules })
  },

  setAlertRecords: (records: AlertRecord[]) => {
    set({ alertRecords: records })
  },

  addAlertRule: (rule: AlertRule) => {
    set((state) => ({
      alertRules: [...state.alertRules, rule],
    }))
  },

  updateAlertRule: (id: string, updates: Partial<AlertRule>) => {
    set((state) => ({
      alertRules: state.alertRules.map((rule) =>
        rule.id === id ? { ...rule, ...updates } : rule
      ),
    }))
  },

  removeAlertRule: (id: string) => {
    set((state) => ({
      alertRules: state.alertRules.filter((rule) => rule.id !== id),
    }))
  },

  setFilter: (filter: AlertFilter) => {
    set({ filter })
  },

  fetchAlertRules: async () => {
    try {
      const response = await fetchAlertRules()

      const alertRules: AlertRule[] = response.data.map((rule) => ({
        id: rule.id,
        metric: rule.metric,
        threshold: rule.threshold,
        level: rule.level,
        nodeId: rule.node_id,
        enabled: rule.enabled,
      }))

      set({ alertRules })
    } catch (error) {
      console.error('Failed to fetch alert rules:', error)
      throw error
    }
  },

  fetchAlertRecords: async () => {
    try {
      const response = await fetchAlertRecords()

      const alertRecords: AlertRecord[] = response.data.map((record) => ({
        id: record.id,
        nodeId: record.node_id,
        metric: record.metric,
        level: record.level,
        status: record.status,
        timestamp: record.timestamp,
      }))

      set({ alertRecords })
    } catch (error) {
      console.error('Failed to fetch alert records:', error)
      throw error
    }
  },
}))
