import { create } from 'zustand'
import { fetchAlertRules, fetchAlertRecords, createAlertRule, deleteAlertRule, updateAlertRule as apiUpdateAlertRule } from '../api/alerts'
import type { CreateAlertRuleRequest, UpdateAlertRuleRequest } from '../api/types'
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
  addAlertRule: (request: CreateAlertRuleRequest) => Promise<void>
  updateAlertRule: (id: string, updates: UpdateAlertRuleRequest) => Promise<void>
  removeAlertRule: (id: string) => Promise<void>
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
export const useAlertsStore = create<AlertsStore>((set) => ({
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

  addAlertRule: async (request: CreateAlertRuleRequest) => {
    const response = await createAlertRule(request)
    const rule: AlertRule = {
      id: response.data.id,
      metric: response.data.metric,
      threshold: response.data.threshold,
      level: response.data.level,
      nodeId: response.data.node_id,
      enabled: response.data.enabled,
    }
    set((state) => ({
      alertRules: [...state.alertRules, rule],
    }))
  },

  updateAlertRule: async (id: string, updates: UpdateAlertRuleRequest) => {
    const response = await apiUpdateAlertRule(id, updates)
    const updated: AlertRule = {
      id: response.data.id,
      metric: response.data.metric,
      threshold: response.data.threshold,
      level: response.data.level,
      nodeId: response.data.node_id,
      enabled: response.data.enabled,
    }
    set((state) => ({
      alertRules: state.alertRules.map((rule) =>
        rule.id === id ? updated : rule
      ),
    }))
  },

  removeAlertRule: async (id: string) => {
    await deleteAlertRule(id)
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

      const alertRules: AlertRule[] = (response.data.alerts || []).map((rule) => ({
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
        timestamp: record.created_at,
      }))

      set({ alertRecords })
    } catch (error) {
      console.error('Failed to fetch alert records:', error)
      throw error
    }
  },
}))
