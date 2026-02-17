/**
 * useTimezone Hook
 *
 * Provides timezone management functionality with localStorage persistence.
 * Supports UTC, local, node local, and multi-timezone display modes.
 */

import { useCallback, useMemo } from 'react'
import {
  useSettingsStore,
  COMMON_TIMEZONES,
  type TimezoneDisplayMode,
  type TimezoneOption,
} from '../stores/settingsStore'

export interface UseTimezoneReturn {
  timezone: string
  displayMode: TimezoneDisplayMode
  setTimezone: (timezone: string) => void
  setDisplayMode: (mode: TimezoneDisplayMode) => void
  timezones: TimezoneOption[]
  currentTimezone: TimezoneOption | undefined
  formatTime: (date: Date | string | number, options?: FormatOptions) => string
  formatTimeWithTimezone: (
    date: Date | string | number,
    timezone: string,
    options?: FormatOptions
  ) => string
  convertToTimezone: (date: Date | string | number, targetTimezone: string) => Date
  getOffsetString: (timezone: string) => string
}

interface FormatOptions {
  showDate?: boolean
  showTime?: boolean
  showSeconds?: boolean
  showTimezone?: boolean
  format?: 'short' | 'medium' | 'long'
}

const DEFAULT_FORMAT_OPTIONS: FormatOptions = {
  showDate: true,
  showTime: true,
  showSeconds: false,
  showTimezone: false,
  format: 'medium',
}

/**
 * Hook for managing timezone settings and formatting
 */
export function useTimezone(): UseTimezoneReturn {
  const timezone = useSettingsStore((state) => state.timezone)
  const displayMode = useSettingsStore((state) => state.timezoneDisplayMode)
  const setTimezone = useSettingsStore((state) => state.setTimezone)
  const setDisplayMode = useSettingsStore((state) => state.setTimezoneDisplayMode)

  // Find current timezone option
  const currentTimezone = useMemo(
    () => COMMON_TIMEZONES.find((tz) => tz.value === timezone),
    [timezone]
  )

  // Format time according to current settings
  const formatTime = useCallback(
    (date: Date | string | number, options?: FormatOptions): string => {
      const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options }
      const d = typeof date === 'string' || typeof date === 'number' ? new Date(date) : date

      if (isNaN(d.getTime())) {
        return 'Invalid Date'
      }

      const formatOpts: Intl.DateTimeFormatOptions = {}

      // Determine date format based on format option
      if (opts.showDate) {
        switch (opts.format) {
          case 'short':
            formatOpts.year = '2-digit'
            formatOpts.month = 'numeric'
            formatOpts.day = 'numeric'
            break
          case 'long':
            formatOpts.year = 'numeric'
            formatOpts.month = 'long'
            formatOpts.day = 'numeric'
            break
          default:
            formatOpts.year = 'numeric'
            formatOpts.month = '2-digit'
            formatOpts.day = '2-digit'
        }
      }

      // Determine time format
      if (opts.showTime) {
        formatOpts.hour = '2-digit'
        formatOpts.minute = '2-digit'
        if (opts.showSeconds) {
          formatOpts.second = '2-digit'
        }
      }

      // Add timezone name if requested
      if (opts.showTimezone) {
        formatOpts.timeZoneName = 'short'
      }

      // Set timezone based on display mode
      let tz: string | undefined
      switch (displayMode) {
        case 'utc':
          tz = 'UTC'
          break
        case 'local':
          tz = timezone
          break
        case 'nodeLocal':
        case 'multi':
          tz = timezone
          break
        default:
          tz = timezone
      }

      formatOpts.timeZone = tz

      return d.toLocaleString(undefined, formatOpts)
    },
    [timezone, displayMode]
  )

  // Format time with a specific timezone
  const formatTimeWithTimezone = useCallback(
    (date: Date | string | number, targetTimezone: string, options?: FormatOptions): string => {
      const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options }
      const d = typeof date === 'string' || typeof date === 'number' ? new Date(date) : date

      if (isNaN(d.getTime())) {
        return 'Invalid Date'
      }

      const formatOpts: Intl.DateTimeFormatOptions = {
        timeZone: targetTimezone,
      }

      if (opts.showDate) {
        formatOpts.year = 'numeric'
        formatOpts.month = '2-digit'
        formatOpts.day = '2-digit'
      }

      if (opts.showTime) {
        formatOpts.hour = '2-digit'
        formatOpts.minute = '2-digit'
        if (opts.showSeconds) {
          formatOpts.second = '2-digit'
        }
      }

      if (opts.showTimezone) {
        formatOpts.timeZoneName = 'short'
      }

      return d.toLocaleString(undefined, formatOpts)
    },
    []
  )

  // Convert date to a specific timezone (returns a new Date adjusted for display)
  const convertToTimezone = useCallback(
    (date: Date | string | number, targetTimezone: string): Date => {
      const d = typeof date === 'string' || typeof date === 'number' ? new Date(date) : date
      // Note: Date objects are always in UTC internally, this is for display purposes
      return new Date(
        d.toLocaleString('en-US', { timeZone: targetTimezone })
      )
    },
    []
  )

  // Get offset string for a timezone (e.g., "+08:00")
  const getOffsetString = useCallback((tz: string): string => {
    const now = new Date()
    const formatter = new Intl.DateTimeFormat('en-US', {
      timeZone: tz,
      timeZoneName: 'shortOffset',
    })
    const parts = formatter.formatToParts(now)
    const offsetPart = parts.find((p) => p.type === 'timeZoneName')
    return offsetPart?.value || '+00:00'
  }, [])

  return {
    timezone,
    displayMode,
    setTimezone,
    setDisplayMode,
    timezones: COMMON_TIMEZONES,
    currentTimezone,
    formatTime,
    formatTimeWithTimezone,
    convertToTimezone,
    getOffsetString,
  }
}
