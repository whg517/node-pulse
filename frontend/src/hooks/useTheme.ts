/**
 * useTheme Hook
 *
 * Provides theme management functionality with localStorage persistence.
 * Supports light, dark, and system modes.
 */

import { useEffect, useCallback } from 'react'
import { useSettingsStore, applyTheme, getEffectiveTheme, type ThemeMode } from '../stores/settingsStore'

export interface UseThemeReturn {
  theme: ThemeMode
  effectiveTheme: 'light' | 'dark'
  setTheme: (theme: ThemeMode) => void
  toggleTheme: () => void
  isDark: boolean
  isLight: boolean
  isSystem: boolean
}

/**
 * Hook for managing application theme
 */
export function useTheme(): UseThemeReturn {
  const theme = useSettingsStore((state) => state.theme)
  const setTheme = useSettingsStore((state) => state.setTheme)

  // Get the effective theme (resolves 'system' to actual value)
  const effectiveTheme = getEffectiveTheme(theme)

  // Apply theme on mount and when it changes
  useEffect(() => {
    applyTheme(theme)
  }, [theme])

  // Listen for system theme changes
  useEffect(() => {
    if (theme !== 'system') return

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = () => {
      applyTheme('system')
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [theme])

  // Toggle between light and dark (skips system)
  const toggleTheme = useCallback(() => {
    const newTheme = effectiveTheme === 'dark' ? 'light' : 'dark'
    setTheme(newTheme)
  }, [effectiveTheme, setTheme])

  return {
    theme,
    effectiveTheme,
    setTheme,
    toggleTheme,
    isDark: effectiveTheme === 'dark',
    isLight: effectiveTheme === 'light',
    isSystem: theme === 'system',
  }
}
