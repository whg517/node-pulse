/**
 * Settings Store
 *
 * Zustand store for user preferences including language, theme, and timezone.
 * Persists settings to localStorage for cross-session persistence.
 */

import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import i18n from '../i18n'

// ============== Types ==============

export type ThemeMode = 'light' | 'dark' | 'system'
export type TimezoneDisplayMode = 'utc' | 'local' | 'nodeLocal' | 'multi'

export interface ReportSchedule {
  id: string
  name: string
  frequency: 'daily' | 'weekly' | 'monthly'
  time: string
  nodeIds: string[]
  format: 'pdf' | 'csv'
  enabled: boolean
  lastRun?: string
}

export interface ConfigTemplate {
  id: string
  name: string
  description?: string
  probes: Array<{
    type: 'TCP' | 'UDP'
    target: string
    port: number
    interval_seconds: number
    timeout_seconds: number
    count: number
  }>
  interval_seconds: number
  timeout_seconds: number
}

export interface AlertRoutingRule {
  id: string
  name: string
  conditions: {
    metric?: string
    severity?: string
    nodeGroup?: string
  }
  action: {
    type: 'webhook' | 'email'
    target: string
  }
  enabled: boolean
}

export interface TimezoneOption {
  value: string
  label: string
  offset: string
}

export interface SettingsState {
  language: string
  theme: ThemeMode
  timezone: string
  timezoneDisplayMode: TimezoneDisplayMode
  reportSchedules: ReportSchedule[]
  configTemplates: ConfigTemplate[]
  routingRules: AlertRoutingRule[]
}

export interface SettingsActions {
  setLanguage: (language: string) => void
  setTheme: (theme: ThemeMode) => void
  setTimezone: (timezone: string) => void
  setTimezoneDisplayMode: (mode: TimezoneDisplayMode) => void
  resetSettings: () => void
  addReportSchedule: (schedule: ReportSchedule) => void
  updateReportSchedule: (id: string, updates: Partial<ReportSchedule>) => void
  deleteReportSchedule: (id: string) => void
  addConfigTemplate: (template: ConfigTemplate) => void
  updateConfigTemplate: (id: string, updates: Partial<ConfigTemplate>) => void
  deleteConfigTemplate: (id: string) => void
  addRoutingRule: (rule: AlertRoutingRule) => void
  updateRoutingRule: (id: string, updates: Partial<AlertRoutingRule>) => void
  deleteRoutingRule: (id: string) => void
}

type SettingsStore = SettingsState & SettingsActions

// ============== Constants ==============

// Common timezones (at least 20 as per requirements)
export const COMMON_TIMEZONES: TimezoneOption[] = [
  { value: 'UTC', label: 'UTC', offset: '+00:00' },
  { value: 'Asia/Shanghai', label: 'Beijing, Shanghai', offset: '+08:00' },
  { value: 'Asia/Hong_Kong', label: 'Hong Kong', offset: '+08:00' },
  { value: 'Asia/Tokyo', label: 'Tokyo', offset: '+09:00' },
  { value: 'Asia/Seoul', label: 'Seoul', offset: '+09:00' },
  { value: 'Asia/Singapore', label: 'Singapore', offset: '+08:00' },
  { value: 'Asia/Dubai', label: 'Dubai', offset: '+04:00' },
  { value: 'Europe/London', label: 'London', offset: '+00:00' },
  { value: 'Europe/Paris', label: 'Paris', offset: '+01:00' },
  { value: 'Europe/Berlin', label: 'Berlin', offset: '+01:00' },
  { value: 'Europe/Moscow', label: 'Moscow', offset: '+03:00' },
  { value: 'America/New_York', label: 'New York', offset: '-05:00' },
  { value: 'America/Chicago', label: 'Chicago', offset: '-06:00' },
  { value: 'America/Denver', label: 'Denver', offset: '-07:00' },
  { value: 'America/Los_Angeles', label: 'Los Angeles', offset: '-08:00' },
  { value: 'America/Toronto', label: 'Toronto', offset: '-05:00' },
  { value: 'America/Vancouver', label: 'Vancouver', offset: '-08:00' },
  { value: 'America/Sao_Paulo', label: 'Sao Paulo', offset: '-03:00' },
  { value: 'Australia/Sydney', label: 'Sydney', offset: '+11:00' },
  { value: 'Australia/Melbourne', label: 'Melbourne', offset: '+11:00' },
  { value: 'Pacific/Auckland', label: 'Auckland', offset: '+13:00' },
  { value: 'Pacific/Honolulu', label: 'Honolulu', offset: '-10:00' },
  { value: 'Asia/Kolkata', label: 'Mumbai, Kolkata', offset: '+05:30' },
  { value: 'Asia/Bangkok', label: 'Bangkok', offset: '+07:00' },
]

