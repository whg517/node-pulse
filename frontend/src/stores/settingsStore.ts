/**
 * Settings Store
 *
 * Zustand store for user preferences including language, theme, and timezone.
 * Persists settings to localStorage for cross-session persistence.
 */

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// ============== Types ==============

export type ThemeMode = 'light' | 'dark' | 'system'
export type TimezoneDisplayMode = 'utc' | 'local' | 'nodeLocal' | 'multi'

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
}

export interface SettingsActions {
  setLanguage: (language: string) => void
  setTheme: (theme: ThemeMode) => void
  setTimezone: (timezone: string) => void
  setTimezoneDisplayMode: (mode: TimezoneDisplayMode) => void
  resetSettings: () => void
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
}

// ============== Store ==============

export const useSettingsStore = create<SettingsStore>()(
  persist(
    (set) => ({
      ...DEFAULT_SETTINGS,

      setLanguage: (language: string) => {
        set({ language })
        // Update i18n language - use dynamic import to avoid localStorage access during module load
        import('../i18n').then((i18n) => {
          i18n.default.changeLanguage(language)
        })
        // Also update localStorage for i18n init
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
    }),
    {
      name: 'settings-store',
      partialize: (state) => ({
        language: state.language,
        theme: state.theme,
        timezone: state.timezone,
        timezoneDisplayMode: state.timezoneDisplayMode,
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
export async function initializeLanguage(): Promise<void> {
  const { language } = useSettingsStore.getState()
  localStorage.setItem('settings:language', language)
  const i18n = await import('../i18n')
  i18n.default.changeLanguage(language)
}