const DEFAULT_SETTINGS: SettingsState = {
  language: navigator.language.startsWith('zh') ? 'zh-CN' : 'en',
  theme: 'system',
  timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
  timezoneDisplayMode: 'local',
  reportSchedules: [],
  configTemplates: [],
  routingRules: [],
}

// ============== Store ==============

export const useSettingsStore = create<SettingsStore>()(
  persist(
    (set) => ({
      ...DEFAULT_SETTINGS,

      setLanguage: (language: string) => {
        set({ language })
        i18n.changeLanguage(language)
        localStorage.setItem('settings:language', language)
      },

      setTheme: (theme: ThemeMode) => {
        set({ theme })
        applyTheme(theme)
      },

      setTimezone: (timezone: string) => {
        set({ timezone })
      },

      setTimezoneDisplayMode: (timezoneDisplayMode: TimezoneDisplayMode) => {
        set({ timezoneDisplayMode })
      },

      resetSettings: () => {
        set(DEFAULT_SETTINGS)
        applyTheme(DEFAULT_SETTINGS.theme)
      },

      addReportSchedule: (schedule) => {
        set((s) => ({ reportSchedules: [...s.reportSchedules, schedule] }))
      },
      updateReportSchedule: (id, updates) => {
        set((s) => ({
          reportSchedules: s.reportSchedules.map((r) => (r.id === id ? { ...r, ...updates } : r)),
        }))
      },
      deleteReportSchedule: (id) => {
        set((s) => ({ reportSchedules: s.reportSchedules.filter((r) => r.id !== id) }))
      },

      addConfigTemplate: (template) => {
        set((s) => ({ configTemplates: [...s.configTemplates, template] }))
      },
      updateConfigTemplate: (id, updates) => {
        set((s) => ({
          configTemplates: s.configTemplates.map((t) => (t.id === id ? { ...t, ...updates } : t)),
        }))
      },
      deleteConfigTemplate: (id) => {
        set((s) => ({ configTemplates: s.configTemplates.filter((t) => t.id !== id) }))
      },

      addRoutingRule: (rule) => {
        set((s) => ({ routingRules: [...s.routingRules, rule] }))
      },
      updateRoutingRule: (id, updates) => {
        set((s) => ({
          routingRules: s.routingRules.map((r) => (r.id === id ? { ...r, ...updates } : r)),
        }))
      },
      deleteRoutingRule: (id) => {
        set((s) => ({ routingRules: s.routingRules.filter((r) => r.id !== id) }))
      },
    }),
    {
      name: 'settings-store',
      partialize: (state) => ({
        language: state.language,
        theme: state.theme,
        timezone: state.timezone,
        timezoneDisplayMode: state.timezoneDisplayMode,
        reportSchedules: state.reportSchedules,
        configTemplates: state.configTemplates,
        routingRules: state.routingRules,
      }),
    }
  )
)

// ============== Theme Helper ==============

/**
 * Apply theme to document
 */
export function applyTheme(theme: ThemeMode): void {
  const root = document.documentElement

  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    root.classList.toggle('dark', prefersDark)
  } else {
    root.classList.toggle('dark', theme === 'dark')
  }
}

/**
 * Get current effective theme (resolves 'system' to actual value)
 */
export function getEffectiveTheme(theme: ThemeMode): 'light' | 'dark' {
  if (theme === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return theme
}

/**
 * Initialize theme from store on app load
 */
export function initializeTheme(): void {
  const { theme } = useSettingsStore.getState()
  applyTheme(theme)

  // Listen for system theme changes
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  const handleChange = () => {
    const currentTheme = useSettingsStore.getState().theme
    if (currentTheme === 'system') {
      applyTheme('system')
    }
  }

  mediaQuery.addEventListener('change', handleChange)
}

/**
 * Initialize language from store on app load
 */
export function initializeLanguage(): void {
  const { language } = useSettingsStore.getState()
  localStorage.setItem('settings:language', language)
  i18n.changeLanguage(language)
}
